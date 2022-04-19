# Build image
FROM golang:1.18-alpine AS build-env

# Setup
ENV PACKAGES curl make git libc-dev bash gcc linux-headers eudev-dev python3 curl nano lz4 jq

# Set working directory for the build
WORKDIR /go/src/github.com/lum-network/chain

# Add source files
COPY . .

# Display used go version
RUN go version

# Patch any issue with go mod file
RUN go mod tidy && go mod download

# Install minimum necessary dependencies, build Cosmos SDK, remove packages
RUN apk add --no-cache $PACKAGES && make install

# Final image
FROM alpine:edge

# Specify used env
ENV CHAIN /chain

# Install dependencies
RUN apk add --update ca-certificates zip python3 py3-pip curl jq lz4
RUN pip3 install pyyaml toml

RUN addgroup chain && adduser -S -G chain chain -h "$CHAIN"

# Escalate to user
USER chain

# Change the working directory to the env
WORKDIR $CHAIN

# Copy over binaries from the build-env
COPY --from=build-env /go/bin/lumd /usr/bin/lumd

# Add the scripts
COPY --from=build-env /go/src/github.com/lum-network/chain/scripts/init_node.py /usr/bin/init_node.py
COPY --from=build-env /go/src/github.com/lum-network/chain/scripts/apply_default_config.py /usr/bin/apply_default_config.py
COPY --from=build-env /go/src/github.com/lum-network/chain/scripts/apply_pruning_config.py /usr/bin/apply_pruning_config.py
COPY --from=build-env /go/src/github.com/lum-network/chain/scripts/apply_rpc_node_config.py /usr/bin/apply_rpc_node_config.py
COPY --from=build-env /go/src/github.com/lum-network/chain/scripts/apply_statesync_params.py /usr/bin/apply_statesync_params.py

# Run lumd by default, omit entrypoint to ease using container with chaincli
CMD ["lumd"]