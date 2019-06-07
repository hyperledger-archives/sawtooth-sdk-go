# Copyright 2017 Intel Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# ------------------------------------------------------------------------------

# Description:
#   Builds an image to be used when developing in Go. The default CMD is to run
#   build_go.
#
# Build:
#   $ cd sawtooth-sdk-go/docker
#   $ docker build . -f sawtooth-build-go-protos -t sawtooth-build-go-protos
#
# Run:
#   $ cd sawtooth-sdk-go
#   $ docker run -v $(pwd):/project/sawtooth-sdk-go sawtooth-build-go-protos

FROM ubuntu:bionic

RUN apt-get update \
 && apt-get install gnupg -y

LABEL "install-type"="mounted"

RUN echo "deb [arch=amd64] http://repo.sawtooth.me/ubuntu/ci bionic universe" >> /etc/apt/sources.list \
 && echo 'deb http://ppa.launchpad.net/gophers/archive/ubuntu bionic main' >> /etc/apt/sources.list \
 && (apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys 8AA7AF1F1091A5FD \
 || apt-key adv --keyserver hkp://p80.pool.sks-keyservers.net:80 --recv-keys 8AA7AF1F1091A5FD) \
 && (apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys 308C15A29AD198E9 \
 || apt-key adv --keyserver hkp://p80.pool.sks-keyservers.net:80 --recv-keys 308C15A29AD198E9) \
 && apt-get update \
 && apt-get install -y -q \
    golang-1.11-go \
    git \
    libssl-dev \
    libzmq3-dev \
    openssl \
    protobuf-compiler \
    python3 \
    python3-grpcio \
    python3-grpcio-tools \
    python3-pkg-resources \
 && apt-get clean \
 && rm -rf /var/lib/apt/lists/*

ENV GOPATH=/go

ENV PATH=$PATH:/go/bin:/usr/lib/go-1.11/bin

RUN mkdir /go

RUN go get -u \
    github.com/btcsuite/btcd/btcec \
    github.com/golang/protobuf/proto \
    github.com/golang/protobuf/protoc-gen-go \
    github.com/golang/mock/gomock \
    github.com/golang/mock/mockgen \
    github.com/pebbe/zmq4 \
    github.com/satori/go.uuid

RUN mkdir -p /go/src/github.com/hyperledger/sawtooth-sdk-go

WORKDIR /go/src/github.com/hyperledger/sawtooth-sdk-go/

CMD go generate
