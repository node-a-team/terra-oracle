package price

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"
	"strconv"

	"github.com/tendermint/tendermint/libs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	cfg "github.com/node-a-team/terra-oracle/config"
)

func (ps *PriceService) mntToKrw(logger log.Logger) {


	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("Unknown error", r)
				}

				time.Sleep(cfg.Config.Options.Interval * time.Second)
			}()

//			resp, err := http.Get("http://www.apilayer.net/api/live?access_key=f4f5c16e99a0f32baeab5be8ced1cd39")
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

			usdToKrw := getUsdPrice(string(body), "KRW")
                        usdToMnt := getUsdPrice(string(body), "MNT")
			mntToKrw := usdToKrw / usdToMnt

			price  := strconv.FormatFloat(mntToKrw, 'f', -1, 64)


			logger.Info(fmt.Sprintf("Recent mnt/krw: %s", price))

			decAmount, err := sdk.NewDecFromStr(price)
			if err != nil {
				logger.Error("Fail to parse price to Dec", err.Error())
				return
			}
			ps.SetPrice("mnt/krw", sdk.NewDecCoinFromDec("krw", decAmount))
		}()
	}
}

func getUsdPrice(apiBody string, currency string) float64 {

	re, _ := regexp.Compile("\"USD" +currency +"\":[0-9.]+")
        str := re.FindString(string(apiBody))

        re, _ = regexp.Compile("[0-9.]+")

        return stringToFloat64(re.FindString(str))

}

func stringToFloat64(str string) float64 {

        var floatResult float64

        floatResult, _ = strconv.ParseFloat(str, 64)

        return floatResult
}

