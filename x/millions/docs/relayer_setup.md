## Starting the project

Before doing the following, please make sure you have a running instance of the lum binary, either via ignite or as a daemon

Run the docker compose in a separate terminal

`docker-compose -f tools/docker-compose.yml up -d`

Wait for the containers to be health, then the relayer will start, you can check this by doing:

`docker ps`

Once the containers are healthy, the relayer container will pop up.

## Create the links between chains

First, enter the lum-relayer container bash

Grab the lum-relayer container id by doing `docker ps` then do (and replace the id inside):

`docker exec -it b60875bae07a bash`

Once inside, you can type the following to create the links (repeat for each chain)

### Create link for Gaia <=> Lum

`hermes create connection --a-chain lum-network-dev-43 --b-chain gaia-devnet`

`hermes create channel --a-chain lum-network-dev-43 --a-connection connection-0 --a-port transfer --b-port transfer`

Transactions may fail but will succeed at some point.

### Create link Juno <=> Lum

`hermes create connection --a-chain lumnetwork-testnet --b-chain juno-devnet`

`hermes create channel --a-chain lumnetwork-testnet --a-connection connection-1 --a-port transfer --b-port transfer`

### Create link Osmosis <=> Lum

`hermes create connection --a-chain lumnetwork-testnet --b-chain osmosis-devnet`

`hermes create channel --a-chain lumnetwork-testnet --a-connection connection-2 --a-port transfer --b-port transfer`

## Sending coins from Gaia to local Lum one

Find the ID of the Gaia container:

`docker ps` then search for "lum-gaia"

Then enter the container (replace ID inside)

`docker exec -it 773da473109d bash`

Then type

`gaiad tx ibc-transfer transfer transfer channel-0 lum16rlynj5wvzwts5lqep0je5q4m3eaepn58hmzu5 1000000uatom --from genesis_key --keyring-backend test --chain-id gaia-devnet --fees 200uatom`

If you type 

`lumd query bank balances lum1vvhncsjsdy2d7cnge2773kd49pkvvv9wsrt8ra`

You should see something like this with the IBC denom from Gaia

```
balances:
- amount: "1000000"
  denom: ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2
- amount: "100000000000"
  denom: uatom
- amount: "119900000000"
  denom: ulum
pagination:
  next_key: null
  total: "0"
```