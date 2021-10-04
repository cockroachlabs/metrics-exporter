module github.com/cockroachlabs/metrics-exporter

go 1.16

replace internal/lib => ./internal/lib

require (
	github.com/NYTimes/gziphandler v1.1.1
	github.com/prometheus/common v0.31.1 // indirect
	internal/lib v1.0.0
)
