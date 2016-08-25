#!/bin/bash -e

export GOPATH=$PWD:$GOPATH

cd src/code.cloudfoundry.org/grootfs-bench

grootsay I AM BENCH
go get -t ./...
ginkgo -r
