package oracle

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	operating "os"
	"time"

//	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/context"
//	"github.com/cosmos/cosmos-sdk/client/keys"
	utils "github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
//	authtxb "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"

	"github.com/tendermint/tendermint/rpc/client"

	"github.com/terra-project/core/x/oracle"
	cfg "github.com/node-a-team/terra-oracle/config"
)

const (
//	FlagValidator = "validator"
//	FlagSoftLimit = "change-rate-soft-limit"
//	FlagHardLimit = "change-rate-hard-limit"
)

func (os *OracleService) init() error {
	os.txBldr = authtypes.NewTxBuilderFromCLI().WithTxEncoder(authtypes.DefaultTxEncoder(os.cdc))
	os.cliCtx = context.NewCLIContext().
		WithCodec(os.cdc)

	if os.cliCtx.BroadcastMode != "block" {
		return errors.New("I recommend to use commit broadcast mode")
	}
/*
	fromName := os.cliCtx.GetFromName()
	_passphrase, err := keys.GetPassphrase(fromName)
	if err != nil {
		return err
	}
	os.passphrase = _passphrase
*/

	os.passphrase =cfg.Config.Feeder.Password

	os.changeRateSoftLimit = cfg.Config.Options.ChangeRateLimit.Soft
	if os.changeRateSoftLimit < 0 {
		return fmt.Errorf("Soft limit should be positive")
	}
	os.changeRateHardLimit = cfg.Config.Options.ChangeRateLimit.Hard
	if os.changeRateHardLimit < 0 {
		return fmt.Errorf("Hard limit should be positive")
	}

	return nil
}

func (os *OracleService) txRoutine() {
	httpClient := client.NewHTTP(os.cliCtx.NodeURI, "/websocket")

	var latestVoteHeight int64 = 0

	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					os.Logger.Error("Unknown error", r)
				}

				time.Sleep(1 * time.Second)
			}()

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

			denoms := []string{"krw", "usd", "sdr", "mnt"}

			if os.prevoteInited {
				os.Logger.Info(fmt.Sprintf("Try to send vote msg (including prevote for next vote msg)"))

				voteMsgs, err := os.makeVoteMsgs(denoms)
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

				if res.Logs[0].Success != true {
					os.prevoteInited = false
				}


				if tick > res.Height/VotePeriod {
					os.Logger.Error("Tx couldn't be sent within vote period")
				}
			//	 os.prevoteInited = false

			} else {
				os.Logger.Info(fmt.Sprintf("Try to send prevote msg"))

				prevoteMsgs, err := os.makePrevoteMsgs(denoms)
				if err != nil {
					os.Logger.Error("Fail to make prevote msgs", err.Error())
				}

				_, err = os.broadcast(prevoteMsgs)
				if err != nil {
					os.Logger.Error("Fail to send prevote msgs#2", err.Error())
					return
				}

				os.prevoteInited = true
			}
		}()
	}
}

func (os *OracleService) makePrevoteMsgs(denoms []string) ([]sdk.Msg, error) {
	feeder := os.cliCtx.GetFromAddress()
//	validator, err := sdk.ValAddressFromBech32(viper.GetString(FlagValidator))
	validator, err := sdk.ValAddressFromBech32(cfg.Config.Validator.OperatorAddr)
	if err != nil {
		return nil, fmt.Errorf("Invalid validator: %s", err.Error())
	}

	prevoteMsgs := make([]sdk.Msg, 0)
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
		voteHash, err := oracle.VoteHash(salt, price.Amount, "u"+denom, validator)
		if err != nil {
			return nil, fmt.Errorf("Fail to vote hash: %s", err.Error())
		}

		prevote := oracle.NewMsgExchangeRatePrevote(hex.EncodeToString(voteHash), "u"+denom, feeder, validator)
		prevoteMsgs = append(prevoteMsgs, prevote)

		os.preLunaPrices[denom] = os.lunaPrices[denom]
	}



	return prevoteMsgs, nil
}

func (os *OracleService) makeVoteMsgs(denoms []string) ([]sdk.Msg, error) {
	feeder := os.cliCtx.GetFromAddress()
//	validator, err := sdk.ValAddressFromBech32(viper.GetString(FlagValidator))
	validator, err := sdk.ValAddressFromBech32(cfg.Config.Validator.OperatorAddr)
	if err != nil {
		return nil, fmt.Errorf("Invalid validator: %s", err.Error())
	}

	voteMsgs := make([]sdk.Msg, 0)
	for _, denom := range denoms {
		price := os.preLunaPrices[denom]

		salt := os.salts[denom]
		if len(salt) == 0 {
			// It can occur before the first prevote was sent
			// So, this error may be temporary
			return nil, fmt.Errorf("Fail to get salt: %s", err.Error())
		}
		vote := oracle.NewMsgExchangeRateVote(price.Amount, salt, "u"+denom, feeder, validator)
		voteMsgs = append(voteMsgs, vote)
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
		voteHash, err := oracle.VoteHash(salt, price.Amount, "u"+denom, validator)
		if err != nil {
			return nil, fmt.Errorf("Fail to vote hash: %s", err.Error())
		}

		prevote := oracle.NewMsgExchangeRatePrevote(hex.EncodeToString(voteHash), "u"+denom, feeder, validator)
		voteMsgs = append(voteMsgs, prevote)

		os.preLunaPrices[denom] = os.lunaPrices[denom]
	}



	return voteMsgs, nil
}

func (os *OracleService) calculatePrice() (abort bool, err error) {
	lunaToKrw := os.ps.GetPrice("luna/krw")
	if lunaToKrw.Denom != "krw" {
		return false, errors.New("Can't get luna/krw")
	}

	if os.prevoteInited {
		preLunaKrw := os.preLunaPrices["krw"].Amount
		changeRate := lunaToKrw.Amount.Sub(preLunaKrw).Quo(preLunaKrw)
		os.Logger.Info(fmt.Sprintf("Change rate: %s", changeRate.String()))

		if os.changeRateHardLimit > 0 {
			hardLimit, err := sdk.NewDecFromStr(fmt.Sprintf("%f", os.changeRateHardLimit))
			if err != nil {
				return false, err
			}
			if changeRate.Abs().GT(hardLimit) {
				return true, fmt.Errorf("Change rate exceeds hard limit")
			}
		}

		if os.changeRateSoftLimit > 0 {
			softLimit, err := sdk.NewDecFromStr(fmt.Sprintf("%f", os.changeRateSoftLimit))
			if err != nil {
				return false, err
			}
			if changeRate.Abs().GT(softLimit) {
				os.Logger.Error("Change rate exceeds soft limit")
				lunaToKrw.Amount = preLunaKrw.Add(preLunaKrw.Mul(softLimit).Mul(sdk.NewDec(int64(changeRate.Sign()))))
				os.Logger.Info("Luna price is adjust by soft limit", lunaToKrw.String())
			}
		}
	}

	usdToKrw := os.ps.GetPrice("usd/krw")
	if usdToKrw.Denom != "krw" {
		return false, errors.New("Can't get usd/krw")
	}

	sdrToKrw := os.ps.GetPrice("sdr/krw")
	if usdToKrw.Denom != "krw" {
		return false, errors.New("Can't get sdr/krw")
	}

	mntToKrw := os.ps.GetPrice("mnt/krw")
        if usdToKrw.Denom != "krw" {
                return false, errors.New("Can't get mnt/krw")
        }

	// If usdToKrw is 0, this will panic
	lunaToUsdAmount := lunaToKrw.Amount.Quo(usdToKrw.Amount)
	lunaToUsd := sdk.NewDecCoinFromDec("usd", lunaToUsdAmount)

	// If sdrToKrw is 0, this will panic
	lunaToSdrAmount := lunaToKrw.Amount.Quo(sdrToKrw.Amount)
	lunaToSdr := sdk.NewDecCoinFromDec("sdr", lunaToSdrAmount)

	lunaToMntAmount := lunaToKrw.Amount.Quo(mntToKrw.Amount)
	lunaToMnt :=  sdk.NewDecCoinFromDec("mnt", lunaToMntAmount)

	os.Logger.Info(fmt.Sprintf("usd/krw: %s", usdToKrw.String()))
	os.Logger.Info(fmt.Sprintf("sdr/krw: %s", sdrToKrw.String()))
	os.Logger.Info(fmt.Sprintf("mnt/krw: %s", mntToKrw.String()))
	os.Logger.Info(fmt.Sprintf("luna/krw: %s", lunaToKrw.String()))
	os.Logger.Info(fmt.Sprintf("luna/usd: %s", lunaToUsd.String()))
	os.Logger.Info(fmt.Sprintf("luna/sdr: %s", lunaToSdr.String()))
	os.Logger.Info(fmt.Sprintf("luna/mnt: %s", lunaToMnt.String()))


	os.lunaPrices["krw"] = lunaToKrw
	os.lunaPrices["usd"] = lunaToUsd
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
	txBytes, err := txBldr.BuildAndSign(fromName, os.passphrase, msgs)
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
