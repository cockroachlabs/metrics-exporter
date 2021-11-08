# metrics-exporter

Proxy to filter and rewrite cockroachDB metrics in Prometheus format.
Currently, it exports all metrics as is, except for histograms.
Histograms are converted from a log-2 linear format (HDR histograms) to a log-10 linear format. 
They have consistent lower/upper buckets to work with Grafana's heatmaps.

Usage: 
```text
$ metrics-exporter -config config.yaml
```
The configuration, specified in yaml format specifies the cockroach db URL the proxy connects to and the port the proxy it listens to.
The log-10 linear format precision is configurable, specifying the lower range (in nanoseconds) and the number of linear bins for each logarithmic bin. 
Optionally, the user can specify the upper range (in nanoseconds), the unit (seconds,milliseconds,microseconds) to convert the bucket ranges, and a regex expression to include/exclude matching histograms (all the buckets matching the include regex will be included, even if the match the exclude regex).
The tls section allows the user to specify CA, cert and private key to connect to the backend. The same configuration is used to configure the HTTPS endpoint that the proxy listen to.

### Sample configuration:

```text
url: https://localhost:8080/_status/vars
port: 8888
bucket:
  startns: 100000 
  bins: 10 
  endns: 20000000000
  unit: millseconds 
  exclude: (.*internal)
  include: (sql_exec_latency_internal_bucket)
tls:
  ca: ./certs/ca.crt
  privatekey: ./certs/client.root.key
  certificate: ./certs/client.root.crt
```

