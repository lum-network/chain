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

### Dev tools

- Run `bash ./scripts/install_tools.sh` to install dev tools
- Tools being used:
    - `gofmt` to format code
    - `goimports` to organize imports
    - `golangci-lint` to check for linting issues
    - `gosec` to check for security issues

## Useful links

* [Javascript client SDK](https://github.com/lum-network/sdk-javascript)
* [Explorer frontend code](https://github.com/lum-network/explorer)
* [Bridge code (aka explorer backend)](https://github.com/lum-network/chain-bridge)
* [Wallet code](https://github.com/lum-network/wallet)

## Mainnet

Information related to the Lum Network mainnet `lum-network-1` can be found in the [mainnet repository](https://github.com/lum-network/mainnet).

### v1.6.4 - `TODO` - Block `TODO`

`TODO`

### v1.6.4 - 2024-02-01 - Block 11390000
CosmosMillions: Make ICA channel restoration unlock all entities and revamp the fee system to allow for more than one fee taker.

[Upgrade guide here](https://github.com/lum-network/mainnet/blob/master/upgrades/v1.6.4/guide.md).

### v1.6.3 - 2023-11-30 - Block 10444000
CosmosMillions: Fix issues related to Withdrawals and Unbondings, unlock entities in stalling state.

[Upgrade guide here](https://github.com/lum-network/mainnet/blob/master/upgrades/v1.6.3/guide.md).

### v1.6.2 - 2023-11-02 - Block 10027000
CosmosMillions: Fix of bug related to Withdrawals and EpochUnbonding which caused some Withdrawals to remain in pending state.

Beam module: Deprecate beam module due to the drop of the use case.

DFract module: Clawback all DFRs as part of DFract sunsetting process as detailed in this [blog post](https://medium.com/lum-network/sunsetting-the-dfract-protocol-beta-version-a2277bce07fb).

[Upgrade guide here](https://github.com/lum-network/mainnet/blob/master/upgrades/v1.6.2/guide.md).

### v.1.6.1 - 2023-09-29 - Block [#9520750](https://www.mintscan.io/lum/blocks/9520750)
Refactoring of Pool lifecycle abstraction which should facilitate different pool type integration.

Improvements of Batch withdrawals with respect to remote zone unbonding frequency.

Manual seed generation consensus-based.

[Upgrade guide here](https://github.com/lum-network/mainnet/blob/master/upgrades/v1.6.1/guide.md).

### v.1.5.2 - 2023-08-04 - Block [#8688700](https://www.mintscan.io/lum/blocks/8688700)
This upgrade is essential for implementing the Epoch module and introducing the Millions batched withdrawals feature. Additionally, it includes an upgrade to Cosmos SDK v0.47.4 for enhanced functionality and stability.

[Upgrade guide here](https://github.com/lum-network/mainnet/blob/master/upgrades/v1.5.2/guide.md).

### v.1.5.1 - 2023-07-24 - Block [#8527300](https://www.mintscan.io/lum/blocks/8527300)
This upgrade fixes an issue introduced in the previous update and update Millions prize strategy.

[Upgrade guide here](https://github.com/lum-network/mainnet/blob/master/upgrades/v1.5.1/guide.md).

### v.1.5.0 - 2023-07-17 - Block [#8424000](https://www.mintscan.io/lum/blocks/8424000)
This upgrade introduces numerous changes over the whole Lum Network codebase, along with specific Millions improvements.

[Upgrade guide here](https://github.com/lum-network/mainnet/blob/master/upgrades/v1.5.0/guide.md).

### ~~v1.4.1~~ v1.4.2 - 2023-05-30 - Block [#7740000](https://www.mintscan.io/lum/blocks/7740000)
Fix IBC Huckleberry vulnerability and pruning issues.

[Upgrade guide here](https://github.com/lum-network/mainnet/blob/master/upgrades/v1.4.1/guide.md).

Version v1.4.1 has been replaced by **v1.4.2** (as decribed in the upgrade guide above) due to a chain halt on block #7799476.

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
