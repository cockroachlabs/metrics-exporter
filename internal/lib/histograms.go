// Copyright 2021 The Cockroach Authors.
//
// Use of this software is governed
// by the Apache License, Version 2.0, included in the file
// LICENSE.md

package lib

import (
	"math"

	dto "github.com/prometheus/client_model/go"
	"google.golang.org/protobuf/proto"
)

type log10Bucket struct {
	BinNums int
	Curr    float64
	Max     float64
	UnitDiv float64
}

func createLog10Bucket(start float64, max float64, bins int, div float64) *log10Bucket {
	return &log10Bucket{
		Curr:    start,
		Max:     max,
		BinNums: bins,
		UnitDiv: div,
	}
}

// Computes the next bin
func (b *log10Bucket) nextBin() {
	c := int(math.Floor(math.Log10(b.Curr)))
	m := math.Pow10(c + 1)
	var n float64
	if b.BinNums < 10 && b.Curr <= math.Pow10(c) {
		n = (m / float64(b.BinNums))
	} else {
		n = b.Curr + (m / float64(b.BinNums))
	}
	if n <= m {
		b.Curr = n
	} else {
		b.Curr = m
	}
}

func (b *log10Bucket) binUpperBound() float64 {
	return b.Curr
}

func (b *log10Bucket) addLog10Buckets(
	currHdrBucket *dto.Bucket, prevHdrBucket *dto.Bucket, newBuckets []*dto.Bucket,
) []*dto.Bucket {
	le := currHdrBucket.GetUpperBound()
	count := currHdrBucket.GetCumulativeCount()
	if le == math.Inf(1) {
		for b.binUpperBound() < b.Max {
			bucket := &dto.Bucket{
				UpperBound:      proto.Float64(b.binUpperBound() / b.UnitDiv),
				CumulativeCount: proto.Uint64(count),
			}
			b.nextBin()
			newBuckets = append(newBuckets, bucket)
		}
		return append(newBuckets, &dto.Bucket{
			UpperBound:      proto.Float64(b.binUpperBound() / b.UnitDiv),
			CumulativeCount: proto.Uint64(count)})

	}
	if prevHdrBucket == nil && b.binUpperBound() < le {
		for b.binUpperBound() < le && b.binUpperBound() <= b.Max {
			b.nextBin()
		}
		return newBuckets
	}
	ple := float64(0)
	pcount := uint64(0)
	if prevHdrBucket != nil {
		ple = prevHdrBucket.GetUpperBound()
		pcount = prevHdrBucket.GetCumulativeCount()
	}
	for b.binUpperBound() < le && b.binUpperBound() <= b.Max {
		// Assuming a uniform distribution within each of the original buckets, adjust the count if the new
		// bucket upper bound falls within the original bucket.
		adj := math.Floor(float64(count-pcount) * (le - b.binUpperBound()) / (le - ple))
		res := count - uint64(adj)
		bucket := &dto.Bucket{
			UpperBound:      proto.Float64(b.binUpperBound() / b.UnitDiv),
			CumulativeCount: proto.Uint64(res),
		}
		b.nextBin()
		newBuckets = append(newBuckets, bucket)
	}
	return newBuckets
}

// TranslateHistogram translates the HDR Histogram into a Log10 linear histogram
func TranslateHistogram(config *BucketConfig, mf *dto.MetricFamily) {
	bins := config.Bins
	for _, m := range mf.Metric {
		var prev *dto.Bucket = nil
		requiredBuckets := 1
		max := 0.0
		if len(m.Histogram.Bucket) >= 2 {
			if config.Endns > 0 {
				max = float64(config.Endns)
			} else {
				for _, b := range m.Histogram.Bucket {
					u := b.GetUpperBound()
					if u != math.Inf(1) && u > max {
						max = u
					}
				}
			}
			requiredBuckets = requiredBuckets + int(math.Ceil(math.Log10(float64(max))))*bins
		}
		newBuckets := make([]*dto.Bucket, 0, requiredBuckets)
		currLog10Bucket := createLog10Bucket(float64(config.Startns), max, bins, config.UnitDiv())
		for _, curr := range m.GetHistogram().GetBucket() {
			newBuckets = currLog10Bucket.addLog10Buckets(curr, prev, newBuckets)
			prev = curr
		}
		m.Histogram.Bucket = newBuckets
	}
}
