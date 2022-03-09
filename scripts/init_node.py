import logging
import argparse
import socket
import os
import toml

# Global variables
coinType = 880

# Global  definitions
class Config:
    chainId: str = ''
    homePath: str = ''
    genesisUrl: str = ''
    snapshotUrl: str = ''
    withSnapshot: bool = False

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
    data['api']['enable'] = True
    data['telemetry']['enabled'] = True
    data['instrumentation']['prometheus'] = True
    data['rpc']['laddr'] = "tcp://0.0.0.0:26657"

    f = open(file, 'w')
    toml.dump(data, f)
    f.close()

# Here we do the things
def initialize_node(config):
    # Create the destination folder if it does not exists
    if not os.path.exists(config.homePath):
        os.mkdir(config.homePath)

    # Initialize the node
    logging.info("Initialize the node")
    os.system("lumd init --chain-id {} --home {} {}".format(config.chainId, config.homePath, config.hostname))

    # Download the genesis file
    logging.info("Download the genesis file")
    os.system("curl -s '{}' > {}/config/genesis.json".format(config.genesisUrl, config.homePath))

    if config.withSnapshot:
        # Clear the old data directory
        logging.info("Delete the old data folder")
        os.system("rm -rf {}/data".format(config.homePath))

        # Download the snapshot
        logging.info("Download the snapshot file")
        os.system("curl -s '{}' > {}/snapshot.tar.lz4".format(config.snapshotUrl, config.homePath))

        # Install it
        logging.info("Decompress the snapshot file")
        os.system("cd {} && lz4 -d snapshot.tar.lz4 | tar xf -".format(config.homePath))

# Main entrypoint
def main():
    # Setup the logging instance
    logging.basicConfig(level=logging.INFO)

    # Setup the arguments
    parser = argparse.ArgumentParser(description='Init a Lum Network passive node')
    parser.add_argument('--chainid', type=str, help='The chain ID to pass to binary arguments', required=True)
    parser.add_argument('--home', type=str, help='The path to the chain root folder', required=True)
    parser.add_argument('--withsnapshot', action=argparse.BooleanOptionalAction, help='Should the script download and install snapshot or not', required=True)
    parser.add_argument('--genesisurl', type=str, help='The URL to the genesis file to download', required=True)
    parser.add_argument('--snapshoturl', type=str, help='The QuickSync snapshot URL to start from')
    args = parser.parse_args()

    # Config object initialize
    config = Config()
    config.chainId = args.chainid
    config.homePath = args.home
    config.genesisUrl = args.genesisurl
    config.snapshotUrl = args.snapshoturl
    config.hostname = socket.gethostname()
    config.withSnapshot = args.withsnapshot

    if config.withSnapshot:
        if not config.snapshotUrl:
            parser.error("--withSnapshot true requires a --snapshoturl argument")

    # Debug print
    logging.info("Machine hostname is {}".format(config.hostname))
    logging.info("Initializing node on network {}".format(config.chainId))
    logging.info("Using {} as home path".format(config.homePath))
    logging.info("Downloading snapshot {}".format(config.withSnapshot))

    # initialize_node(config)

if __name__ == "__main__":
    main()