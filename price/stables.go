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

type APILayerResponse struct {
	Success   bool               `json:"success"`
	Terms     string             `json:"terms"`
	Privacy   string             `json:"privacy"`
	Timestamp int64              `json:"timestamp"`
	Source    string             `json:"source"`
	Quotes    map[string]float64 `json:"quotes"`
}

var (
	stables = []string{"USD", "MNT", "EUR", "CNY", "JPY", "GBP", "INR", "CAD", "CHF", "HKD", "SGD", "AUD"}
)

func (ps *PriceService) stablesToKrw(logger log.Logger) {

	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("Unknown error", r)
				}

				time.Sleep(cfg.Config.Options.Interval * time.Second)
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

			var res APILayerResponse
			err = json.Unmarshal(body, &res)
			if err != nil {
				logger.Error("Fail to unmarshal body", err.Error())
				return
			}

			// Set
			for _, s := range stables {
//				if s == "USD" || s == "MNT" || s == "EUR" {
					setStablesPrice(s, res, ps, logger)
//				} else {
//					logger.Info(fmt.Sprintf("Ready %s/krw: %f", s, res.Quotes["USDKRW"] / res.Quotes["USD"+s]))
//				}
			}
			fmt.Println("")

		}()
	}
}

func setStablesPrice(stable string, res APILayerResponse, ps *PriceService, logger log.Logger) {

	// stable == "USD"
	stableToKrw := res.Quotes["USDKRW"]

	if stable != "USD" {
		stableToKrw = res.Quotes["USDKRW"] / res.Quotes["USD"+stable]
	}


//	} else {
//		fmt.Println("Not Contains in stables: " +stables)
//	}

        price := strconv.FormatFloat(stableToKrw, 'f', -1, 64)
        logger.Info(fmt.Sprintf("Recent %s/krw: %s, timestamp: %d", strings.ToLower(stable), price, res.Timestamp))
        decAmount, err := sdk.NewDecFromStr(price)
        if err != nil {
                logger.Error("Fail to parse price to Dec", err.Error())
	        return
        }
        ps.SetPrice(strings.ToLower(stable) +"/krw", sdk.NewDecCoinFromDec("krw", decAmount), res.Timestamp)
}

