package addresses

import "github.com/ethereum/go-ethereum/common"

var (
	usdcToWeth   = common.HexToAddress("0x853ee4b2a13f8a742d64c8f088be7ba2131f670d")
	usdcToUsdt   = common.HexToAddress("0x2cf7252e74036d1da831d11089d326296e64a728")
	wethToUsdt   = common.HexToAddress("0xf6422b997c7f54d1c6a6e103bcb1499eea0a7046")
	wmaticToWeth = common.HexToAddress("0xadbf1854e5883eb8aa7baf50705338739e558e5b")
	wmaticToUsdc = common.HexToAddress("0x6e7a5fafcec6bb1e78bae2a1f0b612012bf14827")
	wmaticToUsdt = common.HexToAddress("0x604229c960e5cacf2aaeac8be68ac07ba9df81c3")
)

var QUICKSWAP_ROUTER = common.HexToAddress("0xa5E0829CaCEd8fFDD4De3c43696c57F7D7A678ff")

type tokenPair struct {
	token0 string
	token1 string
}

var ADDRESS_TO_TOKENPAIR = map[common.Address]tokenPair{
	wethToUsdt:   tokenPair{"WETH", "USDT"},
	usdcToWeth:   tokenPair{"USDC", "WETH"},
	usdcToUsdt:   tokenPair{"USDC", "USDT"},
	wmaticToWeth: tokenPair{"WMATIC", "WETH"},
	wmaticToUsdc: tokenPair{"WMATIC", "USDC"},
	wmaticToUsdt: tokenPair{"WMATIC", "USDT"},
}
var NODES = map[string]int{
	"WETH": 0, "USDT": 1, "USDC": 2, "WMATIC": 3,
}

var CURRENCIES = []string{"WETH", "USDT", "USDC", "WMATIC"}
