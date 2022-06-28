package main

import (
	"fmt"
	"log"
	"math"
	"math/big"
	"poly/abis/erc20"
	"poly/abis/pair"
	"poly/addresses"
	testarbtables "poly/testArbTables"
	"time"

	"context"

	"github.com/ALTree/bigfloat"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	inf        = new(big.Float).SetInf(true)
	neg_one    = new(big.Float).SetFloat64(-1)
	plain_one  = new(big.Float).SetFloat64(1.1)
	one        = new(big.Float).SetFloat64(1)
	nine_nine  = new(big.Float).SetFloat64(0.99)
	zero       = new(big.Float).SetFloat64(0)
	ten        = new(big.Int).SetInt64(10)
	fee        = new(big.Float).SetFloat64(0.999)
	nodes      map[string]int
	table      map[string][]big.Float
	currencies []string
)

type currency_edge struct {
	FromToken, ToToken string
	Price              big.Float
	PairAddress        common.Address
}

type newPair struct {
	pairet *pair.Pair
	token0 string
	token1 string
	rate   float64
}
type price_quote struct {
	TokenIn       string
	TokenOut      string
	PriceInToOut  *big.Float
	PriceNegOfLog *big.Float
}

type test struct {
	par            pair.Pair
	tokenName1     string
	tokenName2     string
	tokenContract1 erc20.Erc20
	tokenContract2 erc20.Erc20
	tokenSymbol1   string
	tokenSymbol2   string
	bigInt1        big.Int
	bigInt2        big.Int
}

func main() {
	//nodes = testarbtables.NODES
	//currencies = testarbtables.CURRENCIES
	currencies = addresses.CURRENCIES
	nodes = addresses.NODES
	tableInit()
	startBot()
	//testBellmanFord()
}

func tableInit() {
	table = make(map[string][]big.Float)
	var startArr []big.Float
	for _, curr := range currencies {
		_ = curr
		startArr = append(startArr, *zero)
	}
	for _, curr := range currencies {
		tmp := make([]big.Float, len(currencies))
		copy(tmp, startArr)
		fmt.Println(curr)
		table[curr] = tmp
	}
}

func startBot() {

	client, err := ethclient.Dial("")
	if err != nil {
		log.Fatal(err)
	}
	headers := make(chan *types.Header)
	sub, err := client.SubscribeNewHead(context.Background(), headers)
	if err != nil {
		log.Fatal(err)
	}
	pairs, err := new_all_relevant_pairs(*client)

	for {
		select {
		case err := <-sub.Err():
			log.Fatal(err)
		case header := <-headers:
			fmt.Println(header.Hash().Hex())
			start := time.Now()
			if err != nil {
				log.Fatal(err)
			}
			create_edges(pairs, *client)
			duration := time.Since(start)
			fmt.Println(duration)

		}
	}
}

func new_all_relevant_pairs(client ethclient.Client) ([]*test, error) {
	var testArr []*test
	var use_pairs_addrs []common.Address

	for key, _ := range addresses.ADDRESS_TO_TOKENPAIR {
		use_pairs_addrs = append(use_pairs_addrs, key)
	}

	for _, pair_addr := range use_pairs_addrs {
		var tmpTest test
		pair, err := pair.NewPair(pair_addr, &client)

		if err != nil {
			return nil, err
		}
		tmpTest.par = *pair
		token0, err := pair.Token0(nil)
		if err != nil {
			return nil, err
		}

		token1, err := pair.Token1(nil)
		if err != nil {
			return nil, err
		}
		tmpTest.tokenName1 = token0.String()
		tmpTest.tokenName2 = token1.String()

		contract0, err := erc20.NewErc20(token0, &client)
		if err != nil {
			return nil, err
		}

		contract1, err := erc20.NewErc20(token1, &client)
		if err != nil {
			return nil, err
		}

		tmpTest.tokenContract1 = *contract0
		tmpTest.tokenContract2 = *contract1

		symbol0, err := contract0.Symbol(nil)
		if err != nil {
			return nil, err
		}

		symbol1, err := contract1.Symbol(nil)
		if err != nil {
			return nil, err
		}

		tmpTest.tokenSymbol1 = symbol0
		tmpTest.tokenSymbol2 = symbol1

		dec0, err := contract0.Decimals(nil)

		if err != nil {
			return nil, err
		}

		dec1, err := contract1.Decimals(nil)

		if err != nil {
			return nil, err
		}

		one_token0 := new(big.Int).Exp(ten, big.NewInt(int64(dec0)), nil)
		one_token1 := new(big.Int).Exp(ten, big.NewInt(int64(dec1)), nil)

		tmpTest.bigInt1 = *one_token0
		tmpTest.bigInt2 = *one_token1

		testArr = append(testArr, &tmpTest)
	}
	return testArr, nil
}

func create_edges(test []*test, client ethclient.Client) error {
	var quotes []price_quote

	for _, tmp := range test {
		resrs, err := tmp.par.GetReserves(nil)
		if err != nil {
			return err
		}
		tokenRsrs1 := new(big.Int).Div(resrs.Reserve0, &tmp.bigInt1)
		tokenRsrs2 := new(big.Int).Div(resrs.Reserve1, &tmp.bigInt2)
		x, y := new(big.Float).SetInt(tokenRsrs1), new(big.Float).SetInt(tokenRsrs2)
		z1 := new(big.Float).Quo(x, y)
		z2 := new(big.Float).Quo(y, x)

		z1.Mul(z1, fee)
		z2.Mul(z2, fee)

		p0_neg_log := bigfloat.Log(z1)
		p0_neg_log.Mul(p0_neg_log, neg_one)

		p1_neg_log := bigfloat.Log(z2)
		p1_neg_log.Mul(p1_neg_log, neg_one)

		quotes = append(quotes, price_quote{
			TokenIn: tmp.tokenSymbol1, TokenOut: tmp.tokenSymbol2,
			PriceInToOut: z1, PriceNegOfLog: p0_neg_log,
		}, price_quote{
			TokenIn: tmp.tokenSymbol2, TokenOut: tmp.tokenSymbol1,
			PriceInToOut: z2, PriceNegOfLog: p1_neg_log,
		})

	}
	createTable(quotes)
	return bellmanFord(quotes)
}

func bellmanFord(quotes []price_quote) error {
	distances := make([]float64, len(nodes))

	predecessors := make([]string, len(nodes))

	for i := range distances {
		distances[i] = math.Inf(1)
		predecessors[i] = "nil"
	}

	distances[0] = 0

	for i := 0; i < len(nodes)-1; i++ {
		for _, edge := range quotes {
			cost, _ := edge.PriceNegOfLog.Float64()
			token_in, exists := nodes[edge.TokenIn]

			if !exists {
				fmt.Println("BAD -> assign it", edge.TokenIn)
				continue
			}

			token_out, exists := nodes[edge.TokenOut]

			if !exists {
				fmt.Println("BAD -> assign it", edge.TokenOut)
				continue
			}

			a := distances[token_in]
			b := distances[token_out]

			if a+cost < b {
				predecessors[nodes[edge.TokenOut]] = edge.TokenIn
				distances[nodes[edge.TokenOut]] = a + cost
			}

		}

	}

	for _, edge := range quotes {
		cost, _ := edge.PriceNegOfLog.Float64()
		a := distances[nodes[edge.TokenIn]]
		b := distances[nodes[edge.TokenOut]]

		if a+cost < b {
			fmt.Println("Found arb")
			findArbLoop(predecessors)
			return nil
		}
	}
	return nil

}

func findArbLoop(predecessors []string) {
	current := ""
	path := []string{}
	fmt.Println("the path: ", predecessors)
	for current != currencies[0] {
		current = predecessors[nodes[current]]
		path = append(path, current)
	}
	sum := zero
	fmt.Println(path)
	for i := len(path) - 1; i >= 1; i-- {
		tmp := getRate(path[i], path[i-1])
		sum.Add(sum, tmp)
	}
	fmt.Println(sum.String())
	tmp := getRate(path[0], path[len(path)-1])
	sum.Add(sum, tmp)
	sum.Mul(sum, neg_one)
	fmt.Println(sum.String())

	profit := bigfloat.Exp(sum)
	fmt.Println("profit: ", profit.String())

}

func getRate(from string, to string) *big.Float {
	fmt.Println(from, "----->", to)
	test := table[from][nodes[to]]
	fmt.Println("rate: ", test.String())
	return &test
}

func testBellmanFord() {

	var quotes []price_quote

	for i, currency := range testarbtables.CURRENCIES {
		for j, rate := range testarbtables.ARBTABLE[currency] {
			t := new(big.Float).SetFloat64(rate)
			p0_neg_log := bigfloat.Log(t)
			p0_neg_log.Mul(p0_neg_log, neg_one)

			quotes = append(quotes, price_quote{
				TokenIn: testarbtables.CURRENCIES[i], TokenOut: testarbtables.CURRENCIES[j],
				PriceInToOut: zero, PriceNegOfLog: p0_neg_log,
			})
		}

	}
	createTable(quotes)
	bellmanFord(quotes)
}

//create a table from the quotes list
func createTable(quotes []price_quote) {
	for _, quote := range quotes {
		table[quote.TokenIn][nodes[quote.TokenOut]] = *quote.PriceNegOfLog
	}
	//printTable()

}

//prints a table with all the rates
func printTable() {
	fmt.Println("TABLE___________________________________________________")
	for _, currency := range currencies {
		fmt.Print(currency, ":    ")
		for _, rate := range table[currency] {
			fmt.Print(rate.String(), "   ")
		}
		fmt.Println()

	}
}
