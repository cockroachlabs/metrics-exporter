// Copyright 2021 The Cockroach Authors.
//
// Use of this software is governed
// by the Apache License, Version 2.0, included in the file
// LICENSE.md

// The metrics-exported implements a simple proxy that filters and rewrites
// CockroachDB Prometheus metrics.
// In the current implementation it converts HDRHistograms to log10 linear based histograms.
package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/cockroachlabs/metrics-exporter/internal/lib"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/expfmt"
	log "github.com/sirupsen/logrus"
)

var (
	// buildVersion is set by the go linker at build time
	buildVersion           = "<unknown>"
	buildTimestamp         = "<unknown>"
	defaultCustomFrequency = 10
)

func printVersionInfo(buildVersion string) {
	fmt.Println("metrics-exporter", buildVersion)
	fmt.Println("built on", buildTimestamp)

	fmt.Println(runtime.Version(), runtime.GOARCH, runtime.GOOS)
	fmt.Println()
	if bi, ok := debug.ReadBuildInfo(); ok {
		fmt.Println(bi.Main.Path, bi.Main.Version)
		for _, m := range bi.Deps {
			for m.Replace != nil {
				m = m.Replace
			}
			fmt.Println(m.Path, m.Version)
		}
	}
}

func translateFromFile(
	ctx context.Context, config *lib.BucketConfig, filename string, writer *lib.MetricsWriter,
) {
	log.Info("Reading from file :" + filename)
	var parser expfmt.TextParser
	var r, err = os.Open(filename)
	if err != nil {
		panic("File not found")
	}
	metricFamilies, _ := parser.TextToMetricFamilies(r)
	writer.WriteMetrics(ctx, metricFamilies, os.Stdout)

}
func setVersionMetric() {
	ts, err := strconv.ParseInt(buildTimestamp, 10, 64)
	if err == nil {
		promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "metrics_exporter_build_timestamp",
				Help: "metrics exporter build timestamp",
			},
			[]string{"tag"},
		).WithLabelValues(buildVersion).Set(float64(ts))
	}
}
func main() {
	configLocation := flag.String("config", "", "YAML configuration")
	printVersion := flag.Bool("version", false, "print version and exit")
	debug := flag.Bool("debug", false, "log debug info")
	trace := flag.Bool("trace", false, "log trace info")
	localFile := flag.String("local", "", "use local file to read Prometheus metrics (for testing)")
	flag.Parse()
	if *printVersion {
		printVersionInfo(buildVersion)
		return
	}
	if *debug {
		log.SetLevel(log.DebugLevel)
	}
	if *trace {
		log.SetLevel(log.TraceLevel)
	}
	config := lib.ReadConfig(configLocation)
	var err error = nil
	var secureCtx *lib.TLSClientContext = nil
	var transport = &http.Transport{}

	if config.IsSecure() {
		secureCtx, err = config.GetTLSClientContext()
		if err != nil {
			log.Fatal("Error setting up secure context: ", err)
		}
		transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates: []tls.Certificate{secureCtx.Certificate},
				RootCAs:      secureCtx.CertPool,
			},
		}
	}

	writer := lib.CreateMetricsWriter(config)
	ctx := context.Background()

	if *localFile != "" {
		log.Infof("Reading with:\n%+v\n\n", config)
		translateFromFile(ctx, &config.Bucket, *localFile, writer)
		return
	}

	reader := lib.CreateMetricsReader(config, transport)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		metricFamilies, err := reader.ReadMetrics(ctx)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, err)
			return
		}
		writer.WriteMetrics(ctx, metricFamilies, w)
		if config.HasCustom() && config.Custom.Endpoint == "/_status/vars" {
			customHandler := promhttp.InstrumentMetricHandler(
				prometheus.DefaultRegisterer, promhttp.HandlerFor(prometheus.DefaultGatherer,
					promhttp.HandlerOpts{
						DisableCompression: true,
					}),
			)
			customHandler.ServeHTTP(w, r)
		}
	})
	http.Handle("/_status/vars", gziphandler.GzipHandler(handler))
	if config.HasCustom() {
		freq := config.Custom.Frequency
		if freq == 0 {
			freq = defaultCustomFrequency
		}
		if config.Custom.Endpoint != "/_status/vars" {
			if config.Custom.Endpoint == "" {
				config.Custom.Endpoint = "/_status/custom"
			}
			http.Handle(config.Custom.Endpoint, promhttp.Handler())
		}
		go func() {
			db, err := lib.NewCollector(ctx, config.Custom)
			if err != nil {
				log.Error("error connecting to the database", err)
				return
			}
			go func() {
				for {
					err := db.GetCustomMetrics(ctx)
					if err != nil {
						log.Error(err)
					}
					time.Sleep(time.Duration(freq*int(time.Second)) * time.Nanosecond)
				}
			}()
			if !config.Custom.DisableGetStatement {
				http.Handle("/statement/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					args := strings.Split(r.URL.Path, "/")
					if len(args) == 3 {
						res, err := db.GetStatement(ctx, args[2])
						if err != nil {
							w.WriteHeader(http.StatusInternalServerError)
							fmt.Fprintln(w, err)
							return
						}
						w.WriteHeader(http.StatusOK)
						fmt.Fprintln(w, res)
					} else {
						w.WriteHeader(http.StatusBadRequest)
						fmt.Fprintln(w, len(args))
						return
					}
				}))
			}
		}()
	}

	server := &http.Server{
		Addr: ":" + fmt.Sprintf("%d", config.Port),
	}
	log.Infof("Starting metrics exporter %s on port %d", buildVersion, config.Port)
	setVersionMetric()
	log.Debugf("Bucket config: %+v\n Custom config:%+v\n", config.Bucket, config.Custom)
	if !config.IsSecure() {
		err = server.ListenAndServe()

	} else {
		server.TLSConfig, err = config.GetTLSServerContext()
		if err != nil {
			log.Fatal("Error setting up secure server: ", err)
		}
		err = server.ListenAndServeTLS(config.TLS.Certificate, config.TLS.PrivateKey)
	}
	if err != nil {
		log.Fatal("Error starting server: ", err)
	}
}
