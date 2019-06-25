## How to use
For terra oracle, you can separate the validator and the feeder that send oracle transactions repeatably. To set feeder, you can use the cli command "terracli tx oracle set-feeder". To send transactions, it is necessary to find private key, so you should execute this software in an environment with your local wallet. But, I recommend separating the validator and the feeder and execute this in the local wallet that has the only feeder account.  
By default, Tendermint waits 10 seconds for the transaction to be committed. But this timeout is too short to detect the transaction was committed in 12 blocks (default voting period). So I recommend increasing timeout_broadcast_tx_commit option in config.toml.  
And make sure that you include ukrw in minimum gas price in terrad.toml to let users pay the fee by ukrw.  


This software used go module for dependency management, so you should locate this outside of the GOPATH or set GO111MODULE=on in environment variable set.  
Checkout https://github.com/golang/go/wiki/Modules  
```sh
go install ./cmd/terra-oracle
```

Set your basic config for cli.
```sh
terracli config chain-id columbus-2
terracli config node {endpoint_of_your_node_rpc}
```

Set your feeder.
```sh
terracli tx oracle set-feeder --from={name_of_validator_account} --feeder={address_of_feeder} --gas=auto --gas-adjustment=1.25
```

Start service.
```sh
terra-oracle service --from {name_of_feeder} --fees 1000ukrw --gas 60000 --broadcast-mode block --validator terravaloper1~~~~~~~
```

![terra-oracle](https://user-images.githubusercontent.com/16339680/59500255-0800ec80-8ed4-11e9-88f1-2f706b7888a6.png)
