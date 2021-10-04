# metrics-exporter

Proxy to filter and rewrite cockroachDB metrics in Prometheus format.
Currently, it exports all metrics as is, except for histograms.
Histograms are converted from a log-2 linear format (HDR histograms) to a log-10 linear format.

Usage: 
```text
$ metrics-exporter -config config.yaml
```
The configuration, specified in yaml format specifies the cockroach db URL the proxy connects to and the port the proxy it listens to.
The log-10 linear format precision is configurable, specifing the lower range (in nanoseconds) and the number of linear bins for each logarithmic bin. 
The tls section allows the user to specify CA, cert and private key to connect to the backend. The same configuration is used to configure the HTTPS endpoint that the proxy listen to.

```text
url: https://localhost:8080/_status/vars
port: 8888
bucket:
  startns: 100000 
  bins: 10 
tls:
  ca: ./certs/ca.crt
  privatekey: ./certs/client.root.key
  certificate: ./certs/client.root.crt
```

