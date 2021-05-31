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

// APILayerResponse response body from currencylayer
type APILayerResponse struct {
	Success   bool               `json:"success"`
	Terms     string             `json:"terms"`
	Privacy   string             `json:"privacy"`
	Timestamp int64              `json:"timestamp"`
	Source    string             `json:"source"`
	Quotes    map[string]float64 `json:"quotes"`
}

var (
	stables = []string{"XDR", "MNT", "EUR", "CNY", "JPY", "GBP", "INR", "CAD", "CHF", "HKD", "SGD", "AUD", "THB", "SEK", "DKK", "NOK"}
)

func (ps *PriceService) stablesToUsd(logger log.Logger) {
	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("Unknown error", r)
				}

				time.Sleep(cfg.Config.Options.Interval.Stables * time.Second)
			}()

			// resp, err := http.Get("https://api.currencylayer.com/live?access_key=")
			resp, err := http.Get(cfg.Config.APIs.Stables.Currencylayer)
			if err != nil {
				logger.Error("Fail to fetch from currencylayer", err.Error())
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

			res := APILayerResponse{}
			err = json.Unmarshal(body, &res)
			if err != nil {
				logger.Error("Fail to unmarshal body", err.Error())
				return
			}

			for _, s := range stables {
				setStablesPrice(s, res, ps, logger)
			}

			fmt.Println("")
		}()
	}
}

func setStablesPrice(stable string, res APILayerResponse, ps *PriceService, logger log.Logger) {
	stableToUsd := res.Quotes["USD"+stable]
	if stable == "XDR" {
		stable = "SDR"
	}

	price := strconv.FormatFloat(stableToUsd, 'f', -1, 64)
	decAmount, err := sdk.NewDecFromStr(price)
	if err != nil {
		logger.Error("Fail to parse price to Dec", err.Error())
		return
	}

	decAmount = sdk.OneDec().Quo(decAmount)
	logger.Info(fmt.Sprintf("Recent %s/usd: %s, timestamp: %d", strings.ToLower(stable), decAmount, res.Timestamp))

	// Set USD price of stables
	ps.SetPrice(strings.ToLower(stable)+"/usd", sdk.NewDecCoinFromDec("usd", decAmount), res.Timestamp)
}
