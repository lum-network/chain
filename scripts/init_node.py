import sys
import os
import toml
import socket
import logging
import argparse
import hashlib

# Global variables
coinType = 880
hostname = socket.gethostname()


# This is utility function to replace entries in file
def replace_entries_in_file(file, find, replace):
    logging.info("Replacing {} by {} in {}".format(find, replace, file))

    fileIn = open(file, 'r')
    content = fileIn.read()
    fileIn.close()

    newContent = content.replace(find, replace)

    fileOut = open(file, 'w')
    fileOut.write(newContent)
    fileOut.close()


# This is utility function to replace common entries in the config file
def prepare_config_file(file):
    data = toml.load(file)
    # data['api']['enable'] = True
    # data['telemetry']['enabled'] = True
    data['instrumentation']['prometheus'] = True
    data['rpc']['laddr'] = "tcp://0.0.0.0:26657"

    f = open(file, 'w')
    toml.dump(data, f)
    f.close()


# This makes all the required steps for initializing a primary node
def init_primary(chain_id, home, mnemonic, action):
    logging.info("Initializing the primary node")

    # Create the destination folder if it does not exists
    if not os.path.exists(home):
        os.mkdir(home)

    # Init the node and the corresponding wallet
    os.system("lumd init --chain-id {} --home {} {} ".format(chain_id, home, hostname))
    os.system("lumd keys add bootnode --coin-type {} --home {}".format(coinType, home))

    # Init the faucet wallet
    if mnemonic is not None and len(mnemonic) > 0:
        os.system("echo '{}' | lumd keys add faucet --coin-type {} --recover --home {}".format(mnemonic, coinType, home))

    # Add the genesis accounts
    os.system("lumd add-genesis-account $(lumd keys show bootnode -a --home {}) 30000000000000ulum --home {}".format(home, home))
    if mnemonic is not None and len(mnemonic) > 0:
        os.system("lumd add-genesis-account $(lumd keys show faucet -a --home {}) 10000000000000ulum --home {}".format(home, home))

    # Edit the genesis file and replace stake with ulum
    replace_entries_in_file("{}/config/genesis.json".format(home), "stake", "ulum")

    # Prepare the configuration file
    prepare_config_file("{}/config/config.toml".format(home))

    # Generate the first transaction
    os.system("lumd gentx bootnode 100000000ulum --chain-id {} --home {}".format(chain_id, home))

    # Prepare the genesis
    os.system("lumd collect-gentxs --home {}".format(home))
    os.system("lumd validate-genesis --home {}".format(home))

    # Copy the genesis file to the shared storage
    if action == 'copy':
        logging.info("Copying genesis and node id files to shared storage")
        os.system("cp {}/config/genesis.json /config/genesis.json".format(home))

        logging.info("Copying the node id to shared storage")
        node_id = os.popen("lumd tendermint show-node-id --home {}".format(home))
        os.system("echo '{}@{}:26657' >> /config/primary_peer_id.txt".format(node_id, hostname))
    else:
        logging.error("Action download not available for primary node")

    logging.info("Primary node initialized")


# This makes all the steps for initializing a secondary node
def init_secondary(chain_id, home, action, path, checksum):
    logging.info("Initializing the secondary node")

    # Create the destination folder if it does not exists
    if not os.path.exists(home):
        os.mkdir(home)

    # Init the node
    os.system("lumd init --chain-id {} --home {} {} ".format(chain_id, home, hostname))

    # Prepare the configuration file
    prepare_config_file("{}/config/config.toml".format(home))

    if action == 'copy':
        logging.info("Copying genesis and node id files from shared storage")
        os.system("cp /config/genesis.json {}/config/genesis.json".format(home))
    elif action == 'download':
        logging.info("Downloading genesis from remote location")
        os.system("wget {} -O {}/config/genesis.json".format(path, home))

        # Ensure the checksum matches
        csum = hashlib.md5("{}/config/genesis.json".format(path)).hexdigest()
        if csum != checksum:
            logging.error("Both checksum does not match {} != {}".format(csum, checksum))

    logging.info('Secondary node initialized')


# This is the main file entrypoint
def main():
    # Setup the logging instance
    logging.basicConfig(level=logging.INFO)

    # Setup the arguments
    parser = argparse.ArgumentParser(description='Init a Lum Network chain node')
    parser.add_argument('--chainid', type=str, help='The chain ID to pass to nodes', required=True)
    parser.add_argument('--type', choices=['primary', 'secondary'], type=str, help='The type of node to init', required=True)
    parser.add_argument('--home', type=str, help='The path to the root folder of blockchain', required=True)
    parser.add_argument('--mnemonic', type=str, help='The faucet mnemonic used for bootnode')
    parser.add_argument('--action', type=str, choices=['copy', 'download'], help='Once generated, copy the genesis file to the shared storage', required=True)
    parser.add_argument('--path', type=str, help='The path from where to download the genesis file')
    parser.add_argument('--checksum', type=str, help='The downloaded genesis file checksum', required='--path' in sys.argv)
    args = parser.parse_args()

    if args.action == 'download' and (not args.path or not args.checksum):
        parser.error("--action download requires path and checksum")

    # Information logging
    logging.info("Hostname is {}".format(hostname))

    # What type of node do we want to init
    if args.type == "primary":
        init_primary(args.chainid, args.home, args.mnemonic, args.action)
    else:
        init_secondary(args.chainid, args.home, args.action, args.path, args.checksum)


if __name__ == "__main__":
    main()
