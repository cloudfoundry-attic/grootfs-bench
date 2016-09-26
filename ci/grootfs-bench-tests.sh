#!/bin/bash -e

export GOPATH=$PWD:$GOPATH

cd src/code.cloudfoundry.org/grootfs-bench
if ! [ -d vendor ]; then
  glide install
fi

echo "I AM BENCH" | grootsay

ginkgo -r -p -race
