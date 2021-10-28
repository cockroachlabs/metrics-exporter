// Copyright 2021 The Cockroach Authors.
//
// Use of this software is governed
// by the Apache License, Version 2.0, included in the file
// LICENSE.md

package lib

import (
	"context"
	"errors"
	"net/http"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

// A MetricReader reads the metrics from a CockroachDB endpoint (/_status/var)
type MetricsReader struct {
	Config *Config
	//SecureCtx *TlsClientContext
	Transport *http.Transport
}

// Creates a new Reader
func CreateMetricsReader(c *Config, t *http.Transport) *MetricsReader {
	return &MetricsReader{
		Config:    c,
		Transport: t,
	}
}

func (r *MetricsReader) fetch(ctx context.Context) (*http.Response, error) {
	client := http.Client{
		Transport: r.Transport,
	}
	req, err := http.NewRequest(http.MethodGet, r.Config.Url, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	return client.Do(req)
}

// Read the metrics from the endpoint and returns a map of dto.MetricFamily
func (r *MetricsReader) ReadMetrics(ctx context.Context) (map[string]*dto.MetricFamily, error) {
	data, err := r.fetch(ctx)
	if err != nil {
		return nil, err
	}
	defer data.Body.Close()
	if data.StatusCode != http.StatusOK {
		return nil, errors.New(data.Status)
	}
	var parser expfmt.TextParser
	return parser.TextToMetricFamilies(data.Body)
}
