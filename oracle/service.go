package oracle

import (
	"time"

	"github.com/tendermint/go-amino"

//	cmn "github.com/tendermint/tendermint/libs/common"
	service "github.com/tendermint/tendermint/libs/service"

	"github.com/node-a-team/terra-oracle/price"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
//	authtxb "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
	authtxb "github.com/cosmos/cosmos-sdk/x/auth/types"
)

const VotePeriod = 5
//const VotePeriod = 5

type OracleService struct {
	service.BaseService
	ps  *price.PriceService
	cdc *amino.Codec

	passphrase          string
	txBldr              authtxb.TxBuilder
	cliCtx              context.CLIContext
	lunaPrices          map[string]sdk.DecCoin
	prevoteInited       bool
//	changeRateSoftLimit float64
//	changeRateHardLimit float64

	salts         map[string]string
	preLunaPrices map[string]sdk.DecCoin
}

func NewOracleService(ps *price.PriceService, cdc *amino.Codec) *OracleService {
	os := &OracleService{
		ps:            ps,
		cdc:           cdc,
		salts:         make(map[string]string),
		lunaPrices:    make(map[string]sdk.DecCoin),
		preLunaPrices: make(map[string]sdk.DecCoin),
	}
	os.BaseService = *service.NewBaseService(nil, "OracleService", os)
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

//	go os.txRoutine()
	go os.txRoutine()

	return nil
}
