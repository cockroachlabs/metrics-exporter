// Copyright 2021 The Cockroach Authors.
//
// Use of this software is governed
// by the Apache License, Version 2.0, included in the file
// LICENSE.md

package lib

import (
	"context"
	"io"
	"regexp"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	log "github.com/sirupsen/logrus"
)

// MetricsWriter write metrics, after transforming them based on the configuration supplied.
type MetricsWriter struct {
	Config  *Config
	Exclude *regexp.Regexp
	Include *regexp.Regexp
}

// CreateMetricsWriter instantiates a MetricsWriter
func CreateMetricsWriter(config *Config) *MetricsWriter {
	var exc, inc *regexp.Regexp
	if config.Bucket.Exclude != "" {
		exc = regexp.MustCompile(config.Bucket.Exclude)
	}
	if config.Bucket.Include != "" {
		inc = regexp.MustCompile(config.Bucket.Include)
	}
	return &MetricsWriter{
		Config:  config,
		Exclude: exc,
		Include: inc,
	}
}

// WriteMetrics writes the metrics, converting HDR Histogram into Log10 linear histograms.
func (w *MetricsWriter) WriteMetrics(
	ctx context.Context, metricFamilies map[string]*dto.MetricFamily, out io.Writer,
) {
	for _, mf := range metricFamilies {
		if mf.GetType() == dto.MetricType_HISTOGRAM {
			if w.Include != nil && w.Include.MatchString(mf.GetName()) {
				// Processing this even it matches the exclude.
			} else if w.Exclude != nil && w.Exclude.MatchString(mf.GetName()) {
				log.Tracef("Skipping %s", mf.GetName())
				// Skipping this
				continue
			}
			log.Tracef("Translating %s", mf.GetName())
			TranslateHistogram(&w.Config.Bucket, mf)
		}
		expfmt.MetricFamilyToText(out, mf)
	}
}
