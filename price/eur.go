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

func (ps *PriceService) eurToKrw(logger log.Logger) {
	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("Unknown error", r)
				}

				time.Sleep(30 * time.Second)
			}()

			// resp, err := http.Get("https://quotation-api-cdn.dunamu.com/v1/forex/recent?codes=FRX.KRWEUR")
			resp, err := http.Get(cfg.Config.APIs.EUR.Dunamu)
			if err != nil {
				logger.Error("Fail to fetch from dunamu", err.Error())
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

			var res []map[string]interface{}
			err = json.Unmarshal(body, &res)
			if err != nil {
				logger.Error("Fail to unmarshal body", err.Error())
				return
			}

			if len(res) == 0 {
				logger.Error("Fail got empty response")
				return
			}

			price := strconv.FormatFloat(res[0]["basePrice"].(float64), 'f', -1, 64)
			timestamp := int64(res[0]["timestamp"].(float64)) / 1000

			logger.Info(fmt.Sprintf("Recent eur/krw: %s, timestamp: %d", price, timestamp))

			decAmount, err := sdk.NewDecFromStr(price)
			if err != nil {
				logger.Error("Fail to parse price to Dec", err.Error())
				return
			}

			ps.SetPrice("eur/krw", sdk.NewDecCoinFromDec("krw", decAmount), timestamp)
		}()
	}
}
