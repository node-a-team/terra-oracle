package price

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/tendermint/tendermint/libs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	cfg "github.com/node-a-team/terra-oracle/config"
)

type FxResponse struct {
	Result []Result `json:"price_results"`
}

type Result struct {
	Symbol      string  `json:"symbol"`
	Multiplier  float64 `json:"multiplier,string"`
	Px          float64 `json:"px,string"`
	RequestID   string  `json:"request_id"`
	ResolveTime int64   `json:"resolve_time,string"`
}

func (ps *PriceService) fxsToUsd(logger log.Logger) {
	for {
		if !cfg.Config.APIs.Band.Active {
			logger.Info("Warning APIs.Band.Active is false in Config.toml. Let's exit the fxsToUsd().")
			break
		}

		func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("Unknown error", r)
				}

				time.Sleep(cfg.Config.Options.Interval.Luna * time.Second)
			}()

			req, err := http.NewRequest(
				"GET",
				cfg.Config.APIs.Band.Band+"/oracle/v1/request_prices?symbols=LUNA&symbols=XDR&symbols=MNT&symbols=EUR&symbols=CNY&symbols=JPY&symbols=GBP&symbols=INR&symbols=CAD&symbols=CHF&symbols=HKD&symbols=SGD&symbols=AUD&min_count=3&ask_count=4",
				nil,
			)
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				logger.Error("Fail to fetch from band-fx", err.Error())
				return
			}
			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				logger.Error("Fail to read body", err.Error())
				return
			}

			var res FxResponse
			err = json.Unmarshal(body, &res)
			if err != nil {
				logger.Error("Fail to unmarshal body", err.Error())
				return
			}

			logs := ""
			for _, rate := range res.Result {
				symbol := strings.ToLower(rate.Symbol)
				if symbol == "xdr" {
					symbol = "sdr"
				}

				decAmount, err := sdk.NewDecFromStr(strconv.FormatFloat(rate.Px/rate.Multiplier, 'f', -1, 64))
				if err != nil {
					logger.Error("Fail to parse price to Dec", err.Error())
					return
				}

				ps.SetPrice(symbol+"/usd", sdk.NewDecCoinFromDec("usd", decAmount), rate.ResolveTime)
				logs += fmt.Sprintf(" [%s/usd:%v,timestamp:%d] ", symbol, decAmount, rate.ResolveTime)
			}

			logger.Info(fmt.Sprintf("Recent [%s]", logs))
		}()
	}
}
