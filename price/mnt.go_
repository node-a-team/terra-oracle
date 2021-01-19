package price

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
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

func (ps *PriceService) mntToKrw(logger log.Logger) {

	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("Unknown error", r)
				}

				time.Sleep(cfg.Config.Options.Interval * time.Second)
			}()

			//			resp, err := http.Get("http://www.apilayer.net/api/live?access_key=")
			resp, err := http.Get(cfg.Config.APIs.MNT.Currencylayer)
			if err != nil {
				logger.Error("Fail to fetch from freeforexapi", err.Error())
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

			mntToKrw := res.Quotes["USDKRW"] / res.Quotes["USDMNT"]

			price := strconv.FormatFloat(mntToKrw, 'f', -1, 64)

			logger.Info(fmt.Sprintf("Recent mnt/krw: %s, timestamp: %d", price, res.Timestamp))

			decAmount, err := sdk.NewDecFromStr(price)
			if err != nil {
				logger.Error("Fail to parse price to Dec", err.Error())
				return
			}
			ps.SetPrice("mnt/krw", sdk.NewDecCoinFromDec("krw", decAmount), res.Timestamp)
		}()
	}
}
