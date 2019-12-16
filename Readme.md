## How to use
For terra oracle, you can separate the validator and the feeder that send oracle transactions repeatably. To set feeder, you can use the cli command "terracli tx oracle set-feeder". To send transactions, it is necessary to find private key, so you should execute this software in an environment with your local wallet. But, I recommend separating the validator and the feeder and execute this in the local wallet that has the only feeder account.  
By default, Tendermint waits 10 seconds for the transaction to be committed. But this timeout is too short to detect the transaction was committed in 5 blocks (default voting period). So I recommend increasing timeout_broadcast_tx_commit option in config.toml.  
And make sure that you include ukrw in minimum gas price in terrad.toml to let users pay the fee by ukrw.  

## Changelog
#### v0.0.3-alpha.2
Save settings in config.toml file.  
For MNT oracle we need the API value of currencylayer. (https://currencylayer.com/product)  
Added recovery procedure if prevote+vote transaction fails.

#### v0.0.3-alpha.1
VotePeriod = 12 -> VotePeriod = 5.  
Add MNT Oracle vote.  
Dependency was updated. (terra-core v0.2.3 -> v0.3.0-rc2)
#### v0.0.2-alpha.2
Soft/hard limit was added for rate of luna price chagne.  
(The rate of change is calculated per 12 blocks(voting period).  
If it exceeds the soft limit, an alert will occur,  
and the value which does not exceed the soft limit will be submitted continuously.  
If it exceeds the hard limit, the program will be exited.)  
Now vote msgs will be sent with prevote for next vote.  
#### v0.0.2-alpha.1
Dependency was updated. (terra-core v0.2.1 -> v0.2.3)  



This software used go module for dependency management, so you should locate this outside of the GOPATH or set GO111MODULE=on in environment variable set.  
Checkout https://github.com/golang/go/wiki/Modules  
```bash
git clone https://github.com/node-a-team/terra-oracle.git
cd terra-oracle 
go build ./cmd/terra-oracle
```

Set your basic config for cli.
```bash
terracli config chain-id {chain_id}
terracli config node {endpoint_of_your_node_rpc}

ex)
terracli config chain-id columbus-3
terracli config node tcp://localhost:26657
```

Set your feeder.
```bash
terracli tx oracle set-feeder {address_of_feeder} --from={name_of_validator_account} --gas=auto --gas-adjustment=1.25

// ex)
terracli tx oracle set-feeder terra1uq0z26lahq7ekavpf9cgl8ypxnj7ducat60a4w --from=VALIDATOR --gas=auto --gas-adjustment=1.25
```

Start service.
```sh
cd terra-oracle 
terra-oracle service --from={name_of_feeder} --fees=3000ukrw --gas=150000 --broadcast-mode=block

// ex)
cd terra-oracle 
./terra-oracle service --from=ORACLE --fees=3000ukrw --gas=150000 --broadcast-mode=block 
```
