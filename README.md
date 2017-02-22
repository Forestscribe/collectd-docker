# Collect docker container resource usage in Marathon/Mesos

This is a docker image to collect resource usage from docker
containers. Resource usage collected from `docker stats` API and sent to
influxdb installation. Containers can be added and removed on the fly, no need
to restart the service.

## History

Based on work from @bobrik https://github.com/bobrik/collectd-docker with following enhancements:

- Use Alpine instead of Debian for a 9MB compressed docker image
- Do not use collectd framework, but rather dump directly into influxdb
- Use Influxdb instead of graphite
    - Influxdb query langage is arguably easier to understand, and grafana support is better.
    - Influxdb has arbitrary tags support that match well with marathon groups
    - This drops the 64 characters limitation for app+task names.
- Works by default on Marathon without configuration (no need anymore to configure COLLECTD_DOCKER_APP_ENV)
- Create influxdb tags from marathon groups, and taskid splits

## Configuration

This plugin treats containers as tasks that run as parts of apps.

### Setting the App name of a Container

* Set the label `app_id` directly on the container
* Set `collectd_docker_app_label` on the container that points to which actual
label to use. e.g.`collectd_docker_app_label=app_id` will use `app_id` label on the
container
* Set environment variable `MARATHON_APP_ID` on the container

* Set `COLLECTD_DOCKER_APP_ENV` on the container that points to which actual
environment variable to use. For example, marathon sets `MARATHON_APP_ID` and
by setting `COLLECTD_DOCKER_APP_ENV` to `MARATHON_APP_ID` you would get the
marathon app id.

These keys can be changed globally by setting `APP_LABEL_KEY` or `APP_ENV_KEY`
when running the collectd container. For example, if you set `APP_ENV_KEY` to
`MARATHON_APP_ID` on the collectd container, then this will use
`MARATHON_APP_ID` on all running containers.

### Setting the Task name of a Container

* Set the label `collectd_docker_task` directly on the container
* Set `collectd_docker_task_label` on the container that points to which actual
label to use. e.g.`collectd_docker_task_label=task_id` will use `task_id` label on the
container
* Set environment variable `COLLECTD_DOCKER_TASK` on the container
* Set `COLLECTD_DOCKER_TASK_ENV` on the container that points to which actual
environment variable to use. For example, mesos sets `MESOS_TASK_ID` and by
setting `COLLECTD_DOCKER_TASK_ENV` to `MESOS_TASK_ID` you would get the mesos
task id.

These keys can be changed globally by setting `TASK_LABEL_KEY` or `TASK_ENV_KEY`
when running the collectd container. For example, if you set `TASK_ENV_KEY` to
`MESOS_TASK_ID` on the collectd container, then this will use `MESOS_TASK_ID` on
all running containers.

### Limitations

* If a container's app name cannot be identified, it will be not monitored. So
if you are not seeing metrics, then it means you must check whether the app
name is configured correctly.

## Reported metrics

Influxdb table is created with a table per measurement type, and a column per gauge

Gauges:

* CPU (cpu table)
    * `user`
    * `system`
    * `total`

* Memory overview (memory table)
    * `limit`
    * `max`
    * `usage`

* Memory breakdown  (memory table)
    * `active_anon`
    * `active_file`
    * `cache`
    * `inactive_anon`
    * `inactive_file`
    * `mapped_file`
    * `pg_fault`
    * `pg_in`
    * `pg_out`
    * `rss`
    * `rss_huge`
    * `unevictable`
    * `writeback`

* Network (bridge mode only)
    * `rx_bytes`
    * `rx_dropped`
    * `rx_errors`
    * `rx_packets`
    * `tx_bytes`
    * `tx_dropped`
    * `tx_errors`
    * `tx_packets`

## Grafana dashboard

Grafana 2 [dashboard](grafana2.json) is included.

![screenshot](https://github.com/forestscribe/collectd-docker/raw/master/screenshot.png)


## Running

Minimal command:

```
docker run -d -v /var/run/docker.sock:/var/run/docker.sock \
    -e INFLUXDB_URL=<influxdb url> \
    forestscribe/collectd-docker
```

### Environment variables

* `COLLECT_INTERVAL` - metric update interval in seconds, defaults to `10`.
* `INFLUXDB_URL` - influxdb database where carbon is listening for data.
* `INFLUXDB_DATABASE` - influxdb database (will be created if inexistent).
* `INFLUXDB_USERNAME` - influxdb username for that database.
* `INFLUXDB_PASSWORD` - influxdb password for that database.
* `APP_LABEL_KEY` - container label to use for app name, `collectd_docker_app` by default.
* `APP_ENV_KEY` - container environment variable to use for app name, `MARATHON_APP_ID` by default.
* `TASK_LABEL_KEY` - container label to use for task name, `collectd_docker_task` by default.
* `TASK_ENV_KEY` - container environment variable to use for task name, `MESOS_TASK_ID` by default.

Note that this docker image is very minimal and libc inside does not support
`search` directive in `/etc/resolv.conf`. You have to supply full hostname in
`GRAPHITE_HOST` that can be resolved with nameserver.

## License

MIT
