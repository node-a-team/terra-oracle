package price

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/tendermint/tendermint/libs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	cfg "github.com/node-a-team/terra-oracle/config"
)

type BandResponse struct {
	Request BandResult `json:"request"`
}

type RawRequests struct {
	ExternalID   uint64 `json:"external_id,string"`
	DataSourceID uint64 `json:"data_source_id,string"`
	Calldata     []byte `json:"calldata,string"`
}

type IBCChannel struct {
	PortId    string `json:"port_id"`
	ChannelId string `json:"channel_id"`
}

type Request struct {
	OracleScriptID      uint64        `json:"oracle_script_id,string"`
	Calldata            []byte        `json:"calldata,string"`
	RequestedValidators []string      `json:"requested_validators"`
	MinCount            uint64        `json:"min_count,string"`
	RequestHeight       uint64        `json:"request_height,string"`
	RequestTime         int64         `json:"request_time,string"`
	ClientID            string        `json:"client_id"`
	RawRequests         []RawRequests `json:"raw_requests"`
	IBCChannel          *IBCChannel   `json:"ibc_channel"`
	ExecuteGas          uint64        `json:"executeGas,string"`
}

type RawReports struct {
	ExternalID uint64 `json:"external_id,string"`
	Data       string `json:"data"`
}

type Reports struct {
	Validator       string       `json:"validator"`
	InBeforeResolve bool         `json:"in_before_resolve"`
	RawReports      []RawReports `json:"raw_reports"`
}

type RequestPacketData struct {
	ClientID       string `json:"client_id"`
	OracleScriptID uint64 `json:"oracle_script_id,string"`
	Calldata       []byte `json:"calldata,string"`
	AskCount       uint64 `json:"ask_count,string"`
	MinCount       uint64 `json:"min_count,string"`
}

type ResponsePacketData struct {
	ClientID      string `json:"client_id"`
	RequestID     uint64 `json:"request_id,string"`
	AnsCount      uint64 `json:"ans_count,string"`
	RequestTime   uint64 `json:"request_time,string"`
	ResolveTime   uint64 `json:"resolve_time,string"`
	ResolveStatus uint8  `json:"resolve_status"`
	Result        []byte `json:"result,string"`
}

type PacketResult struct {
	ClientID       string `json:"client_id"`
	OracleScriptID int64  `json:"oracle_script_id,string"`
	Calldata       []byte `json:"calldata"`
	AskCount       uint64 `json:"ask_count,string"`
	MinCount       uint64 `json:"min_count,string"`
	RequestID      int64  `json:"request_id,string"`
	AnsCount       uint64 `json:"ans_count,string"`
	RequestTime    int64  `json:"request_time,string"`
	ResolveTime    int64  `json:"resolve_time,string"`
	ResolveStatus  string `json:"resolve_status"`
	Result         []byte `json:"result,string"`
}

type BandResult struct {
	Request Request      `json:"request"`
	Reports []Reports    `json:"reports"`
	Result  PacketResult `json:"result"`
}

type LunaPriceCallData struct {
	Multiplier uint64
}

type OrderBook struct {
	Ask int64 `json:"ask"`
	Bid int64 `json:"bid"`
	Mid int64 `json:"mid"`
}

type LunaPrice struct {
	OrderBooks []OrderBook `json:"order_books"`
}

func (lpcd *LunaPriceCallData) toBytes() []byte {
	b, err := Encode(*lpcd)
	if err != nil {
		panic(err)
	}
	return b
}

var (
	MULTIPLIER           = uint64(1000000000)
	LUNA_PRICE_CALLDATA  = LunaPriceCallData{Multiplier: MULTIPLIER}
	LUNA_PRICE_END_POINT = fmt.Sprintf("/oracle/v1/request_search?oid=13&calldata=%x&min_count=3&ask_count=4", LUNA_PRICE_CALLDATA.toBytes())
)

func (ps *PriceService) bandLunaToKrw(logger log.Logger) {
	for {
		if !cfg.Config.APIs.Band.Active {
			logger.Info("Warning APIs.Band.Active is false in Config.toml. Let's exit the bandLunaToKrw().")
			break
		}

		func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("Unknown error", r)
				}

				time.Sleep(cfg.Config.Options.Interval.Luna * time.Second)
			}()

			resp, err := http.Get(cfg.Config.APIs.Band.Band + LUNA_PRICE_END_POINT)
			if err != nil {
				logger.Error("Fail to fetch from band-luna", err.Error())
				return
			}
			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				logger.Error("Fail to read body", err.Error())
				return
			}

			br := BandResponse{}
			err = json.Unmarshal(body, &br)
			if err != nil {
				logger.Error("Fail to unmarshal json", err.Error())
				return
			}

			var lp LunaPrice
			Decode(br.Request.Result.Result, &lp)
			// Find median
			prices := []float64{}
			for _, order := range lp.OrderBooks[:4] {
				prices = append(prices, float64(order.Mid))
			}
			sort.Float64s(prices)
			medianPrice := (prices[1] + prices[2]) / float64(2*MULTIPLIER)

			// Create dec from float64
			decAmount, err := sdk.NewDecFromStr(strconv.FormatFloat(medianPrice, 'f', -1, 64))
			if err != nil {
				logger.Error("Fail to parse price to Dec")
			}

			price := sdk.NewDecCoinFromDec("krw", decAmount)
			timestamp := int64(br.Request.Result.ResolveTime)

			logger.Info(fmt.Sprintf("Recent luna/krw: %s, timestamp: %d", price, timestamp))

			ps.SetPrice("luna/krw", price, timestamp)
		}()
	}
}
