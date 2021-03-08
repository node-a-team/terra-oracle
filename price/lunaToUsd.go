package price

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/tendermint/tendermint/libs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	cfg "github.com/node-a-team/terra-oracle/config"
)

// BinancePrice response from binance
type BinancePrice struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

func (ps *PriceService) lunaToUsd(logger log.Logger) {
	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("Unknown error", r)
				}

				time.Sleep(cfg.Config.Options.Interval.Luna * time.Second)
			}()

			// resp, err := http.Get("https://api.binance.com/api/v3/ticker/price?symbol=LUNAUSDT")
			resp, err := http.Get(cfg.Config.APIs.Luna.Usd.Binance)
			if err != nil {
				logger.Error("Fail to fetch from binance", err.Error())
				return
			}
			defer func() {
				resp.Body.Close()
			}()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				logger.Error("Fail to read body", err.Error())
				return
			}

			bp := BinancePrice{}
			err = json.Unmarshal(body, &bp)
			if err != nil {
				logger.Error("Fail to unmarshal json", err.Error())
				return
			}

			price := bp.Price

			timestamp := time.Now().UTC().Unix()

			logger.Info(fmt.Sprintf("Recent luna/usd: %s, timestamp: %d", price, timestamp))

			// amount, ok := sdk.NewIntFromString(recent.Price)
			decAmount, err := sdk.NewDecFromStr(price)
			// if !ok {
			if err != nil {
				logger.Error("Fail to parse price to Dec")
			}

			ps.SetPrice("luna/usd", sdk.NewDecCoinFromDec("usd", decAmount), int64(timestamp))
		}()
	}
}
