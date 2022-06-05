// Copyright 2021 The Cockroach Authors.
//
// Use of this software is governed
// by the Apache License, Version 2.0, included in the file
// LICENSE.md

// Utilties functions to read/write/transform metrics.
package lib

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"log"
	"math"
	"net/url"
	"os"

	"gopkg.in/yaml.v2"
)

/*
Metrics Exporter Configuration

 * Bucket: Log10 Bucket Configuration
 * Port: Port that the export is listening to
 * Tls: optional Tls configuration
 * Url: CockroachDB Prometheus endpoint
*/
type Config struct {
	Bucket BucketConfig
	Port   int
	Tls    TlsConfig `yaml:"tls,omitempty"`
	Url    string
	Custom Custom `yaml:"custom,omitempty"`
}

func (c Config) checkConfig() error {
	_, err := url.ParseRequestURI(c.Url)
	if err != nil {
		return err
	}
	if c.Port < 1024 || c.Port > 65535 {
		return errors.New("Invalid port range")
	}
	return c.Bucket.checkConfig()
}

/*
Log10 Bucket Configuration

 * Bins: the number of linear buckets for each log10 bucket
 * Startns: The lower range in nanoseconds.
 * Endns: Optional upper range
 * Exclude: Regex of histogram names to exclude
 * Include: Regex of histogram names to include, regardless of the exclude settings
 * Unit: Time unit to use for the log10 buckets
 *
*/
type BucketConfig struct {
	Bins    int
	Startns int
	// optional
	Endns   int
	Exclude string
	Include string
	Unit    string
}

func (b *BucketConfig) checkConfig() error {
	if b.Bins >= 1 && b.Bins <= 100 && b.Startns >= 1 {
		return nil
	}
	return errors.New("Invalid Bucket Configuration")
}

func (b *BucketConfig) UnitDiv() float64 {
	var div float64 = 1
	if b.Unit != "" {
		switch b.Unit {
		case "seconds":
			div = math.Pow10(9)
		case "milliseconds":
			div = math.Pow10(6)
		case "microseconds":
			div = math.Pow10(3)
		}
	}
	return div
}

/*
TlsConfig  Configuration
 * Ca: CA certificate file location
 * Certificate: X.509 certificate for the server
 * Host: Host name associated with X.509 certificate.
 * PrivateKey: Server private key

*/
type TlsConfig struct {
	Ca          string
	Certificate string
	Host        string
	PrivateKey  string
}

// Context for TLS connections.
type TlsClientContext struct {
	CertPool    *x509.CertPool
	Certificate tls.Certificate
}

// Url to fetch custom metrics from the cluster.
type Custom struct {
	Url   string
	Limit int
}

// Reads yaml configuration from a file
func ReadConfig(configLocation *string) *Config {
	data, err := os.ReadFile(*configLocation)
	if err != nil {
		log.Fatal("Error reading configuration at", configLocation, ": ", err)
	}
	config := Config{}
	err = yaml.Unmarshal([]byte(data), &config)
	if err != nil {
		log.Fatal("Error reading configuration:", err)
	}

	err = config.checkConfig()
	if err != nil {
		log.Fatal("Error reading configuration:", err)
	}
	return &config
}

// Returns true if there is a Tls configuration
func (config *Config) IsSecure() bool {
	return config.Tls != TlsConfig{}
}

// Returns true if there is a custom section
func (config *Config) HasCustom() bool {
	return config.Custom != Custom{}
}

// Builds the Client TLS context
func (config *Config) GetTlsClientContext() (*TlsClientContext, error) {
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

// Builds the Server TLS context
func (config *Config) GetTlsServerContext() (*tls.Config, error) {
	caCert, err := ioutil.ReadFile(config.Tls.Ca)
	if err != nil {
		return nil, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	return &tls.Config{
		ServerName: config.Tls.Host,
		ClientAuth: tls.NoClientCert,
		ClientCAs:  caCertPool,
		MinVersion: tls.VersionTLS12,
	}, nil
}
