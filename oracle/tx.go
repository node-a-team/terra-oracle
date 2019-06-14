package oracle

import (
	"fmt"
	"time"
	"errors"
	"crypto/rand"
	"encoding/hex"

	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/utils"
	authtxb "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/client/keys"

	"github.com/tendermint/tendermint/rpc/client"

	"github.com/terra-project/core/x/oracle"
)

const (
	FlagValidator = "validator"
)

const VotePeriod = 12

var passphrase string
var txBldr authtxb.TxBuilder
var cliCtx context.CLIContext
var salts map[string]string = make(map[string]string)
var lunaPrices map[string]sdk.DecCoin = make(map[string]sdk.DecCoin)

func (os *OracleService) init() error {
	txBldr = authtxb.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(os.cdc))
	cliCtx = context.NewCLIContext().
		WithCodec(os.cdc).
		WithAccountDecoder(os.cdc)

	if cliCtx.BroadcastMode != "block" {
		return errors.New("I recommend to use commit broadcast mode")
	}

	fromName := cliCtx.GetFromName()
	_passphrase, err := keys.GetPassphrase(fromName)
	if err != nil {
		return err
	}
	passphrase = _passphrase
	return nil
}

func (os *OracleService) txRoutine() {
	httpClient := client.NewHTTP(cliCtx.NodeURI, "/websocket")

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
			if tick <= latestVoteHeight / VotePeriod {
				return
			}
			latestVoteHeight = latestHeignt
			
			lunaToKrw := os.ps.GetPrice("luna/krw")
			if lunaToKrw.Denom != "krw" {
				os.Logger.Error("Can't get luna/krw")
				return
			}
			
			usdToKrw := os.ps.GetPrice("usd/krw")
			if usdToKrw.Denom != "krw" {
				os.Logger.Error("Can't get usd/krw")
				return
			}

			sdrToKrw := os.ps.GetPrice("sdr/krw")
			if usdToKrw.Denom != "krw" {
				os.Logger.Error("Can't get sdr/krw")
				return
			}
			
			// If usdToKrw is 0, this will panic
			lunaToUsdAmount := lunaToKrw.Amount.Quo(usdToKrw.Amount)
			lunaToUsd := sdk.NewDecCoinFromDec("usd", lunaToUsdAmount)

			// If sdrToKrw is 0, this will panic
			lunaToSdrAmount := lunaToKrw.Amount.Quo(sdrToKrw.Amount)
			lunaToSdr := sdk.NewDecCoinFromDec("sdr", lunaToSdrAmount)

			os.Logger.Info(fmt.Sprintf("usd/krw: %s", usdToKrw.String()))
			os.Logger.Info(fmt.Sprintf("sdr/krw: %s", sdrToKrw.String()))
			os.Logger.Info(fmt.Sprintf("luna/krw: %s", lunaToKrw.String()))
			os.Logger.Info(fmt.Sprintf("luna/usd: %s", lunaToUsd.String()))
			os.Logger.Info(fmt.Sprintf("luna/sdr: %s", lunaToSdr.String()))

			feeder := cliCtx.GetFromAddress()
			validator, err := sdk.ValAddressFromBech32(viper.GetString(FlagValidator))
			if err != nil {
				os.Logger.Error("Invalid validator", err.Error())
				return
			}
			denoms := []string{"krw", "usd", "sdr"}
			os.Logger.Info(fmt.Sprintf("Tick: %d", tick))
			if (tick % 2 == 0) {
				lunaPrices["krw"] = lunaToKrw
				lunaPrices["usd"] = lunaToUsd
				lunaPrices["sdr"] = lunaToSdr

				prevoteMsgs := make([]sdk.Msg, 0)
				for _, denom := range denoms {
					price := lunaPrices[denom]
					if price.Denom != denom {
						os.Logger.Error("???")
						return 
					}

					salt, err := GenerateRandomString(4)
					if err != nil {
						os.Logger.Error("Fail to generate salt", err.Error())
						return
					}
					salts[denom] = salt
					voteHash, err := oracle.VoteHash(salt, price.Amount, "u" + denom, validator)
					if err != nil {
						os.Logger.Error("Fail to vote hash", err.Error())
						return
					}

					prevote := oracle.NewMsgPricePrevote(hex.EncodeToString(voteHash), "u" + denom, feeder, validator)
					prevoteMsgs = append(prevoteMsgs, prevote)
				}

				err = Broadcast(prevoteMsgs)
				if err != nil {
					os.Logger.Error("Fail to send prevote msgs", err.Error())
					return
				}
			}	

			if (tick % 2 == 1) {
				voteMsgs := make([]sdk.Msg, 0)
				for _, denom := range denoms {
					price := lunaPrices[denom]

					salt := salts[denom]
					if len(salt) == 0 {
						// It can occur before the first prevote was sent
						// So, this error may be temporary
						os.Logger.Error("Fail to get salt", err.Error())
						return
					}
					vote := oracle.NewMsgPriceVote(price.Amount, salt,"u" + denom, feeder, validator)
					voteMsgs = append(voteMsgs, vote)
				}

				err = Broadcast(voteMsgs)
				if err != nil {
					os.Logger.Error("Fail to send vote msgs", err.Error())
					return
				}
			}
		}()
	}
}

func Broadcast(msgs []sdk.Msg) error {
	txBldr, err := utils.PrepareTxBuilder(txBldr, cliCtx)
	if err != nil {
		return err
	}

	fromName := cliCtx.GetFromName()

	// build and sign the transaction
	txBytes, err := txBldr.BuildAndSign(fromName, passphrase, msgs)
	if err != nil {
		return err
	}

	// broadcast to a Tendermint node
	res, err := cliCtx.BroadcastTx(txBytes)
	if err != nil {
		return err
	}

	return cliCtx.PrintOutput(res)
}

// GenerateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomBytes(n int) ([]byte, error) {
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
func GenerateRandomString(n int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	bytes, err := GenerateRandomBytes(n)
	if err != nil {
		return "", err
	}
	for i, b := range bytes {
		bytes[i] = letters[b%byte(len(letters))]
	}
	return string(bytes), nil
}
