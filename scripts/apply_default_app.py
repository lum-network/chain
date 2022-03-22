import logging
import socket
import argparse
import os
import toml

def apply_params(home):
    path = "{}/config/app.toml".format(home)
    data = toml.load(path)
    data['api']['rpc-read-timeout'] = 30
    data['api']['rpc-write-timeout'] = 30

    f = open(path, 'w')
    toml.dump(data, f)
    f.close()

def main():
    # Setup the logging instance
    logging.basicConfig(level=logging.INFO)

    # Setup the arguments
    parser = argparse.ArgumentParser(description='Apply default parameters to app.toml file')
    parser.add_argument('--home', type=str, help='The path to the chain root folder', required=True)
    args = parser.parse_args()

    apply_params(args.home)
    logging.info("Applied params to app.toml file")

if __name__ == "__main__":
    main()