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
	"log"
	"net/http"
	"runtime"
	"runtime/debug"

	"internal/lib"

	"github.com/NYTimes/gziphandler"
)

var (
	// buildVersion is set by the go linker at build time
	buildVersion = "<unknown>"
)

func printVersionInfo(buildVersion string) {
	fmt.Println("metrics-exporter", buildVersion)
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

func main() {
	configLocation := flag.String("config", "", "YAML configuration")
	printVersion := flag.Bool("version", false, "print version and exit")
	localFile := flag.String("local", "", "local file")
	flag.Parse()
	if *printVersion {
		printVersionInfo(buildVersion)
		return
	}

	config := lib.ReadConfig(configLocation)
	var err error = nil
	var secureCtx *lib.TlsClientContext = nil
	var transport = &http.Transport{}

	if config.IsSecure() {
		secureCtx, err = config.GetTlsClientContext()
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

	if *localFile != "" {
		lib.TranslateFromFile(&config.Bucket, *localFile)
		return
	}

	reader := lib.CreateMetricsReader(config, transport)
	writer := lib.CreateMetricsWriter(config)
	ctx := context.Background()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		metricFamilies, err := reader.ReadMetrics(ctx)
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprintln(w, err)
			return
		}
		writer.WriteMetrics(ctx, metricFamilies, w)

	})

	http.Handle("/_status/vars", gziphandler.GzipHandler(handler))

	server := &http.Server{
		Addr: ":" + fmt.Sprintf("%d", config.Port),
	}
	log.Printf("Starting proxy with:\n%+v\n\n", config)

	if !config.IsSecure() {
		err = server.ListenAndServe()

	} else {
		server.TLSConfig, err = config.GetTlsServerContext()
		if err != nil {
			log.Fatal("Error setting up secure server: ", err)
		}
		err = server.ListenAndServeTLS(config.Tls.Certificate, config.Tls.PrivateKey)
	}
	if err != nil {
		log.Fatal("Error starting server: ", err)
	}
}
