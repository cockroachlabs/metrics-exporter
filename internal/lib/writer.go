// Copyright 2021 The Cockroach Authors.
//
// Use of this software is governed
// by the Apache License, Version 2.0, included in the file
// LICENSE.md

package lib

import (
	"context"
	"net/http"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

// A MetricsWriter write metrics, after transforming them based on the configuration supplied.
type MetricsWriter struct {
	Config *Config
}

// Create a MetricsWriter
func CreateMetricsWriter(config *Config) *MetricsWriter {
	return &MetricsWriter{
		Config: config,
	}
}

// Write the metrics, converting HDR Histogram into Log10 linear histograms.
func (w *MetricsWriter) WriteMetrics(
	ctx context.Context,
	metricFamilies map[string]*dto.MetricFamily,
	h http.ResponseWriter) {
	for _, mf := range metricFamilies {
		if mf.GetType() == dto.MetricType_HISTOGRAM {
			TranslateHistogram(&w.Config.Bucket, mf)
		}
		expfmt.MetricFamilyToText(h, mf)
	}
}
