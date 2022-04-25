import logging
import socket
import argparse
import os
import toml

def apply_params(home, pruning, keepRecent=None, keepEvery=None, interval=None):
    path = "{}/config/app.toml".format(home)
    data = toml.load(path)
    data['pruning'] = pruning

    if pruning == "custom":
        if not keepRecent or not keepEvery or not interval:
            raise ValueError('missing_parameters')

        data['pruning-keep-recent'] = keepRecent
        data['pruning-keep-every'] = keepEvery
        data['pruning-interval'] = interval

    f = open(path, 'w')
    toml.dump(data, f)
    f.close()

def main():
    # Setup the logging instance
    logging.basicConfig(level=logging.INFO)

    # Setup the arguments
    parser = argparse.ArgumentParser(description='Apply default parameters to app.toml file')
    parser.add_argument('--home', type=str, help='The path to the chain root folder', required=True)
    parser.add_argument('--pruning', type=str, help='The pruning level to use', required=True)
    parser.add_argument('--keep-recent', type=str, help='The pruning keep recent settings', required=False)
    parser.add_argument('--keep-every', type=str, help='The pruning keep every settings', required=False)
    parser.add_argument('--interval', type=str, help='The pruning interval settings', required=False)
    args = parser.parse_args()

    apply_params(args.home, args.pruning, args.keeprecent, args.keepevery, args.interval)
    logging.info("Applied params to app.toml file")

if __name__ == "__main__":
    main()