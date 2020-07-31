# Introduction
`infping` is a simple Go program to parse the output of [`fping`](https://github.com/schweikert/fping) and store it in [InfluxDB](https://www.influxdata.com/time-series-platform/influxdb/). Based upon [software](https://hveem.no/visualizing-latency-variance-with-grafana) by [Tor Hveem](https://hveem.no/).

# Configuration
This program uses the [Viper configuration package](https://github.com/spf13/viper) and can process configuration files in [JSON](https://json.org/), [YAML](http://yaml.org/), [TOML](https://github.com/toml-lang/toml), and others. Save your configuration file as `infping.<json|yaml|toml|...>` in `/etc/`, `/usr/local/etc`, or the program directory. A sample configuration file is provided in JSON format.

### influx
* **host**: The hostname to connect to
* **port**: The TCP port number
* **user**: The username, if needed
* **pass**: The password, if needed
* **db**: The database name to connect to – this will be created if it does not exist
* **secure**: Set to true to enable HTTPS

### fping
* **backoff**: The value for the `-B` argument
* **retries**: The value for the `-r` argument
* **tos**: The value for the `-O` argument
* **summary**: The value for the `-Q` argument – this determines how often data is collected
* **period**: The value for the `-p` argument
* **dualstack**: If true: implies `-m -n -A` (send to all addresses, print both DNS name and IP address)

### hosts
* **hosts**: An array of hostnames to ping

# Influx Storage
Data is stored in Influx with the following fields and tags:
* **min**: *field* showing minimum ping time during the run
* **max**: *field* showing maximum ping time during the run
* **avg**: *field* showing average ping time during the run
* **loss**: *field* showing packet loss during the run
* **rx_host**: *tag* showing the target host
* **tx_host**: *tag* showing the originating host of the ping check

# Grafana Dashboard
A sample Grafana dashboard is included, that plots all four of the collected ping statistics in something approximating the display of [Smokeping](https://smokeping.org/). Simply create a datasource named "infping" pointing to Influx, and then import this dashboard. The `hostname` variable will be automatically populated with all the host names found in the database, and can be used to select different graphs.

![dashboard screenshot](https://raw.githubusercontent.com/mrngm/infping/master/grafana_dashboard.png)
