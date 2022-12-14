`/tools` contains unmaintained docker services to help working on the ICA implementation on the Lum Network blockchain.

## Docker compose

Worth to mention that the whole docker compose state is being reset at each start (cold or hot)

### Starting the project

Starting the docker compose file is as simple as running

`docker-compose -f tools/docker-compose.yml up`

This will start three zones that we believe are representatives enough of the ecosystem's diversity : Osmosis, Juno and Gaia

Once these three zones reach an healthy state (verified by HTTP calls on RPC endpoints), the relayer starts.

The relayer is being initialized by using the same mnemonics defined on the environment file `.env`.
These accounts were created specifically for the development purposes, and do not exists on production mainnets.

### Creating IBC links

It is your responsability to init IBC links between the chains you want to use for development.

It can be done using the following command on the relayer container (you would need to replace parameters)

`hermes create channel --a-chain ibc-0 --b-chain ibc-1 --a-port transfer --b-port transfer --new-client-connection`


### Lum Node

The docker compose does not start the lum chain on purpose, you must do it by yourself, either by using Ignite CLI or raw commands.
Then you must create the IBC links and fill up the relayer with the relevant configuration.

TO BE CONTINUED WITH INFORMATION