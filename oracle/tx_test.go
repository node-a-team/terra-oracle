package oracle

import (
	"fmt"
	"testing"

	price "github.com/node-a-team/terra-oracle/price"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestLimit(t *testing.T) {
	ps := price.NewPriceService()
	os := NewOracleService(ps, nil)
	os.prevoteInited = true
	os.changeRateSoftLimit = 0.1
	os.changeRateHardLimit = 0.5

	ps.SetPrice("usd/krw", sdk.NewDecCoin("krw", sdk.NewInt(1)))
	ps.SetPrice("sdr/krw", sdk.NewDecCoin("krw", sdk.NewInt(1)))
	ps.SetPrice("luna/krw", sdk.NewDecCoin("krw", sdk.NewInt(100)))
	os.preLunaPrices["usd"] = sdk.NewDecCoin("krw", sdk.NewInt(100))
	os.preLunaPrices["sdr"] = sdk.NewDecCoin("krw", sdk.NewInt(100))
	os.preLunaPrices["krw"] = sdk.NewDecCoin("krw", sdk.NewInt(100))

	abort, err := os.calculatePrice()
	if err != nil {
		panic(err)
	}
	if abort {
		panic(nil)
	}

	// Within soft limit (greater)
	ps.SetPrice("luna/krw", sdk.NewDecCoin("krw", sdk.NewInt(105)))
	abort, err = os.calculatePrice()
	if err != nil {
		panic(err)
	}
	if abort {
		panic(nil)
	}
	if !os.lunaPrices["krw"].Amount.Equal(sdk.NewDec(105)) {
		panic(fmt.Errorf("luna price should be (%s), but (%s)", sdk.NewDec(105).String(), os.lunaPrices["krw"].Amount.String()))
	}

	// Within soft limit (lesser)
	ps.SetPrice("luna/krw", sdk.NewDecCoin("krw", sdk.NewInt(95)))
	abort, err = os.calculatePrice()
	if err != nil {
		panic(err)
	}
	if abort {
		panic(nil)
	}
	if !os.lunaPrices["krw"].Amount.Equal(sdk.NewDec(95)) {
		panic(fmt.Errorf("luna price should be (%s), but (%s)", sdk.NewDec(95).String(), os.lunaPrices["krw"].Amount.String()))
	}

	// Exceed soft limit (greater)
	ps.SetPrice("luna/krw", sdk.NewDecCoin("krw", sdk.NewInt(120)))
	abort, err = os.calculatePrice()
	if err != nil {
		panic(err)
	}
	if abort {
		panic(nil)
	}
	if !os.lunaPrices["krw"].Amount.Equal(sdk.NewDec(110)) {
		panic(fmt.Errorf("luna price should be (%s), but (%s)", sdk.NewDec(110).String(), os.lunaPrices["krw"].Amount.String()))
	}

	// Exceed soft limit (lesser)
	ps.SetPrice("luna/krw", sdk.NewDecCoin("krw", sdk.NewInt(80)))
	abort, err = os.calculatePrice()
	if err != nil {
		panic(err)
	}
	if abort {
		panic(nil)
	}
	if !os.lunaPrices["krw"].Amount.Equal(sdk.NewDec(90)) {
		panic(fmt.Errorf("luna price should be (%s), but (%s)", sdk.NewDec(90).String(), os.lunaPrices["krw"].Amount.String()))
	}

	// Exceed hard limit (greater)
	ps.SetPrice("luna/krw", sdk.NewDecCoin("krw", sdk.NewInt(40)))
	abort, err = os.calculatePrice()
	if err == nil {
		panic("should return err")
	}
	if !abort {
		panic("should be aborted")
	}

	// Exceed soft limit (lesser)
	ps.SetPrice("luna/krw", sdk.NewDecCoin("krw", sdk.NewInt(160)))
	abort, err = os.calculatePrice()
	if err == nil {
		panic("should return err")
	}
	if !abort {
		panic("should be aborted")
	}

	os.changeRateSoftLimit = 0
	os.changeRateHardLimit = 0

	// Exceed hard limit (greater)
	ps.SetPrice("luna/krw", sdk.NewDecCoin("krw", sdk.NewInt(40)))
	abort, err = os.calculatePrice()
	if err != nil {
		panic(nil)
	}
	if abort {
		panic(nil)
	}
	if !os.lunaPrices["krw"].Amount.Equal(sdk.NewDec(40)) {
		panic(fmt.Errorf("luna price should be (%s), but (%s)", sdk.NewDec(40).String(), os.lunaPrices["krw"].Amount.String()))
	}

	// Exceed soft limit (lesser)
	ps.SetPrice("luna/krw", sdk.NewDecCoin("krw", sdk.NewInt(160)))
	abort, err = os.calculatePrice()
	if err != nil {
		panic(nil)
	}
	if abort {
		panic(nil)
	}
	if !os.lunaPrices["krw"].Amount.Equal(sdk.NewDec(160)) {
		panic(fmt.Errorf("luna price should be (%s), but (%s)", sdk.NewDec(160).String(), os.lunaPrices["krw"].Amount.String()))
	}
}
