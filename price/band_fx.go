package price

import (
	"bytes"
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
	Height string   `json:"height"`
	Result []Result `json:"result"`
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

				time.Sleep(cfg.Config.Options.Interval * time.Second)
			}()

			req, err := http.NewRequest(
				"POST",
				cfg.Config.APIs.Band.Band+"/oracle/request_prices",
				bytes.NewBuffer([]byte(`{"symbols":["LUNA", "XDR", "MNT", "EUR", "CNY", "JPY", "GBP", "INR", "CAD", "CHF", "HKD", "SGD", "AUD"],"min_count":3,"ask_count":4}`)),
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
