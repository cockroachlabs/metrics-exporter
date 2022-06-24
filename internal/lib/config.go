// Copyright 2021 The Cockroach Authors.
//
// Use of this software is governed
// by the Apache License, Version 2.0, included in the file
// LICENSE.md

// Package lib provides utility functions to read/write/transform metrics.
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

// Config has the configuration for the metrics-exporter
// * Bucket: Log10 Bucket Configuration
// * Port: Port that the export is listening to
// * Tls: optional Tls configuration
// * Url: CockroachDB Prometheus endpoint
type Config struct {
	Bucket BucketConfig
	Port   int
	TLS    TLSConfig `yaml:"tls,omitempty"`
	URL    string
	Custom Custom `yaml:"custom,omitempty"`
}

func (c Config) checkConfig() error {
	_, err := url.ParseRequestURI(c.URL)
	if err != nil {
		return err
	}
	if c.Port < 1024 || c.Port > 65535 {
		return errors.New("Invalid port range")
	}
	return c.Bucket.checkConfig()
}

// BucketConfig defines the config parameters for each histogram bucket
// * Bins: the number of linear buckets for each log10 bucket
// * Startns: The lower range in nanoseconds.
// * Endns: Optional upper range
// * Exclude: Regex of histogram names to exclude
// * Include: Regex of histogram names to include, regardless of the exclude settings
// * Unit: Time unit to use for the log10 buckets
type BucketConfig struct {
	Bins    int
	Startns int
	// optional
	Endns   int
	Exclude string
	Include string
	Unit    string
}

// UnitDiv converts time units into nano secods
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

// TLSConfig  Configuration
// * Ca: CA certificate file location
// * Certificate: X.509 certificate for the server
// * Host: Host name associated with X.509 certificate.
// * PrivateKey: Server private key
type TLSConfig struct {
	Ca          string
	Certificate string
	Host        string
	PrivateKey  string
}

// TLSClientContext is the context for TLS connections.
type TLSClientContext struct {
	CertPool    *x509.CertPool
	Certificate tls.Certificate
}

// Custom provides the configuration to retrieve custom metrics
type Custom struct {
	URL                 string
	DisableGetStatement bool
	Limit               int
	SkipActivity        bool
	SkipEfficiency      bool
	Frequency           int
	Endpoint            string
}

// ReadConfig reads yaml configuration from a file
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

// IsSecure returns true if there is a Tls configuration
func (c *Config) IsSecure() bool {
	return c.TLS != TLSConfig{}
}

// HasCustom returns true if there is a custom section
func (c *Config) HasCustom() bool {
	return c.Custom != Custom{}
}

// GetTLSClientContext builds the Client TLS context
func (c *Config) GetTLSClientContext() (*TLSClientContext, error) {
	var cert tls.Certificate
	var err error
	if c.TLS.Certificate != "" && c.TLS.PrivateKey != "" {
		cert, err = tls.LoadX509KeyPair(c.TLS.Certificate, c.TLS.PrivateKey)
		if err != nil {
			return nil, err
		}
	}
	caCert, err := ioutil.ReadFile(c.TLS.Ca)
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	return &TLSClientContext{
		CertPool:    caCertPool,
		Certificate: cert,
	}, err
}

// GetTLSServerContext builds the Server TLS context
func (c *Config) GetTLSServerContext() (*tls.Config, error) {
	caCert, err := ioutil.ReadFile(c.TLS.Ca)
	if err != nil {
		return nil, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	return &tls.Config{
		ServerName: c.TLS.Host,
		ClientAuth: tls.NoClientCert,
		ClientCAs:  caCertPool,
		MinVersion: tls.VersionTLS12,
	}, nil
}

func (b *BucketConfig) checkConfig() error {
	if b.Bins >= 1 && b.Bins <= 100 && b.Startns >= 1 {
		return nil
	}
	return errors.New("Invalid Bucket Configuration")
}
