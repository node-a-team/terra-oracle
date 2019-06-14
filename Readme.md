## How to use
```
go install ./cmd/terra-oracle
terra-oracle service --from {name_of_feeder} --fees 1000ukrw --gas 60000 --broadcast-mode block --validator terravaloper1~~~~~~~
```

This should be executed in an environment with your local  wallet.  
This should wait long enough for the transaction to be committed, so it is recommended to increase the timeout_broadcast_tx_commit option sufficiently.
