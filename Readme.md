## How to use
For terra oracle, you can separate the validator and the feeder that send oracle transactions repeatably. To set feeder, you can use the cli command "terracli tx oracle set-feeder". To send transactions, it is necessary to find private key, so you should execute this software in an environment with your local wallet. But, I recommend separating the validator and the feeder and execute this in the local wallet that has the only feeder account.  
By default, Tendermint waits 10 seconds for the transaction to be committed. But this timeout is too short to detect the transaction was committed in 5 blocks (default voting period). So I recommend increasing timeout_broadcast_tx_commit option in config.toml.  
And make sure that you include ukrw in minimum gas price in terrad.toml to let users pay the fee by ukrw.  

## Changelog
#### v0.0.5-alpha.10
Add vote list for Proposal#83: `{DKK, NOK}`  

#### v0.0.5-alpha.9-b
Add vote list for Proposal#78: `{SEK}`(by <strong>@hanjukim</strong>)    
To use the Systemd service, you need to enter the `name` and `password` in the **[Feeder]** part of `config.toml`  

#### v0.0.5-alpha.8-a
Running `terra-oracle register-systemd` creates `terra-oracle_starter` in `$GOPATH` and creates `/etc/systemd/system/terra-oracle.service`   
Restart oracle when `reveal period` or `failed to execute message` occurs in rawlog  
Fix typos(Thanks <strong>@petes-fan-club</strong>)  

#### v0.0.5-alpha.7
Changed the price base of KRW from Bithumb to Coinone  

#### v0.0.5-alpha.6
Add vote list for Proposal#38: `{THB}`   
Configuration change in `config.toml`

#### v0.0.5-alpha.5
Patch terra oracle to use data from band guanyu mainnet(by <strong>@Benzbeeb</strong>)  
Make sure the band's API is `https://terra-lcd.bandchain.org` in `config.toml`.

#### v0.0.5-alpha.4
Update price server is using USD quote instead KRW & also seperate KRW from the USD price(by <strong>@YunSuk-Yeo</strong>)   
Code Simplification(`oracle/tx.go`)   

#### v0.0.5-alpha.3
The prices of all stable coins come from the API of the `currencylayer.com`  
Add vote list for Proposal#26: `{CNY, JPY, GBP, INR, CAD, CHF, HKD, AUD, SGD}`  
Added Band API activation option to `config.toml`(true/false)  
Configuration change in `config.toml`  

#### v0.0.5-alpha.2
Add EUR Oracle vote.   

#### v0.0.5-alpha.1
Added Bandchain API (by <strong>@prin-r</strong>)

#### v0.0.4-alpha.1
Add Flag in `Service` command for systemd service: `--vote-mode` (default `aggregate`)  
Dependency was updated. (terra-core v0.3.0 -> v0.4.0 / cosmos-sdk v0.39.1 / tendermint v0.33.7)  

#### v0.0.3-alpha.4
Fix error: "Fail to parse price to int    module=price market=luna/krw"

#### v0.0.3-alpha.3
Add Command: `terra-oracle version`  
Add Flag in `Service` command for systemd service: `--config`  

#### v0.0.3-alpha.2
Save settings in config.toml file.  
For MNT oracle we need the API Key of currencylayer. (https://currencylayer.com/product)  
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

## Install
```bash
cd $HOME
git clone https://github.com/node-a-team/terra-oracle.git
cd $HOME/terra-oracle 
go install ./cmd/terra-oracle

terra-oracle version
## v0.0.5-alpha.8-a
```

## Set your basic config for cli.

```bash
terracli config chain-id {chain_id}
terracli config node {endpoint_of_your_node_rpc}

ex)
terracli config chain-id columbus-4
terracli config node tcp://localhost:26657
```

## Set your feeder.

```bash
terracli tx oracle set-feeder {address_of_feeder} --from={name_of_validator_account} --fees 35610000ukrw 

// ex)
terracli tx oracle set-feeder terra1uq0z26lahq7ekavpf9cgl8ypxnj7ducat60a4w --from=VALIDATOR --fees 35610000ukrw 
```

## Start terra-oracle.
  
```sh
terra-oracle service --from={name_of_feeder} --broadcast-mode=block --config={path_to_config.toml} --vote-mode aggregate

// ex)
terra-oracle service --from=ORACLE --broadcast-mode=block --config=$HOME/terra-oracle --vote-mode aggregate
```

## Start with Systemd service.
1. You must enter the `name` and `password` in the `[Feeder]` part of `config.toml`  
2. You need to install `expect`: `sudo apt install expect -y`  
```sh
terra-oracle register-systemd --from={name_of_feeder} --config={path_to_config.toml}

// ex)
terra-oracle register-systemd --from=ORACLE --config=$HOME/terra-oracle
sudo systemctl start terra-oracle.service

// log)
journalctl -f -u terra-oracle.service

```
