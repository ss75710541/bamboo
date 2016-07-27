#!/bin/bash

set -e

VERSION=$(cat VERSION)
IMAGE="bamboo-v${VERSION}"
HOST="catalog.shurenyun.com"
NAMESPACE="library"
TAG="omega.v2.5.7"

docker build --no-cache -t ${IMAGE}:${TAG} -f Dockerfile .
