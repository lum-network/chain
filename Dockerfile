# Build image
FROM golang:1.16-alpine AS build-env

# Setup
ENV PACKAGES curl make git libc-dev bash gcc linux-headers eudev-dev python3 curl

# Set working directory for the build
WORKDIR /go/src/github.com/lum-network/chain

# Add source files
COPY . .

# Display used go version
RUN go version

# Install minimum necessary dependencies, build Cosmos SDK, remove packages
RUN apk add --no-cache $PACKAGES && make install

# Final image
FROM alpine:edge

# Specify used env
ENV CHAIN /chain

# Install dependencies
RUN apk add --update ca-certificates zip python3 py3-pip
RUN pip3 install pyyaml

RUN addgroup chain && adduser -S -G chain chain -h "$CHAIN"

# Escalate to user
USER chain

# Change the working directory to the env
WORKDIR $CHAIN

# Copy over binaries from the build-env
COPY --from=build-env /go/bin/chaind /usr/bin/chaind

# Run lumd by default, omit entrypoint to ease using container with chaincli
CMD ["chaind"]