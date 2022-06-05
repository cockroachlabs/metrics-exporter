// Copyright 2021 The Cockroach Authors.
//
// Use of this software is governed
// by the Apache License, Version 2.0, included in the file
// LICENSE.md

package lib

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	log "github.com/sirupsen/logrus"
)

const defaultLimit = 50

type Database struct {
	config       *pgxpool.Config
	pool         *pgxpool.Pool
	rowArrayLast RowLioSample
}

type RowLioSample struct {
	aggEpochSecs int
	lioTotal     int
	fullLio      int
	iJoinLio     int
	explicitLio  int
	healthyLio   int
}

func InitDb(ctx context.Context, url string) (*Database, error) {
	config, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, err
	}
	var pool *pgxpool.Pool
	for i := 0; i < 10; i++ {
		pool, err = pgxpool.ConnectConfig(ctx, config)
		if err != nil {
			log.Warn("Unable to connect to the db. Retrying")
			time.Sleep(time.Second * 5)
		} else {
			break
		}
	}
	if err != nil {
		return nil, err
	}
	return &Database{
		config: config,
		pool:   pool,
	}, nil
}

func (db *Database) GetConnection(ctx context.Context) (*pgxpool.Conn, error) {
	return db.pool.Acquire(ctx)
}

const isNodeQuery = `select max(node_id) = crdb_internal.node_id() from crdb_internal.gossip_nodes;`
const statementQuery = `SELECT
    metadata->'query'  
   from crdb_internal.statement_statistics 
   where fingerprint_id = $1 limit 1;`

//go:embed stmtHostActivity.sql
var stmtActivity string

//go:embed stmtEfficiency.sql
var stmtEfficiency string

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

var (
	labelNames = []string{"statementid", "app", "database"}
	cnt        = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "crdb_custom_cnt",
			Help: "SQL Activity: count",
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
		}, []string{"lio"})
)

func (db *Database) InitMetrics() {
	prometheus.MustRegister(cnt)
	prometheus.MustRegister(maxDiskUsage)
	prometheus.MustRegister(networkBytes)
	prometheus.MustRegister(runLat)
	prometheus.MustRegister(rowsRead)
	prometheus.MustRegister(numRows)
	prometheus.MustRegister(bytesRead)
	prometheus.MustRegister(maxRetries)
	prometheus.MustRegister(maxMem)
	prometheus.MustRegister(contTime)
	prometheus.MustRegister(stmtStats)

}

func (db *Database) GetCustomMetrics(ctx context.Context, limit int) error {
	mainNode, err := db.IsMainNode(ctx)
	if err != nil {
		log.Error(err)
		return err
	}
	log.Tracef("Main node: %t ", mainNode)
	if mainNode {
		err = db.GetEfficiency(ctx)
		if err != nil {
			log.Error(err)
			return err
		}
	}
	if limit == 0 {
		limit = defaultLimit
	}
	err = db.GetActivity(ctx, limit)
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func (db *Database) IsMainNode(ctx context.Context) (bool, error) {
	tx, err := db.GetConnection(ctx)
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
func (db *Database) GetActivity(ctx context.Context, limit int) error {
	tx, err := db.GetConnection(ctx)
	if err != nil {
		return err
	}
	defer tx.Release()
	rows, err := tx.Query(ctx, stmtActivity, limit)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		r := &activity{}
		err := rows.Scan(&r.time, &r.id, &r.app, &r.database, &r.cnt, &r.maxDiskUsage,
			&r.networkBytes, &r.runLat, &r.rowsRead, &r.numRows, &r.bytesRead,
			&r.maxRetries, &r.maxMem, &r.contTime)
		if err != nil {
			log.Error(err)
			continue
		}
		id := r.id
		labels := []string{id, r.app, r.database}
		cnt.WithLabelValues(labels...).Set(float64(r.cnt))
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
	return nil
}

func (db *Database) GetStatement(ctx context.Context, id string) (string, error) {
	s, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return "", err
	}
	tx, err := db.GetConnection(ctx)
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

func noNegVals(a int, b int) float64 {
	if a > b {
		return float64(a - b)
	} else {
		return float64(0)
	}
}

func (db *Database) GetEfficiency(ctx context.Context) error {

	// Run SQL to extract statement statistics and normalize to LIO
	// These values are returned as a data structure of Rows which
	// is then operated on by various statements to show potential inefficiencies
	//var rowArray Row
	rowArray := RowLioSample{}

	// Sample Last Hour or all History
	rows, err := db.pool.Query(ctx, stmtEfficiency)

	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&rowArray.aggEpochSecs, &rowArray.lioTotal, &rowArray.fullLio, &rowArray.iJoinLio, &rowArray.explicitLio, &rowArray.healthyLio)
		if err != nil {
			return err
		}

		if db.rowArrayLast.aggEpochSecs != rowArray.aggEpochSecs {
			stmtStats.Reset()
		} else {
			stmtStats.WithLabelValues("full").Add(noNegVals(rowArray.fullLio, db.rowArrayLast.fullLio))
			stmtStats.WithLabelValues("ijoin").Add(noNegVals(rowArray.iJoinLio, db.rowArrayLast.iJoinLio))
			stmtStats.WithLabelValues("explicit").Add(noNegVals(rowArray.explicitLio, db.rowArrayLast.explicitLio))
			stmtStats.WithLabelValues("optimized").Add(noNegVals(rowArray.healthyLio, db.rowArrayLast.healthyLio))
		}
		db.rowArrayLast.aggEpochSecs = rowArray.aggEpochSecs
		db.rowArrayLast.lioTotal = rowArray.lioTotal
		db.rowArrayLast.fullLio = rowArray.fullLio
		db.rowArrayLast.iJoinLio = rowArray.iJoinLio
		db.rowArrayLast.explicitLio = rowArray.explicitLio
		db.rowArrayLast.healthyLio = rowArray.healthyLio
	}

	return nil
}
