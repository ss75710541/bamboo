#!/bin/bash
export GOROOT=/usr/lib/go
export GOPATH=/gopath
export GOBIN=/gopath/bin
export PATH=$PATH:$GOROOT/bin:$GOPATH/bin
export GO15VENDOREXPERIMENT=1

cd /gopath/src/github.com/QubitProducts/bamboo
go build && \
mkdir -p /var/bamboo && \
cp  /gopath/src/github.com/QubitProducts/bamboo/bamboo /var/bamboo/bamboo && \
mkdir -p /var/log/supervisor && \
cd /

rm -rf /tmp/* /var/tmp/*
rm -f /etc/ssh/ssh_host_*
rm -rf /gopath
rm -rf /usr/lib/go
