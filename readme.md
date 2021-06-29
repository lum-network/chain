# Chain - Lum Network Implementation
The first implementation of the Lum Network protocol, built on top of the Cosmos SDK.

Lum Network is a customer reward protocol and introduces mutable data channel in immutable ledgers.

[![codecov](https://codecov.io/gh/lum-network/chain/branch/master/graph/badge.svg)](https://codecov.io/gh/lum-network/chain)
[![Go Report Card](https://goreportcard.com/badge/github.com/lum-network/chain)](https://goreportcard.com/report/github.com/lum-network/chain)
[![license](https://img.shields.io/github/license/lum-network/chain.svg)](https://github.com/lum-network/chain/blob/main/LICENSE)
[![LoC](https://tokei.rs/b1/github.com/lum-network/chain)](github.com/lum-network/chain)
[![GolangCI](https://golangci.com/badges/github.com/lum-network/chain.svg)](https://golangci.com/r/github.com/lum-network/chain)

## Development

Lum Network is starport compatible and thus, you will require it to start the project in development mode.

You can install it by following the instructions [here](https://docs.starport.network/intro/install.html)

Then just use the following command to generate the protobuf and the chain binary

    starport serve

Please also make sure to have a look to the [contributing guidelines](https://github.com/lum-network/chain/blob/master/CONTRIBUTING.md)

## Useful links

* [Javascript client SDK](https://github.com/lum-network/sdk-javascript)
* [Explorer frontend code](https://github.com/lum-network/explorer)
* [Bridge code (aka explorer backend)](https://github.com/lum-network/chain-bridge)
* [Wallet code](https://github.com/lum-network/wallet)