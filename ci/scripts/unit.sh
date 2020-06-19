#!/bin/bash -eux

pushd dp-census-alpha-api-proxy
  make test
popd
