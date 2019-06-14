module github.com/node-a-team/terra-oracle

go 1.12

require (
	github.com/cosmos/cosmos-sdk v0.0.0-00010101000000-000000000000
	github.com/rakyll/statik v0.1.6
	github.com/spf13/cobra v0.0.3
	github.com/spf13/viper v1.3.2
	github.com/tendermint/go-amino v0.14.1
	github.com/tendermint/tendermint v0.31.5
	github.com/terra-project/core v0.2.1
)

replace github.com/cosmos/cosmos-sdk => github.com/YunSuk-Yeo/cosmos-sdk v0.34.7-terra

replace golang.org/x/crypto => github.com/tendermint/crypto v0.0.0-20180820045704-3764759f34a5
