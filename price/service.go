package price

import (
	"sync"

	sdk "github.com/cosmos/cosmos-sdk/types"
	cmn "github.com/tendermint/tendermint/libs/common"
)

type PriceService struct {
	cmn.BaseService
	mutex *sync.RWMutex
	prices map[string]sdk.DecCoin
}

func NewPriceService() *PriceService {
	ps := &PriceService{
		mutex: new(sync.RWMutex),
		prices: make(map[string]sdk.DecCoin),
	}
	ps.BaseService = *cmn.NewBaseService(nil, "PriceService", ps)
	return ps
}

func (ps *PriceService) OnStart() error {
	// TODO: gracefully quit go routine
	go ps.coinoneToLuna(ps.Logger.With("market", "luna/krw"))
	go ps.sdrToKrw(ps.Logger.With("market", "sdr/krw"))
	go ps.usdToKrw(ps.Logger.With("market", "usd/krw"))
	return nil
}

func (ps *PriceService) GetPrice(market string)sdk.DecCoin {
	ps.mutex.RLock()
	defer func() {
		ps.mutex.RUnlock()
	}()
	return ps.prices[market]
}

func (ps *PriceService) setPrice(market string, coin sdk.DecCoin) {
	ps.mutex.Lock()
	defer func() {
		ps.mutex.Unlock()
	}()
	ps.prices[market] = coin
}
