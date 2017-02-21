FROM debian:jessie

COPY . /go/src/github.com/forestscribe/collectd-docker

RUN /go/src/github.com/forestscribe/collectd-docker/docker/build.sh

ENTRYPOINT ["/run.sh"]
