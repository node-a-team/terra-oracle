package oracle

import (
	"time"

	"github.com/tendermint/go-amino"

	cmn "github.com/tendermint/tendermint/libs/common"

	"github.com/node-a-team/terra-oracle/price"
)

type OracleService struct {
	cmn.BaseService
	ps *price.PriceService
	cdc *amino.Codec
}

func NewOracleService(ps *price.PriceService, cdc *amino.Codec) *OracleService {
	os := &OracleService{
		ps: ps,
		cdc: cdc,
	}
	os.BaseService = *cmn.NewBaseService(nil, "OracleService", os)
	return os
}

func (os *OracleService) OnStart() error {
	err := os.init()
	if err != nil {
		return err
	}

	err = os.ps.Start()
	if err != nil {
		return err
	}

	// Wait a second until price service fetchs price initially
	time.Sleep(3 * time.Second)

	go os.txRoutine()

	return nil
}
