# RCoin

RCoin is a new cryptocurrency using blockchain and all that stuff.

## What works and what doesn't

* RPC
* Peer connections
* Mining
* Transfers
* Difficulty calculation

TODO:

* Partial nodes without full blockchain

## Details

* Dynamic difficulty based on blockchain height and last 2 block times
* Keys on Edwards Curve (ed12559)
* Uses IPFS Pub-Sub for communication

## Requirements

* Golang 1.9+
* IPFS (running with --enable-pubsub-experiment)

# Starting RCoin

If you have not initialized ipfs
```
>ipfs init
```

Start RCoin
```
>ipfs daemon --enable-pubsub-experiment
>rcoind
```

You can now open the tkwallet