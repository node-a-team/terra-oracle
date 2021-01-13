package main

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/libs/cli"
	tenderOS "github.com/tendermint/tendermint/libs/os"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/terra-project/core/app"
	"github.com/terra-project/core/types/util"

//	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	flags "github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"

	_ "github.com/terra-project/core/client/lcd/statik"

	"github.com/node-a-team/terra-oracle/oracle"
	"github.com/node-a-team/terra-oracle/price"

	cfg "github.com/node-a-team/terra-oracle/config"
)

var (
	version = "v0.0.5-alpha.2"
	logger = log.NewTMLogger(log.NewSyncWriter(os.Stdout))
)

func main() {
	// Configure cobra to sort commands
	cobra.EnableCommandSorting = false

	// Instantiate the codec for the command line application
	cdc := app.MakeCodec()


	// Read in the configuration file for the sdk
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(util.Bech32PrefixAccAddr, util.Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(util.Bech32PrefixValAddr, util.Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(util.Bech32PrefixConsAddr, util.Bech32PrefixConsPub)
	config.Seal()

	rootCmd := &cobra.Command{
		Use: "terra-oracle",
	}

	// Add --chain-id to persistent flags and mark it required
	rootCmd.PersistentFlags().String(flags.FlagChainID, "", "Chain ID of tendermint node")
	rootCmd.PersistentPreRunE = func(_ *cobra.Command, _ []string) error {
		return initConfig(rootCmd)
	}

	// Construct Root Command
	rootCmd.AddCommand(
		rpc.StatusCommand(),
		svcCmd(cdc),
		versionCmd(),
		flags.LineBreak,
		keys.Commands(),
	)

	// Add flags and prefix all env exposed with GA
	executor := cli.PrepareMainCmd(rootCmd, "TE", app.DefaultCLIHome)

	err := executor.Execute()
	if err != nil {
		fmt.Printf("Failed executing CLI command: %s, exiting...\n", err)
		os.Exit(1)
	}


}

func svcCmd(cdc *amino.Codec) *cobra.Command {
	svcCmd := &cobra.Command{
		Use:   "service",
		Short: "Transactions subcommands",
		RunE: func(cmd *cobra.Command, args []string) error {
			ps := price.NewPriceService()
			ps.SetLogger(logger.With("module", "price"))

			os := oracle.NewOracleService(ps, cdc)
			os.SetLogger(logger.With("module", "oracle"))

			// Stop upon receiving SIGTERM or CTRL-C.
			tenderOS.TrapSignal(logger, func() {
				if os.IsRunning() {
					os.Stop()
				}
			})

			// Read in configuration file for local config.toml
			cfg.Init()

			if err := os.Start(); err != nil {
				return fmt.Errorf("Failed to start node: %v", err)
			}

			// Run forever.
			select {}
		},
	}

//	svcCmd.Flags().String(oracle.FlagValidator, "", "")
//	svcCmd.Flags().Float64(oracle.FlagSoftLimit, 0, "")
//	svcCmd.Flags().Float64(oracle.FlagHardLimit, 0, "")

	svcCmd.Flags().String(cfg.ConfigPath, "", "Directory for config.toml")
	svcCmd.MarkFlagRequired(cfg.ConfigPath)

	svcCmd.Flags().StringP(cfg.VoteMode, "", "aggregate", "Vote mode (singular|aggregate)")
        svcCmd.MarkFlagRequired(cfg.VoteMode)

	svcCmd = flags.PostCommands(svcCmd)[0]
	svcCmd.MarkFlagRequired(flags.FlagFrom)
//	svcCmd.MarkFlagRequired(oracle.FlagValidator)

	return svcCmd
}


func versionCmd() *cobra.Command {
        versionCmd := &cobra.Command{
                Use:   "version",
                Short: "Version check",
                Run: func(cmd *cobra.Command, args []string)  {
			fmt.Println(version)
                },
        }

	return versionCmd
}


func initConfig(cmd *cobra.Command) error {
	home, err := cmd.PersistentFlags().GetString(cli.HomeFlag)
	if err != nil {
		return err
	}

	cfgFile := path.Join(home, "config", "config.toml")
	if _, err := os.Stat(cfgFile); err == nil {
		viper.SetConfigFile(cfgFile)

		if err := viper.ReadInConfig(); err != nil {
			return err
		}
	}
	if err := viper.BindPFlag(flags.FlagChainID, cmd.PersistentFlags().Lookup(flags.FlagChainID)); err != nil {
		return err
	}
	if err := viper.BindPFlag(cli.EncodingFlag, cmd.PersistentFlags().Lookup(cli.EncodingFlag)); err != nil {
		return err
	}
	return viper.BindPFlag(cli.OutputFlag, cmd.PersistentFlags().Lookup(cli.OutputFlag))
}
