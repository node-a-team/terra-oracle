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

// TradeHistory response from coinone
type TradeHistory struct {
	Trades []Trade `json:"trades"`
}

// Trade response from coinone
type Trade struct {
	Timestamp     uint64 `json:"timestamp"`
	Price         string `json:"price"`
	Volume        string `json:"volume"`
	IsSellerMaker bool   `json:"is_seller_maker"`
}

func (ps *PriceService) lunaToKrw(logger log.Logger) {
	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("Unknown error", r)
				}

				time.Sleep(cfg.Config.Options.Interval * time.Second)
			}()

			// resp, err := http.Get("https://tb.coinone.co.kr/api/v1/tradehistory/recent/?market=krw&target=luna")
			resp, err := http.Get(cfg.Config.APIs.Luna.Krw.Coinone)
			if err != nil {
				logger.Error("Fail to fetch from coinone", err.Error())
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

			th := TradeHistory{}
			err = json.Unmarshal(body, &th)
			if err != nil {
				logger.Error("Fail to unmarshal json", err.Error())
				return
			}

			trades := th.Trades
			recent := trades[len(trades)-1]
			logger.Info(fmt.Sprintf("Recent luna/krw: %s, timestamp: %d", recent.Price, recent.Timestamp))

			decAmount, err := sdk.NewDecFromStr(recent.Price)
			if err != nil {
				logger.Error("Fail to parse price to Dec")
			}

			ps.SetPrice("luna/krw", sdk.NewDecCoinFromDec("krw", decAmount), int64(recent.Timestamp))
		}()
	}
}
