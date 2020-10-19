module github.com/node-a-team/terra-oracle

go 1.14

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/cosmos/cosmos-sdk v0.39.1
	github.com/spf13/cobra v1.0.0
	github.com/spf13/viper v1.7.1
	github.com/tendermint/go-amino v0.15.1
	github.com/tendermint/tendermint v0.33.7
	github.com/terra-project/core v0.4.0-rc.4
)

replace github.com/CosmWasm/go-cosmwasm => github.com/terra-project/go-cosmwasm v0.10.1-terra
