// Copyright 2021 The Cockroach Authors.
//
// Use of this software is governed
// by the Apache License, Version 2.0, included in the file
// LICENSE.md

package lib

import (
	"bytes"
	_ "embed"
	"strings"
	"testing"

	"github.com/prometheus/common/expfmt"
	"github.com/stretchr/testify/assert"
)

//go:embed testdata/input.txt
var input string

//go:embed testdata/output.txt
var output string

//go:embed testdata/multistore.txt
var multistore string

//go:embed testdata/multistoreout.txt
var multistoreout string

func TestHistogramConversion(t *testing.T) {
	assert := assert.New(t)
	var parser expfmt.TextParser
	config := &BucketConfig{
		Startns: 100,
		Bins:    10}

	metricFamilies, _ := parser.TextToMetricFamilies(strings.NewReader(input))

	for _, mf := range metricFamilies {
		TranslateHistogram(config, mf)
		var buf bytes.Buffer
		expfmt.MetricFamilyToText(&buf, mf)

		assert.Equal(buf.String(), output)
	}
}

func TestMultiStoreConversion(t *testing.T) {
	assert := assert.New(t)
	var parser expfmt.TextParser
	config := &BucketConfig{
		Startns: 1000,
		Bins:    10}

	metricFamilies, _ := parser.TextToMetricFamilies(strings.NewReader(multistore))
	for _, mf := range metricFamilies {
		TranslateHistogram(config, mf)
		var buf bytes.Buffer
		expfmt.MetricFamilyToText(&buf, mf)
		assert.Equal(buf.String(), multistoreout)
	}

}

func TestIdentityConversion(t *testing.T) {
	assert := assert.New(t)
	var parser expfmt.TextParser
	config := &BucketConfig{
		Startns: 100,
		Bins:    10}

	metricFamilies, _ := parser.TextToMetricFamilies(strings.NewReader(output))
	for _, mf := range metricFamilies {
		TranslateHistogram(config, mf)
		var buf bytes.Buffer
		expfmt.MetricFamilyToText(&buf, mf)
		assert.Equal(buf.String(), output)
	}

}
