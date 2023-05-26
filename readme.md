# Chain - Lum Network Implementation
The first implementation of the Lum Network protocol, built on top of the Cosmos SDK.

[![codecov](https://codecov.io/gh/lum-network/chain/branch/master/graph/badge.svg)](https://codecov.io/gh/lum-network/chain)
[![Go Report Card](https://goreportcard.com/badge/github.com/lum-network/chain)](https://goreportcard.com/report/github.com/lum-network/chain)
[![license](https://img.shields.io/github/license/lum-network/chain.svg)](https://github.com/lum-network/chain/blob/main/LICENSE)
[![LoC](https://tokei.rs/b1/github.com/lum-network/chain)](github.com/lum-network/chain)
[![GolangCI](https://golangci.com/badges/github.com/lum-network/chain.svg)](https://golangci.com/r/github.com/lum-network/chain)

## Development

Lum Network is ignite compatible and thus, you will require it to start the project in development mode.

You can install it by following the instructions [here](https://docs.ignite.com/)

Then just use the following command to generate the protobuf and the chain binary

    ignite chain serve

Please also make sure to have a look to the [contributing guidelines](https://github.com/lum-network/chain/blob/master/CONTRIBUTING.md)

## Useful links

* [Javascript client SDK](https://github.com/lum-network/sdk-javascript)
* [Explorer frontend code](https://github.com/lum-network/explorer)
* [Bridge code (aka explorer backend)](https://github.com/lum-network/chain-bridge)
* [Wallet code](https://github.com/lum-network/wallet)

## Mainnet

Information related to the Lum Network mainnet `lum-network-1` can be found in the [mainnet repository](https://github.com/lum-network/mainnet).

### v1.4.1 - 2023-05-30 - Block [#7740000](https://www.mintscan.io/lum/blocks/7740000)
Fix IBC Huckleberry vulnerability and pruning issues.

[Upgrade guide here](https://github.com/lum-network/mainnet/blob/master/upgrades/v1.4.1/guide.md).

### v1.4.0 - 2023-05-24 - Block [#7652000](https://www.mintscan.io/lum/blocks/7652000)
Introduce new Millions module.

[Upgrade guide here](https://github.com/lum-network/mainnet/blob/master/upgrades/v1.4.0/guide.md).

### v1.3.1 - 2023-01-25 - Block [#5890500](https://www.mintscan.io/lum/blocks/5890500)
IBCFee module removal alongside minor improvements.

[Upgrade guide here](https://github.com/lum-network/mainnet/blob/master/upgrades/v1.3.1/guide.md).

### v1.3.0 - 2023-01-10 - Block [#5672000](https://www.mintscan.io/lum/blocks/5672000)
Cosmos v0.46.7 upgrade alongside Dragonberry security patch.

[Upgrade guide here](https://github.com/lum-network/mainnet/blob/master/upgrades/v1.3.0/guide.md).

### v1.2.1 - 2022-09-23 - Block [#4066500](https://www.mintscan.io/lum/blocks/4066500)
Introduce new DFract module.

[Upgrade guide here](https://github.com/lum-network/mainnet/blob/master/upgrades/v1.2.1/guide.md).

### v1.2.0 - 2022-09-19 - Block [#4009300](https://www.mintscan.io/lum/blocks/4009300)
Cosmos v0.45.7 upgrade alongside IBC v3 integration (interchain accounts).

[Upgrade guide here](https://github.com/lum-network/mainnet/blob/master/upgrades/v1.2.0/guide.md).

### v1.1.3 - 2022-04-29 - Block [#1960665](https://www.mintscan.io/lum/blocks/1960665)
Hard upgrade due to a chain halt following the v1.1.0 upgrade on Apr. 28 at 18:50 UTC.

[Upgrade guide here](https://github.com/lum-network/mainnet/blob/master/upgrades/v1.1.3/guide.md).

### v1.1.0 - 2022-04-28 - Block [#1960300](https://www.mintscan.io/lum/blocks/1960300)
Cosmos v0.45.0 upgrade alongside minor upgrades and fixes.

[Upgrade guide here](https://github.com/lum-network/mainnet/blob/master/upgrades/v1.1.0/guide.md).

### v1.0.5 - 2021-12-20 - Block [#90300](https://www.mintscan.io/lum/blocks/90300)
Critical upgrade that fixes issues related to the Staking module and IBC

[Upgrade guide here](https://github.com/lum-network/mainnet/blob/master/upgrades/v1.0.5/guide.md).

### v1.0.0 - 2021-12-14 - Block [#1](https://www.mintscan.io/lum/blocks/1)
Genesis launch of the Lum Network.
