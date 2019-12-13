module github.com/node-a-team/terra-oracle

go 1.12

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/cosmos/cosmos-sdk v0.37.4
	github.com/spf13/viper v1.4.0
	github.com/tendermint/go-amino v0.15.0
	github.com/tendermint/tendermint v0.32.7
	github.com/terra-project/core v0.3.0
)

replace github.com/node-a-team/terra-oracle/config => /data_terra/terra/terra-oracle/config
