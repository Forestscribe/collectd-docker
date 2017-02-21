package main

import (
	"flag"
	"log"
	"os"
	"strconv"

	"path"

	"github.com/forestscribe/collectd-docker/collector"
	"github.com/fsouza/go-dockerclient"
)

func main() {
	var err error
	var default_interval int
	default_interval, err = strconv.Atoi(collector.Getenv("COLLECT_INTERVAL", "10"))
	if err != nil {
		log.Fatal(err)
	}
	e := flag.String("endpoint", "unix:///var/run/docker.sock", "docker endpoint")
	c := flag.String("cert", "", "cert path for tls")
	h := flag.String("dburl", collector.Getenv("INFLUXDB_URL", ""), "influxdb server to report (env: INFLUXDB_URL)")
	db := flag.String("db", collector.Getenv("INFLUXDB_DATABASE", ""), "influxdb db where to report (env: INFLUXDB_DATABASE)")
	i := flag.Int("interval", default_interval, "interval to report (env: COLLECT_INTERVAL)")
	flag.Parse()

	if *h == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	var client *docker.Client

	if *c != "" {
		client, err = docker.NewTLSClient(*e, path.Join(*c, "cert.pem"), path.Join(*c, "key.pem"), path.Join(*c, "ca.pem"))
	} else {
		client, err = docker.NewClient(*e)
	}

	if err != nil {
		log.Fatal(err)
	}

	writer := collector.NewInfluxdbWriter(*h, *db, os.Getenv("INFLUXDB_USERNAME"), os.Getenv("INFLUXDB_PASSWORD"))

	collector := collector.NewCollector(client, writer, *i)

	err = collector.Run(5)
	if err != nil {
		log.Fatal(err)
	}
}
