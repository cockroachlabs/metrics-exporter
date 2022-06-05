module github.com/cockroachlabs/metrics-exporter

go 1.16

replace internal/lib => ./internal/lib

require (
	github.com/NYTimes/gziphandler v1.1.1
	github.com/cockroachdb/crlfmt v0.0.0-20210128092314-b3eff0b87c79
	github.com/jackc/pgx/v4 v4.16.1 // indirect
	github.com/prometheus/client_golang v1.12.2
	github.com/prometheus/common v0.34.0
	github.com/sirupsen/logrus v1.8.1 // indirect
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b
	golang.org/x/tools v0.0.0-20200923014426-f5e916c686e1
	honnef.co/go/tools v0.0.1-2020.1.4
	internal/lib v1.0.0
)
