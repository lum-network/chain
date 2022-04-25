import logging
import argparse
import socket
import os
import toml

def apply_params(home, servers, height, hash):
    path = "{}/config/config.toml".format(home)
    data = toml.load(path)
    data['statesync']['enable'] = True
    data['statesync']['rpc_servers'] = servers
    data['statesync']['trust_height'] = height
    data['statesync']['trust_hash'] = hash
    data['statesync']['trust_period'] = '336h'

    f = open(path, 'w')
    toml.dump(data, f)
    f.close()

def main():
    # Setup the logging instance
    logging.basicConfig(level=logging.INFO)

    # Setup the arguments
    parser = argparse.ArgumentParser(description='Apply state sync parameters to config.toml file')
    parser.add_argument('--home', type=str, help='The path to the chain root folder', required=True)
    parser.add_argument('--servers', type=str, help='The snapshot servers to use', required=True)
    parser.add_argument('--height', type=str, help='The snapshot height', required=True)
    parser.add_argument('--hash', type=str, help='The snapshot hash', required=True)
    args = parser.parse_args()

    apply_params(args.home, args.servers, args.height, args.hash)
    logging.info("Applied params to config.toml file")

if __name__ == "__main__":
    main()