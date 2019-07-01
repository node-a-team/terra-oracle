package price

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/tendermint/tendermint/libs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (ps *PriceService) sdrToKrw(logger log.Logger) {
	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("Unknown error", r)
				}

				time.Sleep(30 * time.Second)
			}()

			resp, err := http.Get("https://www.imf.org/external/np/fin/data/rms_five.aspx?tsvflag=Y")
			if err != nil {
				logger.Error("Fail to fetch from imf", err.Error())
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

			re, _ := regexp.Compile("Korean won[\\s]+[0-9.,]+")
			strs := re.FindAllString(string(body), 2)
			if len(strs) < 2 {
				logger.Error("Fail to find sdr-won")
				return
			}
			re, _ = regexp.Compile("[0-9.,]+")
			price := re.FindString(strs[1])
			price = strings.ReplaceAll(price, ",", "")

			logger.Info(fmt.Sprintf("Recent sdr/krw: %s", price))

			decAmount, err := sdk.NewDecFromStr(price)
			if err != nil {
				logger.Error("Fail to parse price to Dec", err.Error())
				return
			}
			ps.SetPrice("sdr/krw", sdk.NewDecCoinFromDec("krw", decAmount))
		}()
	}
}
