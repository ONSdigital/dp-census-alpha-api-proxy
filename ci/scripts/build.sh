#!/bin/bash -eux

pushd dp-census-alpha-api-proxy
  make build
  cp build/dp-census-alpha-api-proxy Dockerfile.concourse ../build
popd
