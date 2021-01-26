package oracle

import (
	"crypto/rand"
	//	"encoding/hex"
	"errors"
	"fmt"
	operating "os"
	"strings"
	"time"

	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/context"

	//	"github.com/cosmos/cosmos-sdk/client/keys"
	utils "github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	//	authtxb "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"

	client "github.com/tendermint/tendermint/rpc/client/http"

	cfg "github.com/node-a-team/terra-oracle/config"
	"github.com/terra-project/core/x/oracle"
)

const (
//	FlagValidator = "validator"
//	FlagSoftLimit = "change-rate-soft-limit"
//	FlagHardLimit = "change-rate-hard-limit"
)

var (
	voteMode               string
	salt, exchangeRatesStr = []string{"1234", ""}, []string{"", ""}
)

func (os *OracleService) init() error {

	os.passphrase = cfg.Config.Feeder.Password
	ioReader := strings.NewReader(os.passphrase)

	os.txBldr = authtypes.NewTxBuilderFromCLI(ioReader).WithTxEncoder(authtypes.DefaultTxEncoder(os.cdc))

	fmt.Printf("\n\033[31m")
	os.cliCtx = context.NewCLIContext().WithCodec(os.cdc)

	fmt.Printf("\033[0m")
	if os.cliCtx.BroadcastMode != "block" {
		return errors.New("I recommend to use commit broadcast mode")
	}

	voteMode = viper.GetString(cfg.VoteMode)
	if voteMode != "singular" && voteMode != "aggregate" {
		return errors.New("The vote mode must be set to \"singular\" or \"aggregate\"")
	}
	/*
		fromName := os.cliCtx.GetFromName()
		_passphrase, err := keys.GetPassphrase(fromName)
		if err != nil {
			return err
		}
		os.passphrase = _passphrase
	*/

	//	os.passphrase =cfg.Config.Feeder.Password

	/*
		os.changeRateSoftLimit = cfg.Config.Options.ChangeRateLimit.Soft
		if os.changeRateSoftLimit < 0 {
			return fmt.Errorf("Soft limit should be positive")
		}
		os.changeRateHardLimit = cfg.Config.Options.ChangeRateLimit.Hard
		if os.changeRateHardLimit < 0 {
			return fmt.Errorf("Hard limit should be positive")
		}
	*/
	return nil
}

func (os *OracleService) txRoutine() {
	httpClient, _ := client.New(os.cliCtx.NodeURI, "/websocket")

	var voteMsgs []sdk.Msg
	var latestVoteHeight int64 = 0

	denoms := []string{"krw", "usd", "aud", "cad", "chf", "cny", "gbp", "hkd", "inr", "jpy", "sgd", "eur", "sdr", "mnt"}

	for {
		func() {
			/*
				defer func() {
					if r := recover(); r != nil {
						os.Logger.Error("Unknown error", r)
					}

					time.Sleep(1 * time.Second)
				}()
			*/
			time.Sleep(1 * time.Second)

			status, err := httpClient.Status()
			if err != nil {
				os.Logger.Error("Fail to fetch status", err.Error())
				return
			}
			latestHeignt := status.SyncInfo.LatestBlockHeight

			var tick int64 = latestHeignt / VotePeriod
			if tick <= latestVoteHeight/VotePeriod {

				return
			}
			latestVoteHeight = latestHeignt
			os.Logger.Info(fmt.Sprintf("Tick: %d", tick))

			abort, err := os.calculatePrice()
			if err != nil {
				os.Logger.Error("Error when calculate price", err.Error())
			}
			if abort {
				operating.Exit(1)
			}

			os.Logger.Info(fmt.Sprintf("Try to send vote msg (including prevote for next vote msg)"))

			if voteMode == "singular" {
				voteMsgs, err = os.makeSingularVoteMsgs(denoms)
			} else if voteMode == "aggregate" {
				voteMsgs, err = os.makeAggregateVoteMsgs(denoms)
			}

			if err != nil {
				os.Logger.Error("Fail to make vote msgs", err.Error())
			}

			// Because vote tx includes prevote for next price,
			// use twice as much gas.
			res, err := os.broadcast(voteMsgs)
			if err != nil {
				os.Logger.Error("Fail to send vote msgs#1", err.Error())
				return
			}

			// reveal period of submitted vote do not match with registered prevote
			if strings.Contains(res.RawLog, "reveal period") {
				os.prevoteInited = false
			}

			if tick > res.Height/VotePeriod {
				os.Logger.Error("Tx couldn't be sent within vote period")
			}

		}()
	}
}

// ----------------------------------------------- singular
func (os *OracleService) makeSingularVoteMsgs(denoms []string) ([]sdk.Msg, error) {

	msgs := make([]sdk.Msg, 0)

	feeder := os.cliCtx.GetFromAddress()

	validator, err := sdk.ValAddressFromBech32(cfg.Config.Validator.OperatorAddr)
	if err != nil {
		return nil, fmt.Errorf("Invalid validator: %s", err.Error())
	}

	if os.prevoteInited {

		// voteMsgs
		for _, denom := range denoms {
			price := os.preLunaPrices[denom]

			salt := os.salts[denom]
			if len(salt) == 0 {
				// It can occur before the first prevote was sent
				// So, this error may be temporary
				return nil, fmt.Errorf("Fail to get salt: %s", err.Error())
			}
			vote := oracle.NewMsgExchangeRateVote(price.Amount, salt, "u"+denom, feeder, validator)
			msgs = append(msgs, vote)
		}

		for _, denom := range denoms {
			price := os.lunaPrices[denom]
			if price.Denom != denom {
				return nil, errors.New("Price is not initialized")
			}

			salt, err := generateRandomString(4)
			if err != nil {
				return nil, fmt.Errorf("Fail to generate salt: %s", err.Error())
			}
			os.salts[denom] = salt

			voteHash := oracle.GetVoteHash(salt, price.Amount, "u"+denom, validator)
			if err != nil {
				return nil, fmt.Errorf("Fail to vote hash: %s", err.Error())
			}

			prevote := oracle.NewMsgExchangeRatePrevote(voteHash, "u"+denom, feeder, validator)
			msgs = append(msgs, prevote)

			os.preLunaPrices[denom] = os.lunaPrices[denom]
		}

	}

	// preVote
	for _, denom := range denoms {
		price := os.lunaPrices[denom]
		if price.Denom != denom {
			return nil, errors.New("Price is not initialized")
		}

		salt, err := generateRandomString(4)
		if err != nil {
			return nil, fmt.Errorf("Fail to generate salt: %s", err.Error())
		}
		os.salts[denom] = salt
		voteHash := oracle.GetVoteHash(salt, price.Amount, "u"+denom, validator)

		prevote := oracle.NewMsgExchangeRatePrevote(voteHash, "u"+denom, feeder, validator)
		msgs = append(msgs, prevote)

		os.preLunaPrices[denom] = os.lunaPrices[denom]
	}

	os.prevoteInited = true

	return msgs, nil
}

// ----------------------------------------------- aggregate
func (os *OracleService) makeAggregateVoteMsgs(denoms []string) ([]sdk.Msg, error) {

	msgs := make([]sdk.Msg, 0)

	feeder := os.cliCtx.GetFromAddress()

	validator, err := sdk.ValAddressFromBech32(cfg.Config.Validator.OperatorAddr)
	if err != nil {
		return nil, fmt.Errorf("Invalid validator: %s", err.Error())
	}

	if os.prevoteInited {

		// voteMsgs
		aggregateVote := oracle.NewMsgAggregateExchangeRateVote(salt[0], exchangeRatesStr[0], feeder, validator)
		msgs = append(msgs, aggregateVote)
	}

	// preVote
	salt[1], err = generateRandomString(4)
	if err != nil {
		return nil, fmt.Errorf("Fail to generate salt: %s", err.Error())
	}

	for i, denom := range denoms {

		price := os.lunaPrices[denom]
		if price.Denom != denom {
			return nil, errors.New("Price is not initialized")
		}

		if i == len(denoms)-1 {
			exchangeRatesStr[1] = exchangeRatesStr[1] + fmt.Sprint(price)
		} else {
			exchangeRatesStr[1] = exchangeRatesStr[1] + fmt.Sprint(price) + ","
		}

		exchangeRatesStr[1] = strings.Replace(exchangeRatesStr[1], price.Denom, "u"+price.Denom, -1)
	}

	aggregateVoteHash := oracle.GetAggregateVoteHash(salt[1], exchangeRatesStr[1], validator)
	aggregatePreVote := oracle.NewMsgAggregateExchangeRatePrevote(aggregateVoteHash, feeder, validator)
	msgs = append(msgs, aggregatePreVote)

	os.prevoteInited = true

	salt[0] = salt[1]
	exchangeRatesStr[0] = exchangeRatesStr[1]
	exchangeRatesStr[1] = ""

	return msgs, nil
}

func (os *OracleService) calculatePrice() (abort bool, err error) {

	lunaToKrw := os.ps.GetPrice("luna/krw")
	if lunaToKrw.Denom != "krw" {
		return false, errors.New("Can't get luna/krw")
	}

	usdToKrw := os.ps.GetPrice("usd/krw")
	if usdToKrw.Denom != "krw" {
		return false, errors.New("Can't get usd/krw")
	}

	audToKrw := os.ps.GetPrice("aud/krw")
	if audToKrw.Denom != "krw" {
		return false, errors.New("Can't get aud/krw")
	}

	cadToKrw := os.ps.GetPrice("cad/krw")
	if cadToKrw.Denom != "krw" {
		return false, errors.New("Can't get cad/krw")
	}

	chfToKrw := os.ps.GetPrice("chf/krw")
	if chfToKrw.Denom != "krw" {
		return false, errors.New("Can't get chf/krw")
	}

	cnyToKrw := os.ps.GetPrice("cny/krw")
	if cnyToKrw.Denom != "krw" {
		return false, errors.New("Can't get cny/krw")
	}

	eurToKrw := os.ps.GetPrice("eur/krw")
	if eurToKrw.Denom != "krw" {
		return false, errors.New("Can't get eur/krw")
	}

	gbpToKrw := os.ps.GetPrice("gbp/krw")
	if gbpToKrw.Denom != "krw" {
		return false, errors.New("Can't get gbp/krw")
	}

	hkdToKrw := os.ps.GetPrice("hkd/krw")
	if hkdToKrw.Denom != "krw" {
		return false, errors.New("Can't get hkd/krw")
	}

	inrToKrw := os.ps.GetPrice("inr/krw")
	if inrToKrw.Denom != "krw" {
		return false, errors.New("Can't get inr/krw")
	}

	jpyToKrw := os.ps.GetPrice("jpy/krw")
	if jpyToKrw.Denom != "krw" {
		return false, errors.New("Can't get jpy/krw")
	}

	sgdToKrw := os.ps.GetPrice("sgd/krw")
	if sgdToKrw.Denom != "krw" {
		return false, errors.New("Can't get sgd/krw")
	}

	sdrToKrw := os.ps.GetPrice("sdr/krw")
	if sdrToKrw.Denom != "krw" {
		return false, errors.New("Can't get sdr/krw")
	}

	mntToKrw := os.ps.GetPrice("mnt/krw")
	if mntToKrw.Denom != "krw" {
		return false, errors.New("Can't get mnt/krw")
	}

	// If usdToKrw is 0, this will panic
	lunaToUsdAmount := lunaToKrw.Amount.Quo(usdToKrw.Amount)
	lunaToUsd := sdk.NewDecCoinFromDec("usd", lunaToUsdAmount)

	// If audToKrw is 0, this will panic
	lunaToAudAmount := lunaToKrw.Amount.Quo(audToKrw.Amount)
	lunaToAud := sdk.NewDecCoinFromDec("aud", lunaToAudAmount)

	// If cadToKrw is 0, this will panic
	lunaToCadAmount := lunaToKrw.Amount.Quo(cadToKrw.Amount)
	lunaToCad := sdk.NewDecCoinFromDec("cad", lunaToCadAmount)

	// If chfToKrw is 0, this will panic
	lunaToChfAmount := lunaToKrw.Amount.Quo(chfToKrw.Amount)
	lunaToChf := sdk.NewDecCoinFromDec("chf", lunaToChfAmount)

	// If cnyToKrw is 0, this will panic
	lunaToCnyAmount := lunaToKrw.Amount.Quo(cnyToKrw.Amount)
	lunaToCny := sdk.NewDecCoinFromDec("cny", lunaToCnyAmount)

	// If eurToKrw is 0, this will panic
	lunaToEurAmount := lunaToKrw.Amount.Quo(eurToKrw.Amount)
	lunaToEur := sdk.NewDecCoinFromDec("eur", lunaToEurAmount)

	// If gbpToKrw is 0, this will panic
	lunaToGbpAmount := lunaToKrw.Amount.Quo(gbpToKrw.Amount)
	lunaToGbp := sdk.NewDecCoinFromDec("gbp", lunaToGbpAmount)

	// If hkdToKrw is 0, this will panic
	lunaToHkdAmount := lunaToKrw.Amount.Quo(hkdToKrw.Amount)
	lunaToHkd := sdk.NewDecCoinFromDec("hkd", lunaToHkdAmount)

	// If inrToKrw is 0, this will panic
	lunaToInrAmount := lunaToKrw.Amount.Quo(inrToKrw.Amount)
	lunaToInr := sdk.NewDecCoinFromDec("inr", lunaToInrAmount)

	// If jpyToKrw is 0, this will panic
	lunaToJpyAmount := lunaToKrw.Amount.Quo(jpyToKrw.Amount)
	lunaToJpy := sdk.NewDecCoinFromDec("jpy", lunaToJpyAmount)

	// If sgdToKrw is 0, this will panic
	lunaToSgdAmount := lunaToKrw.Amount.Quo(sgdToKrw.Amount)
	lunaToSgd := sdk.NewDecCoinFromDec("sgd", lunaToSgdAmount)

	// If sdrToKrw is 0, this will panic
	lunaToSdrAmount := lunaToKrw.Amount.Quo(sdrToKrw.Amount)
	lunaToSdr := sdk.NewDecCoinFromDec("sdr", lunaToSdrAmount)

	// If mntToKrw is 0, this will panic
	lunaToMntAmount := lunaToKrw.Amount.Quo(mntToKrw.Amount)
	lunaToMnt := sdk.NewDecCoinFromDec("mnt", lunaToMntAmount)

	os.Logger.Info(fmt.Sprintf("usd/krw: %s", usdToKrw.String()))
	os.Logger.Info(fmt.Sprintf("aud/krw: %s", audToKrw.String()))
	os.Logger.Info(fmt.Sprintf("cad/krw: %s", cadToKrw.String()))
	os.Logger.Info(fmt.Sprintf("chf/krw: %s", chfToKrw.String()))
	os.Logger.Info(fmt.Sprintf("cny/krw: %s", cnyToKrw.String()))
	os.Logger.Info(fmt.Sprintf("eur/krw: %s", eurToKrw.String()))
	os.Logger.Info(fmt.Sprintf("gbp/krw: %s", gbpToKrw.String()))
	os.Logger.Info(fmt.Sprintf("hkd/krw: %s", hkdToKrw.String()))
	os.Logger.Info(fmt.Sprintf("inr/krw: %s", inrToKrw.String()))
	os.Logger.Info(fmt.Sprintf("jpy/krw: %s", jpyToKrw.String()))
	os.Logger.Info(fmt.Sprintf("sgd/krw: %s", sgdToKrw.String()))
	os.Logger.Info(fmt.Sprintf("mnt/krw: %s", mntToKrw.String()))
	os.Logger.Info(fmt.Sprintf("luna/krw: %s", lunaToKrw.String()))
	os.Logger.Info(fmt.Sprintf("luna/usd: %s", lunaToUsd.String()))
	os.Logger.Info(fmt.Sprintf("luna/aud: %s", lunaToAud.String()))
	os.Logger.Info(fmt.Sprintf("luna/cad: %s", lunaToCad.String()))
	os.Logger.Info(fmt.Sprintf("luna/chf: %s", lunaToChf.String()))
	os.Logger.Info(fmt.Sprintf("luna/cny: %s", lunaToCny.String()))
	os.Logger.Info(fmt.Sprintf("luna/eur: %s", lunaToEur.String()))
	os.Logger.Info(fmt.Sprintf("luna/gbp: %s", lunaToGbp.String()))
	os.Logger.Info(fmt.Sprintf("luna/hkd: %s", lunaToHkd.String()))
	os.Logger.Info(fmt.Sprintf("luna/inr: %s", lunaToInr.String()))
	os.Logger.Info(fmt.Sprintf("luna/jpy: %s", lunaToJpy.String()))
	os.Logger.Info(fmt.Sprintf("luna/sgd: %s", lunaToSgd.String()))
	os.Logger.Info(fmt.Sprintf("luna/sdr: %s", lunaToSdr.String()))
	os.Logger.Info(fmt.Sprintf("luna/mnt: %s", lunaToMnt.String()))

	os.lunaPrices["krw"] = lunaToKrw
	os.lunaPrices["usd"] = lunaToUsd
	os.lunaPrices["aud"] = lunaToAud
	os.lunaPrices["cad"] = lunaToCad
	os.lunaPrices["chf"] = lunaToChf
	os.lunaPrices["cny"] = lunaToCny
	os.lunaPrices["eur"] = lunaToEur
	os.lunaPrices["gbp"] = lunaToGbp
	os.lunaPrices["hkd"] = lunaToHkd
	os.lunaPrices["inr"] = lunaToInr
	os.lunaPrices["jpy"] = lunaToJpy
	os.lunaPrices["sgd"] = lunaToSgd
	os.lunaPrices["sdr"] = lunaToSdr
	os.lunaPrices["mnt"] = lunaToMnt

	return false, nil
}

func (os *OracleService) broadcast(msgs []sdk.Msg) (*sdk.TxResponse, error) {

	txBldr, err := utils.PrepareTxBuilder(os.txBldr, os.cliCtx)
	if err != nil {
		return nil, err
	}

	fromName := os.cliCtx.GetFromName()

	// build and sign the transaction
	fmt.Printf("\n\033[31m")
	txBytes, err := txBldr.BuildAndSign(fromName, os.passphrase, msgs)
	fmt.Printf("\033[0m")
	if err != nil {
		return nil, err
	}

	// broadcast to a Tendermint node
	res, err := os.cliCtx.BroadcastTx(txBytes)
	if err != nil {
		return nil, err
	}

	return &res, os.cliCtx.PrintOutput(res)
}

// GenerateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

// GenerateRandomString returns a securely generated random string.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func generateRandomString(n int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	bytes, err := generateRandomBytes(n)
	if err != nil {
		return "", err
	}
	for i, b := range bytes {
		bytes[i] = letters[b%byte(len(letters))]
	}
	return string(bytes), nil
}
