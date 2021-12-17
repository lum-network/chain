# CHANGELOG

All notable changes to the project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v1.0.5]

-   Critical upgrade fixing issues related to the Staking module and IBC
    -   Upgrade to IBC v2
    -   Fix Staking module initialization
-   Soft improvements
    -   Always enable api and telemetry
    -   Backport sync and node info rest endpoints for compatibility purpose

## [v1.0.4]

-   Software upgrade v1.0.4 has been added to the upgrade handler
-   Migration of the Airdrop KV from version 2 to version 3
-   Airdrop module updates:
    -   Recompute and update module account balance
    -   Fix claim actions to properly handle free vs vested coins
