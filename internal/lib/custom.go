// Copyright 2021 The Cockroach Authors.
//
// Use of this software is governed
// by the Apache License, Version 2.0, included in the file
// LICENSE.md

package lib

import (
	"context"
	// to embed sql files
	_ "embed"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang/groupcache/lru"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	log "github.com/sirupsen/logrus"
)

const defaultLimit = 50

const isNodeQuery = `
SELECT
  max(node_id) = crdb_internal.node_id()
FROM
  crdb_internal.gossip_nodes;`
const statementQuery = `
SELECT
  metadata->'query'
FROM
  crdb_internal.statement_statistics
WHERE
  fingerprint_id = $1 limit 1;`

//go:embed stmtHostActivity.sql
var stmtActivity string

//go:embed stmtEfficiency.sql
var stmtEfficiency string

var (
	labelNames = []string{"statementid", "app", "database"}

	requests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "crdb_custom_cnt",
			Help: "SQL Activity: request count",
		},
		labelNames,
	)
	maxDiskUsage = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "crdb_custom_maxDiskUsage",
			Help: "SQL Activity: maxDiskUsage",
		},
		labelNames,
	)
	networkBytes = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "crdb_custom_networkBytes",
			Help: "SQL Activity: networkBytes",
		},
		labelNames,
	)
	runLat = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "crdb_custom_runLat",
			Help: "SQL Activity: service latency",
		},
		labelNames,
	)
	rowsRead = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "crdb_custom_rowsRead",
			Help: "SQL Activity: rowsRead",
		},
		labelNames,
	)
	numRows = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "crdb_custom_numRows",
			Help: "SQL Activity: numRows",
		},
		labelNames,
	)
	bytesRead = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "crdb_custom_bytesRead",
			Help: "SQL Activity: bytesRead",
		},
		labelNames,
	)
	maxRetries = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "crdb_custom_maxRetries",
			Help: "SQL Activity: maxRetries",
		},
		labelNames,
	)
	maxMem = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "crdb_custom_maxMem",
			Help: "SQL Activity: maxMem",
		},
		labelNames,
	)
	contTime = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "crdb_custom_contTime",
			Help: "SQL Activity: contTime",
		},
		labelNames,
	)
	stmtStats = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "crdb_custom_efficiency",
			Help: "SQL Activity: overall query efficiency",
		}, []string{"type"})
)

type activity struct {
	time         int64
	id           string
	app          string
	database     string
	bytesRead    *float64
	cnt          int64
	contTime     *float64
	maxDiskUsage *float64
	maxMem       *float64
	maxRetries   *float64
	networkBytes *float64
	numRows      *float64
	runLat       *float64
	rowsRead     *float64
}

// Collector queries the database to collect custom metrics
type Collector struct {
	first        bool
	config       Custom
	pool         *pgxpool.Pool
	rowArrayLast rowLioSample
	countCache   *lru.Cache
}
type rowLioSample struct {
	lioTotal    int
	fullLio     int
	iJoinLio    int
	explicitLio int
	healthyLio  int
}

// NewCollector creates a new collector for retrieving sql activity from the
// internal CRDB tables
func NewCollector(ctx context.Context, config Custom) (*Collector, error) {
	poolConfig, err := pgxpool.ParseConfig(config.URL)
	if err != nil {
		return nil, err
	}
	var pool *pgxpool.Pool
	sleep := 5
	for {
		pool, err = pgxpool.ConnectConfig(ctx, poolConfig)
		if err != nil {
			log.Warnf("Unable to connect to the db. Retrying in %d seconds", sleep)
			time.Sleep(time.Duration(sleep * int(time.Second)))
		} else {
			break
		}
		if sleep < 60 {
			sleep += 5
		}
	}
	if config.Limit == 0 {
		config.Limit = defaultLimit
	}
	return &Collector{
		first:      true,
		config:     config,
		pool:       pool,
		countCache: lru.New(config.Limit * 4),
	}, nil
}

// GetCustomMetrics retrieves all the custom metrics
func (c *Collector) GetCustomMetrics(ctx context.Context) error {
	var err error

	err = c.getEfficiency(ctx)
	if err != nil {
		log.Errorf("getEfficiency %s", err.Error())
		return err
	}

	err = c.getActivity(ctx)
	if err != nil {
		log.Errorf("getActivity %s", err.Error())
		return err
	}
	c.first = false
	return nil
}

// GetStatement retrieves the statement associated to the given in
func (c *Collector) GetStatement(ctx context.Context, id string) (string, error) {
	if c.config.DisableGetStatement {
		return "", nil
	}
	s, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return "", err
	}
	tx, err := c.getConnection(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Release()
	rows, err := tx.Query(ctx, statementQuery, fmt.Sprintf("\\x%x", s))
	if err != nil {
		return "", err
	}
	defer rows.Close()
	if rows.Next() {
		var res string
		err := rows.Scan(&res)
		if err != nil {
			return "", err
		}
		return res, nil
	}
	return "", errors.New("Not found")
}

// IsMainNode returns true if the node id of the node is the max(id) in the cluster.
func (c *Collector) IsMainNode(ctx context.Context) (bool, error) {
	tx, err := c.getConnection(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Release()
	var res bool
	row := tx.QueryRow(ctx, isNodeQuery)
	err = row.Scan(&res)
	if err != nil {
		return false, err
	}
	return res, nil
}

func (c *Collector) getActivity(ctx context.Context) error {
	if c.config.SkipActivity {
		return nil
	}
	tx, err := c.getConnection(ctx)
	if err != nil {
		return err
	}
	defer tx.Release()
	rows, err := tx.Query(ctx, stmtActivity, c.config.Limit)
	if err != nil {
		return err
	}
	defer rows.Close()
	resetGauges()
	for rows.Next() {
		r := &activity{}
		err := rows.Scan(&r.time, &r.id, &r.app, &r.database, &r.cnt, &r.maxDiskUsage,
			&r.networkBytes, &r.runLat, &r.rowsRead, &r.numRows, &r.bytesRead,
			&r.maxRetries, &r.maxMem, &r.contTime)
		if err != nil {
			log.Warnf("getActivity %s", err.Error())
			continue
		}

		labels := []string{r.id, r.app, r.database}
		key := r.id + "|" + r.app + "|" + r.database
		if cached, ok := c.countCache.Get(key); ok {
			lastCount := cached.(int64)
			if r.cnt >= lastCount {
				requests.WithLabelValues(labels...).Add(float64(r.cnt - lastCount))
			} else {
				requests.WithLabelValues(labels...).Add(float64(r.cnt))
			}
		} else if !c.first {
			requests.WithLabelValues(labels...).Add(float64(r.cnt))
		}
		c.countCache.Add(key, r.cnt)
		if r.maxDiskUsage != nil {
			maxDiskUsage.WithLabelValues(labels...).Set(*r.maxDiskUsage)
		}
		if r.networkBytes != nil {
			networkBytes.WithLabelValues(labels...).Set(*r.networkBytes)
		}
		if r.runLat != nil {
			runLat.WithLabelValues(labels...).Set(*r.runLat)
		}
		if r.rowsRead != nil {
			rowsRead.WithLabelValues(labels...).Set(*r.rowsRead)
		}
		if r.numRows != nil {
			numRows.WithLabelValues(labels...).Set(*r.numRows)
		}
		if r.bytesRead != nil {
			bytesRead.WithLabelValues(labels...).Set(*r.bytesRead)
		}
		if r.maxRetries != nil {
			maxRetries.WithLabelValues(labels...).Set(*r.maxRetries)
		}
		if r.maxMem != nil {
			maxMem.WithLabelValues(labels...).Set(*r.maxMem)
		}
		if r.contTime != nil {
			contTime.WithLabelValues(labels...).Set(*r.contTime)
		}
	}
	log.Tracef("Cache len:%d", c.countCache.Len())
	return nil
}

func (c *Collector) getConnection(ctx context.Context) (*pgxpool.Conn, error) {
	return c.pool.Acquire(ctx)
}

func (c *Collector) getEfficiency(ctx context.Context) error {
	if c.config.SkipEfficiency {
		return nil
	}
	tx, err := c.getConnection(ctx)
	if err != nil {
		log.Tracef("getEfficiency getConnection %s", err.Error())
		return err
	}
	defer tx.Release()
	rows, err := tx.Query(ctx, stmtEfficiency)
	if err != nil {
		log.Tracef("getEfficiency Query%s", err.Error())
		return err
	}
	defer rows.Close()

	for rows.Next() {
		rowArray := rowLioSample{}
		err := rows.Scan(&rowArray.lioTotal, &rowArray.fullLio, &rowArray.iJoinLio, &rowArray.explicitLio, &rowArray.healthyLio)
		if err != nil {
			log.Tracef("getEfficiency Scan %s", err.Error())
			return err
		}
		log.Debugf("time:%d, explicitTotal:%d, adding: %f",
			time.Now().Unix(),
			rowArray.explicitLio,
			noNegVals(rowArray.explicitLio, c.rowArrayLast.explicitLio))
		if !c.first {
			stmtStats.WithLabelValues("full").Add(noNegVals(rowArray.fullLio, c.rowArrayLast.fullLio))
			stmtStats.WithLabelValues("ijoin").Add(noNegVals(rowArray.iJoinLio, c.rowArrayLast.iJoinLio))
			stmtStats.WithLabelValues("explicit").Add(noNegVals(rowArray.explicitLio, c.rowArrayLast.explicitLio))
			stmtStats.WithLabelValues("optimized").Add(noNegVals(rowArray.healthyLio, c.rowArrayLast.healthyLio))
		}
		c.rowArrayLast.lioTotal = rowArray.lioTotal
		c.rowArrayLast.fullLio = rowArray.fullLio
		c.rowArrayLast.iJoinLio = rowArray.iJoinLio
		c.rowArrayLast.explicitLio = rowArray.explicitLio
		c.rowArrayLast.healthyLio = rowArray.healthyLio
	}
	return nil
}

func noNegVals(a int, b int) float64 {
	if a >= b {
		return float64(a - b)
	}
	return float64(a)
}

func resetGauges() {
	maxDiskUsage.Reset()
	networkBytes.Reset()
	runLat.Reset()
	rowsRead.Reset()
	numRows.Reset()
	bytesRead.Reset()
	maxRetries.Reset()
	maxMem.Reset()
	contTime.Reset()
}
