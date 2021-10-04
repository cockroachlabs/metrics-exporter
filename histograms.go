package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"math"
	"net/http"
	"time"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"google.golang.org/protobuf/proto"
)

type Log10Bucket struct {
	Index   int
	Exp     int
	Bin     int
	BinNums int
	Max     float64
}

func (bucket *Log10Bucket) nextBin() {
	bucket.Bin++
	if (bucket.BinUpperBound()) >= math.Pow10(bucket.Exp+1) {
		bucket.Bin = 0
		bucket.Exp++
	}
}

func (bucket *Log10Bucket) BinUpperBound() float64 {
	return math.Pow10(bucket.Exp) + float64(bucket.Bin)*(math.Pow10(bucket.Exp+1)/float64(bucket.BinNums))
}

func GetMetricsFromServer(config Config, tlsCtx *TlsClientContext) (*http.Response, error) {

	if (config.Tls == TlsConfig{}) {
		client := http.Client{Timeout: 5 * time.Second}
		return client.Get(config.Url)
	} else {
		t := &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates: []tls.Certificate{tlsCtx.Certificate},
				RootCAs:      tlsCtx.CertPool,
			},
		}
		client := http.Client{Transport: t, Timeout: 5 * time.Second}
		return client.Get(config.Url)
	}
}

func GetMetrics(config Config, tlsCtx *TlsClientContext, w http.ResponseWriter) {

	data, err := GetMetricsFromServer(config, tlsCtx)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintln(w, err)
		return
	}
	defer data.Body.Close()
	if data.StatusCode != 200 {
		w.WriteHeader(data.StatusCode)
		fmt.Fprintln(w, data.Status)
		return
	}
	start := time.Now()
	var parser expfmt.TextParser
	metricFamilies, err := parser.TextToMetricFamilies(data.Body)

	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintln(w, err)
		return
	}
	for _, mf := range metricFamilies {
		if mf.GetType() == dto.MetricType_HISTOGRAM {
			// Convert HDR histograms to log(10)linear.
			TranslateHistogram(&config.Bucket, mf)
		}
		expfmt.MetricFamilyToText(w, mf)
	}
	elapsed := time.Since(start)

	log.Printf("Timing: %s\n ", elapsed)

}

func TranslateHistogram(config *BucketConfig, mf *dto.MetricFamily) {
	start := config.Startexp
	bins := config.Bins
	var prev *dto.Bucket = nil
	for _, m := range mf.Metric {
		requiredBuckets := 1
		max := 0.0
		if len(m.Histogram.Bucket) >= 2 {
			max = m.Histogram.Bucket[len(m.Histogram.Bucket)-2].GetUpperBound()

			requiredBuckets = requiredBuckets + int(math.Ceil(math.Log10(float64(max))))*bins
		}
		newBuckets := make([]*dto.Bucket, requiredBuckets)
		currLog10Bucket := &Log10Bucket{
			Index:   0,
			Bin:     0,
			Exp:     start,
			Max:     max,
			BinNums: bins}
		for _, curr := range m.GetHistogram().GetBucket() {
			AddLog10Buckets(currLog10Bucket, curr, prev, newBuckets)
			prev = curr
		}
		m.Histogram.Bucket = newBuckets[0:currLog10Bucket.Index]
	}
}

func AddLog10Buckets(
	currLog10Bucket *Log10Bucket,
	currHdrBucket *dto.Bucket,
	prevHdrBucket *dto.Bucket,
	newBuckets []*dto.Bucket) {
	le := currHdrBucket.GetUpperBound()
	count := currHdrBucket.GetCumulativeCount()
	// last bucket has le = +Inf.
	if le == math.Inf(1) {
		newBuckets[currLog10Bucket.Index] = &dto.Bucket{
			UpperBound:      proto.Float64(currLog10Bucket.BinUpperBound()),
			CumulativeCount: proto.Uint64(count)}
		currLog10Bucket.Index++
		return
	}
	// skip over lower buckets
	if prevHdrBucket == nil && currLog10Bucket.BinUpperBound() < le {
		for currLog10Bucket.BinUpperBound() < le && currLog10Bucket.BinUpperBound() <= currLog10Bucket.Max {
			currLog10Bucket.nextBin()
		}
		return
	}
	ple := float64(0)
	pcount := uint64(0)
	if prevHdrBucket != nil {
		ple = prevHdrBucket.GetUpperBound()
		pcount = prevHdrBucket.GetCumulativeCount()
	}
	for currLog10Bucket.BinUpperBound() < le && currLog10Bucket.BinUpperBound() <= currLog10Bucket.Max {
		// Assuming a uniform distribution within each of the original buckets, adjust the count if the new
		// bucket spans across two original buckets
		adj := math.Floor(float64(count-pcount) * (le - currLog10Bucket.BinUpperBound()) / (le - ple))
		res := count - uint64(adj)
		bucket := &dto.Bucket{
			UpperBound:      proto.Float64(currLog10Bucket.BinUpperBound()),
			CumulativeCount: proto.Uint64(res)}

		newBuckets[currLog10Bucket.Index] = bucket
		currLog10Bucket.Index++
		currLog10Bucket.nextBin()
	}
}
