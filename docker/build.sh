#!/bin/sh -e

apk add --no-cache --update git mercurial bzr make go ca-certificates musl-dev

export GOPATH="/go"
export PATH="${GOPATH}/bin:/usr/local/go/bin:${PATH}"

go get github.com/docker-infra/reefer
go get github.com/tools/godep

cd /go/src/github.com/forestscribe/collectd-docker/collector
godep restore
go get github.com/forestscribe/collectd-docker/collector/...

cd /

cp /go/bin/influxdb-docker-collector /usr/bin/influxdb-docker-collector

apk del git mercurial bzr make go musl-dev

rm -rf /go /usr/local/go
rm -rf /var/cache/apk
