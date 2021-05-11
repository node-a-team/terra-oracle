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

// for bithumn
type Ticker_bithumb struct {
	Data struct {
		Closing_price	string `json:"closing_price"`
		Date		string `json:"date"`
	}
}



func (ps *PriceService) lunaToKrw(logger log.Logger) {

	coinone(ps, logger)
//	bithumb(ps, logger)
}

func bithumb(ps *PriceService, logger log.Logger) {

        for {
                func() {
                        defer func() {
                                if r := recover(); r != nil {
                                        logger.Error("Unknown error", r)

					// Abstain
					abstain(ps, logger)
                                }

                                time.Sleep(cfg.Config.Options.Interval.Luna * time.Second)
                        }()

//                         resp, err := http.Get("https://api.bithumb.com/public/ticker/luna_krws")
                        resp, err := http.Get(cfg.Config.APIs.Luna.Krw.Bithumb)
                        if err != nil {
                                logger.Error("Fail to fetch from coinone", err.Error())
                                panic(err) //return
                        }
                        defer func() {
                                resp.Body.Close()
                        }()

                        body, err := ioutil.ReadAll(resp.Body)
                        if err != nil {
                                logger.Error("Fail to read body", err.Error())
                                panic(err) //return
                        }

                        t := Ticker_bithumb{}
                        err = json.Unmarshal(body, &t)
                        if err != nil {
                                logger.Error("Fail to unmarshal json", err.Error())
                                panic(err) //return
                        }

			timestamp := time.Now().UTC().Unix()
                        logger.Info(fmt.Sprintf("Recent luna/krw: %s, timestamp: %d", t.Data.Closing_price, timestamp))

			decAmount, err := sdk.NewDecFromStr(t.Data.Closing_price)
                        if err != nil {
                                logger.Error("Fail to parse price to Dec")
				panic(err) //return
                        }

                        ps.SetPrice("luna/krw", sdk.NewDecCoinFromDec("krw", decAmount),  int64(timestamp))

                }()
        }

}



func coinone(ps *PriceService, logger log.Logger) {

	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("Unknown error", r)
				}

				time.Sleep(cfg.Config.Options.Interval.Luna * time.Second)
			}()

//			resp, err := http.Get("https://tb.coinone.co.kr/api/v1/tradehistory/recent/?market=krw&target=luna")
			resp, err := http.Get(cfg.Config.APIs.Luna.Krw.Coinone)
			if err != nil {
				logger.Error("Fail to fetch from coinone", err.Error())
				panic(err) //return
			}
			defer func() {
				resp.Body.Close()
			}()

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				logger.Error("Fail to read body", err.Error())
				panic(err) //return
			}

			th := TradeHistory{}
			err = json.Unmarshal(body, &th)
			if err != nil {
				logger.Error("Fail to unmarshal json", err.Error())
				panic(err) //return
			}

			timestamp := time.Now().UTC().Unix()

			trades := th.Trades
			recent := trades[len(trades)-1]
			logger.Info(fmt.Sprintf("Recent luna/krw: %s, timestamp: %d", recent.Price, timestamp))

			decAmount, err := sdk.NewDecFromStr(recent.Price)
			if err != nil {
				logger.Error("Fail to parse price to Dec")
				panic(err) //return
			}

			ps.SetPrice("luna/krw", sdk.NewDecCoinFromDec("krw", decAmount), int64(timestamp))
		}()
	}

}


func abstain(ps *PriceService, logger log.Logger) {
	// Abstain
	decAmount, _ := sdk.NewDecFromStr("0")

        timestamp := time.Now().UTC().Unix()
        logger.Info(fmt.Sprintf("Abstain luna/krw: %s, timestamp: %d", "0", timestamp))

        ps.SetPrice("luna/krw", sdk.NewDecCoinFromDec("krw", decAmount), int64(timestamp))

}
