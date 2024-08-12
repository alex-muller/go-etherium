FROM alpine:3

WORKDIR /work

RUN apk add libc6-compat
RUN wget https://go.dev/dl/go1.22.6.linux-amd64.tar.gz

RUN tar -C /usr/local -xzf go1.22.6.linux-amd64.tar.gz && rm go1.22.6.linux-amd64.tar.gz

ENV PATH="$PATH:/usr/local/go/bin:/root/go/bin"
ENV GOPATH="/root/go/"

RUN go install github.com/ethereum/go-ethereum/cmd/abigen@latest

RUN wget https://github.com/ethereum/solidity/releases/download/v0.8.19/solc-static-linux
RUN mv solc-static-linux /usr/local/bin/solc && chmod a+x /usr/local/bin/solc
