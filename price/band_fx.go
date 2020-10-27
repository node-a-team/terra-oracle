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

func (ps *PriceService) fxsToKrw(logger log.Logger) {
	for {
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
				bytes.NewBuffer([]byte(`{"symbols":["KRW","XDR","MNT"],"min_count":3,"ask_count":4}`)),
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

			rates := map[string]PriceWithTimestamp{}
			for _, rate := range res.Result {
				symbol := strings.ToLower(rate.Symbol)
				if symbol == "xdr" {
					symbol = "sdr"
				}
				decAmount, err := sdk.NewDecFromStr(strconv.FormatFloat(rate.Multiplier/rate.Px, 'f', -1, 64))
				if err != nil {
					logger.Error("Fail to parse price to Dec", err.Error())
					return
				}
				rates[symbol] = PriceWithTimestamp{Px: sdk.NewDecCoinFromDec("krw", decAmount), Timestamp: rate.ResolveTime}
			}

			rates["mnt"] = PriceWithTimestamp{Px: sdk.NewDecCoinFromDec("krw", rates["krw"].Px.Amount.Quo(rates["mnt"].Px.Amount)), Timestamp: rates["mnt"].Timestamp}
			rates["sdr"] = PriceWithTimestamp{Px: sdk.NewDecCoinFromDec("krw", rates["krw"].Px.Amount.Quo(rates["sdr"].Px.Amount)), Timestamp: rates["sdr"].Timestamp}

			ps.SetPrice("usd/krw", rates["krw"].Px, rates["krw"].Timestamp)
			ps.SetPrice("mnt/krw", rates["mnt"].Px, rates["mnt"].Timestamp)
			ps.SetPrice("sdr/krw", rates["sdr"].Px, rates["sdr"].Timestamp)

			logger.Info(
				fmt.Sprintf("Recent [[usd/krw:%v,timestamp:%d], [mnt/krw:%v,timestamp:%d], [sdr/krw:%v,timestamp:%d]]",
					rates["krw"].Px, rates["krw"].Timestamp,
					rates["mnt"].Px, rates["mnt"].Timestamp,
					rates["sdr"].Px, rates["sdr"].Timestamp,
				))
		}()
	}
}
