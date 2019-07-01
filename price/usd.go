package price

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"

	"github.com/tendermint/tendermint/libs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (ps *PriceService) usdToKrw(logger log.Logger) {
	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("Unknown error", r)
				}

				time.Sleep(30 * time.Second)
			}()

			resp, err := http.Get("https://quotation-api-cdn.dunamu.com/v1/forex/recent?codes=FRX.KRWUSD")
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

			re, _ := regexp.Compile("\"basePrice\":[0-9.]+")
			str := re.FindString(string(body))
			re, _ = regexp.Compile("[0-9.]+")
			price := re.FindString(str)

			logger.Info(fmt.Sprintf("Recent usd/krw: %s", price))

			decAmount, err := sdk.NewDecFromStr(price)
			if err != nil {
				logger.Error("Fail to parse price to Dec", err.Error())
				return
			}
			ps.SetPrice("usd/krw", sdk.NewDecCoinFromDec("krw", decAmount))
		}()
	}
}
