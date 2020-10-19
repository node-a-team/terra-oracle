package price

import (
	"sync"

	sdk "github.com/cosmos/cosmos-sdk/types"

	//	cmn "github.com/tendermint/tendermint/libs/common"
	service "github.com/tendermint/tendermint/libs/service"
)

type PriceWithTimestamp struct {
	Px        sdk.DecCoin
	Timestamp int64
}

type PriceService struct {
	service.BaseService
	mutex  *sync.RWMutex
	prices map[string]PriceWithTimestamp
}

func NewPriceService() *PriceService {
	ps := &PriceService{
		mutex:  new(sync.RWMutex),
		prices: make(map[string]PriceWithTimestamp),
	}
	ps.BaseService = *service.NewBaseService(nil, "PriceService", ps)
	return ps
}

func (ps *PriceService) OnStart() error {
	// TODO: gracefully quit go routine
	go ps.coinoneToLuna(ps.Logger.With("market", "luna/krw"))
	go ps.sdrToKrw(ps.Logger.With("market", "sdr/krw"))
	go ps.usdToKrw(ps.Logger.With("market", "usd/krw"))
	go ps.mntToKrw(ps.Logger.With("market", "mnt/krw"))
	go ps.bandLunaToKrw(ps.Logger.With("band", "luna/krw"))
	go ps.fxsToKrw(ps.Logger.With("band", "fxs/krw"))
	return nil
}

func (ps *PriceService) GetPrice(market string) sdk.DecCoin {
	ps.mutex.RLock()
	defer func() {
		ps.mutex.RUnlock()
	}()
	return ps.prices[market].Px
}

func (ps *PriceService) SetPrice(market string, coin sdk.DecCoin, timestamp int64) {
	ps.mutex.Lock()
	defer func() {
		ps.mutex.Unlock()
	}()
	if timestamp > ps.prices[market].Timestamp {
		ps.prices[market] = PriceWithTimestamp{Px: coin, Timestamp: timestamp}
	}
}
