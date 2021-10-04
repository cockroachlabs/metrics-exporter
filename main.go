// Copyright 2021 The Cockroach Authors.
// Use of this software will be governed
// by the Apache License, Version 2.0, included in the file LICENSE.md.

package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"

	"github.com/NYTimes/gziphandler"
	"gopkg.in/yaml.v2"
)

type BucketConfig struct {
	Startns  int
	Startexp int
	Bins     int
}

type TlsConfig struct {
	Host        string
	Ca          string
	PrivateKey  string
	Certificate string
}

type Config struct {
	Url    string
	Port   int
	Bucket BucketConfig
	Tls    TlsConfig `yaml:"tls,omitempty"`
}

type TlsClientContext struct {
	CertPool    *x509.CertPool
	Certificate tls.Certificate
}

func GetTlsClientContext(config Config) (*TlsClientContext, error) {
	var cert tls.Certificate
	var err error
	if config.Tls.Certificate != "" && config.Tls.PrivateKey != "" {
		cert, err = tls.LoadX509KeyPair(config.Tls.Certificate, config.Tls.PrivateKey)
		if err != nil {
			return nil, err
		}
	}
	caCert, err := ioutil.ReadFile(config.Tls.Ca)
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	return &TlsClientContext{
		CertPool:    caCertPool,
		Certificate: cert,
	}, err
}

func GetTlsServerContext(config Config) *tls.Config {
	caCert, err := ioutil.ReadFile(config.Tls.Ca)
	if err != nil {
		log.Fatal("Error opening cert file", config.Tls.Ca, ", error ", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	return &tls.Config{
		ServerName: config.Tls.Host,
		ClientAuth: tls.NoClientCert,
		ClientCAs:  caCertPool,
		MinVersion: tls.VersionTLS12,
	}
}

func main() {
	configLocation := flag.String("config", "", "YAML configuration")
	flag.Parse()
	data, err := os.ReadFile(*configLocation)

	if err != nil {
		log.Fatal("Error reading configuration at", configLocation, ": ", err)
	}

	config := Config{}

	err = yaml.Unmarshal([]byte(data), &config)

	config.Bucket.Startexp = int(math.Log10(float64(config.Bucket.Startns)))

	if err != nil {
		log.Fatal("Error reading configuration: ", err)
	}
	var ctx *TlsClientContext = nil

	if (config.Tls != TlsConfig{}) {
		ctx, err = GetTlsClientContext(config)
		if err != nil {
			log.Fatal("Error setting up secure context: ", err)
		}
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		GetMetrics(config, ctx, w)
	})

	http.Handle("/_status/vars/", gziphandler.GzipHandler(handler))

	server := &http.Server{
		Addr: ":" + fmt.Sprintf("%d", config.Port),
	}
	fmt.Printf("Starting proxy with:\n%v\n\n", config)

	if (config.Tls == TlsConfig{}) {
		server.ListenAndServe()
	} else {
		server.TLSConfig = GetTlsServerContext(config)
		server.ListenAndServeTLS(config.Tls.Certificate, config.Tls.PrivateKey)
	}

}
