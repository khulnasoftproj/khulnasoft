#!/usr/bin/env bash

set -eux

version=${1:-}
if [ "$version" = latest ]; then
	docker build -t khulnasoftproj-khulnasoft-dev -f Dockerfile-prebuilt .
else
	GOOS=linux go build -o dist/khulnasoft-docker ./cmd/khulnasoft
	docker build -t khulnasoftproj-khulnasoft-dev .
fi

docker run --rm -ti khulnasoftproj-khulnasoft-dev bash
