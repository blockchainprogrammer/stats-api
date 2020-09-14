package models

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gochain/gochain/v3/accounts/abi/bind"
	"github.com/gochain/gochain/v3/common"
	"github.com/goswap/stats-api/contracts"
	"github.com/goswap/stats-api/utils"
	"github.com/shopspring/decimal"
	"github.com/treeder/gotils"
)

// Pair represents a GoSwap pair
type Pair struct {
	Index   int            `firestore:"index" json:"index"`
	Address common.Address `firestore:"-" json:"address"`

	PairContract *contracts.Pair `firestore:"-" json:"-"` // this is the object to interact with the contract
	Token0       *Token          `firestore:"-" json:"token0"`
	Token1       *Token          `firestore:"-" json:"token1"`

	// for database
	AddressHex    string `firestore:"address" json:"-"`
	Token0Address string `firestore:"token0address" json:"-"`
	Token1Address string `firestore:"token1address" json:"-"`
}

func (pb *Pair) PreSave() {
	pb.AddressHex = pb.Address.Hex()
	pb.Token0Address = pb.Token0.Address.Hex()
	pb.Token1Address = pb.Token1.Address.Hex()

}
func (pb *Pair) AfterLoad(ctx context.Context) {
	pb.Address = common.HexToAddress(pb.AddressHex)
	// can't really load tokens here...
}
func (td *Pair) String() string {
	return fmt.Sprintf("%v-%v", td.Token0.Symbol, td.Token1.Symbol)
}

func (td *Pair) GetReserves(ctx context.Context) (decimal.Decimal, decimal.Decimal, error) {
	var opts *bind.CallOpts
	// if blockNumber > 0 {
	// 	opts = &bind.CallOpts{BlockNumber: big.NewInt(blockNumber)}
	// }
	reserves, err := td.PairContract.GetReserves(opts)
	if err != nil {
		return decimal.Zero, decimal.Zero, gotils.C(ctx).Errorf("error getting reserves: %v", err)
	}
	return utils.IntToDec(reserves.Reserve0, td.Token0.Decimals), utils.IntToDec(reserves.Reserve1, td.Token1.Decimals), nil
}

// PriceInUSD only calls this on a USDC pair
func (td *Pair) PriceInUSD(ctx context.Context) (decimal.Decimal, error) {
	// calc is getReserves()
	// USDC reserve, shifted token.decimals over (6)
	// divided by token reserve shifted token.decimals over
	// that will give us the correct amount
	var opts *bind.CallOpts
	// if blockNumber > 0 {
	// 	opts = &bind.CallOpts{BlockNumber: big.NewInt(blockNumber)}
	// }
	reserves, err := td.PairContract.GetReserves(opts)
	if err != nil {
		return decimal.Zero, gotils.C(ctx).Errorf("error getting reserves: %v", err)
	}
	// var usdc, other *Token
	var usdcReserve, otherReserve decimal.Decimal
	if td.Token0.Symbol == "USDC" {
		// usdc = td.Token0
		// other = td.Token1
		usdcReserve = utils.IntToDec(reserves.Reserve0, td.Token0.Decimals)
		otherReserve = utils.IntToDec(reserves.Reserve1, td.Token1.Decimals)
	} else if td.Token1.Symbol == "USDC" {
		// usdc = td.Token1
		// other = td.Token0
		usdcReserve = utils.IntToDec(reserves.Reserve1, td.Token1.Decimals)
		otherReserve = utils.IntToDec(reserves.Reserve0, td.Token0.Decimals)
	} else {
		return decimal.Zero, errors.New("NO USDC IN PAIR")
	}
	return usdcReserve.Div(otherReserve), nil
}

// Token represents an ERC20
type Token struct {
	Address common.Address `firestore:"-" json:"address"`

	Name        string          `firestore:"name" json:"name"`
	Symbol      string          `firestore:"symbol" json:"symbol"`
	Decimals    uint8           `firestore:"decimals" json:"decimals"`
	TotalSupply decimal.Decimal `firestore:"-" json:"totalSupply"`
	CMCPrice    decimal.Decimal `firestore:"-" json:"CMCPrice"`

	TotalSupplyS string `firestore:"totalSupply" json:"-"`
	CMCPriceS    string `firestore:"CMCPrice" json:"-"`

	// database
	AddressHex string `firestore:"address" json:"-"`
}

func (pb *Token) PreSave() {
	pb.AddressHex = pb.Address.Hex()
	pb.TotalSupplyS = pb.TotalSupply.String()
	pb.CMCPriceS = pb.CMCPrice.String()
}
func (pb *Token) AfterLoad(ctx context.Context) {
	pb.Address = common.HexToAddress(pb.AddressHex)
	pb.TotalSupply, _ = decimal.NewFromString(pb.TotalSupplyS)
	pb.CMCPrice, _ = decimal.NewFromString(pb.CMCPriceS)
}
func (td *Token) String() string {
	return fmt.Sprintf("%v", td.Symbol)
}

type PairLiquidity struct {
	// Address is the ID of the pair
	Address string `firestore:"address" json:"address"`

	Time   time.Time `firestore:"time" json:"time"`
	Pair   string    `firestore:"pair" json:"pair"`
	Token0 string    `firestore:"token0" json:"token0"`
	Token1 string    `firestore:"token1" json:"token1"`

	TotalSupply decimal.Decimal `firestore:"-" json:"totalSupply"`
	Reserve0    decimal.Decimal `firestore:"-" json:"reserve0"`
	Reserve1    decimal.Decimal `firestore:"-" json:"reserve1"`
	Price0      decimal.Decimal `firestore:"-" json:"price0"`
	Price1      decimal.Decimal `firestore:"-" json:"price1"`

	// firebase stuff
	TotalSupplyS string `firestore:"totalSupply" json:"-"`
	Reserve0S    string `firestore:"reserve0" json:"-"`
	Reserve1S    string `firestore:"reserve1" json:"-"`
	Price0S      string `firestore:"price0" json:"-"`
	Price1S      string `firestore:"price1" json:"-"`
}

func (s *PairLiquidity) ValUSD() decimal.Decimal {
	reserve0val := s.Reserve0.Mul(s.Price0)
	reserve1val := s.Reserve1.Mul(s.Price1)
	totalPoolVal := reserve0val.Add(reserve1val)
	return totalPoolVal
}
func (pb *PairLiquidity) PreSave() {
	pb.TotalSupplyS = pb.TotalSupply.String()
	pb.Reserve0S = pb.Reserve0.String()
	pb.Reserve1S = pb.Reserve1.String()
	pb.Price0S = pb.Price0.String()
	pb.Price1S = pb.Price1.String()

}
func (pb *PairLiquidity) AfterLoad(ctx context.Context) {
	// t.Ref = ref
	// t.ID = t.Ref.ID
	pb.TotalSupply, _ = decimal.NewFromString(pb.TotalSupplyS)
	pb.Reserve0, _ = decimal.NewFromString(pb.Reserve0S)
	pb.Reserve1, _ = decimal.NewFromString(pb.Reserve1S)
	pb.Price0, _ = decimal.NewFromString(pb.Price0S)
	pb.Price1, _ = decimal.NewFromString(pb.Price1S)
}

type TokenLiquidity struct {
	// Address is the ID of the token
	Address string `firestore:"address" json:"address"`

	Time   time.Time `firestore:"time" json:"time"`
	Symbol string    `firestore:"symbol"`

	Reserve decimal.Decimal `firestore:"-" json:"reserve"`
	Price   decimal.Decimal `firestore:"-" json:"price"`

	// firebase stuff
	ReserveS string `firestore:"reserve" json:"-"`
	PriceS   string `firestore:"price" json:"-"`
}

func (s *TokenLiquidity) TokenLiquidity() decimal.Decimal {
	reserve0val := s.Reserve.Mul(s.Price)
	return reserve0val
}
func (pb *TokenLiquidity) PreSave() {
	pb.ReserveS = pb.Reserve.String()
	pb.PriceS = pb.Price.String()
}

func (pb *TokenLiquidity) AfterLoad(ctx context.Context) {
	// t.Ref = ref
	// t.ID = t.Ref.ID
	pb.Reserve, _ = decimal.NewFromString(pb.ReserveS)
	pb.Price, _ = decimal.NewFromString(pb.PriceS)
}

type PairBucket struct {
	// Address is the ID of the pair
	Address string `firestore:"address" json:"address"`

	Time time.Time `firestore:"time" json:"time"`
	Pair string    `firestore:"pair" json:"pair"`

	Amount0In  decimal.Decimal `firestore:"-" json:"amount0In"`
	Amount1In  decimal.Decimal `firestore:"-" json:"amount1In"`
	Amount0Out decimal.Decimal `firestore:"-" json:"amount0Out"`
	Amount1Out decimal.Decimal `firestore:"-" json:"amount1Out"`
	Price0USD  decimal.Decimal `firestore:"-" json:"price0USD"`
	Price1USD  decimal.Decimal `firestore:"-" json:"price1USD"`
	VolumeUSD  decimal.Decimal `firestore:"-" json:"volumeUSD"` // in USD

	// For firebase
	Amount0InS  string `firestore:"amount0In" json:"-"`
	Amount1InS  string `firestore:"amount1In" json:"-"`
	Amount0OutS string `firestore:"amount0Out" json:"-"`
	Amount1OutS string `firestore:"amount1Out" json:"-"`
	Price0USDS  string `firestore:"price0USD" json:"-"`
	Price1USDS  string `firestore:"price1USD" json:"-"`
	VolumeUSDS  string `firestore:"volumeUSD" json:"-"`
}

// PreSave Need these annoying things because firebase doesn't handle things properly
func (pb *PairBucket) PreSave() {
	pb.Amount0InS = pb.Amount0In.String()
	pb.Amount1InS = pb.Amount1In.String()
	pb.Amount0OutS = pb.Amount0Out.String()
	pb.Amount1OutS = pb.Amount1Out.String()
	pb.Price0USDS = pb.Price0USD.String()
	pb.Price1USDS = pb.Price1USD.String()
	pb.VolumeUSDS = pb.VolumeUSD.String()

}
func (pb *PairBucket) AfterLoad(ctx context.Context) {
	// t.Ref = ref
	// t.ID = t.Ref.ID
	pb.Amount0In, _ = decimal.NewFromString(pb.Amount0InS)
	pb.Amount1In, _ = decimal.NewFromString(pb.Amount1InS)
	pb.Amount0Out, _ = decimal.NewFromString(pb.Amount0OutS)
	pb.Amount1Out, _ = decimal.NewFromString(pb.Amount1OutS)
	pb.Price0USD, _ = decimal.NewFromString(pb.Price0USDS)
	pb.Price1USD, _ = decimal.NewFromString(pb.Price1USDS)
	pb.VolumeUSD, _ = decimal.NewFromString(pb.VolumeUSDS)
}

type TokenBucket struct {
	// Address is the ID of the token
	Address string `firestore:"address" json:"address"`

	Time   time.Time `firestore:"time"`
	Symbol string    `firestore:"symbol"`

	AmountIn  decimal.Decimal `firestore:"-" json:"amountIn"`
	AmountOut decimal.Decimal `firestore:"-" json:"amountOut"`
	PriceUSD  decimal.Decimal `firestore:"-" json:"priceUSD"`
	VolumeUSD decimal.Decimal `firestore:"-" json:"volumeUSD"` // in USD

	// firebase bullshit:
	AmountInS  string `firestore:"amountIn" json:"-"`
	AmountOutS string `firestore:"amountOut" json:"-"`
	PriceUSDS  string `firestore:"priceUSD" json:"-"`
	VolumeUSDS string `firestore:"volumeUSD" json:"-"`
}

// PreSave Need these annoying things because firebase doesn't handle things properly
func (pb *TokenBucket) PreSave() {
	pb.AmountInS = pb.AmountIn.String()
	pb.AmountOutS = pb.AmountOut.String()
	pb.PriceUSDS = pb.PriceUSD.String()
	pb.VolumeUSDS = pb.VolumeUSD.String()

}
func (pb *TokenBucket) AfterLoad(ctx context.Context) {
	// t.Ref = ref
	// t.ID = t.Ref.ID
	pb.AmountIn, _ = decimal.NewFromString(pb.AmountInS)
	pb.AmountOut, _ = decimal.NewFromString(pb.AmountOutS)
	pb.PriceUSD, _ = decimal.NewFromString(pb.PriceUSDS)
	pb.VolumeUSD, _ = decimal.NewFromString(pb.VolumeUSDS)
}

type TotalBucket struct {
	Time time.Time `firestore:"time"`

	VolumeUSD    decimal.Decimal `firestore:"-" json:"volumeUSD"`    // in USD
	LiquidityUSD decimal.Decimal `firestore:"-" json:"liquidityUSD"` // in USD

	// fireabase :(
	VolumeUSDS    string `firestore:"volumeUSD" json:"-"`
	LiquidityUSDS string `firestore:"liquidityUSD" json:"-"`
}

// PreSave Need these annoying things because firebase doesn't handle things properly
func (pb *TotalBucket) PreSave() {
	pb.VolumeUSDS = pb.VolumeUSD.String()
	pb.LiquidityUSDS = pb.LiquidityUSD.String()

}
func (pb *TotalBucket) AfterLoad(ctx context.Context) {
	// t.Ref = ref
	// t.ID = t.Ref.ID
	pb.VolumeUSD, _ = decimal.NewFromString(pb.VolumeUSDS)
	pb.LiquidityUSD, _ = decimal.NewFromString(pb.LiquidityUSDS)
}