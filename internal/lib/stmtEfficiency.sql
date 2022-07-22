WITH
	stmt_hr_calc
		AS (
			SELECT
				statement_id,
				IF(implicit_txn = 'false', 1, 0) AS explicittxn,
				IF(full_scan, 1, 0) AS fullscan,
				IF(sample_plan::STRING LIKE '%index join%', 1, 0) AS ijoinstmt,
				rows_avg::INT8 AS numrows,
				rows_read_avg::INT8 AS rowsread,
				greatest(rows_avg::INT8, rows_read_avg::INT8) AS rowsmean,
				count AS execcnt
			FROM
				crdb_internal.node_statement_statistics
			WHERE
				application_name NOT LIKE '$ internal-%'
		),
	sql_distinct_cnt
		AS (
			SELECT
				statement_id,
				sum(fullscan * execcnt) AS fullcnt,
				sum(ijoinstmt * execcnt) AS ijoincnt,
				sum(explicittxn * execcnt) AS explicitcnt,
				sum(
					(
						IF(
							fullscan = 0
							AND ijoinstmt = 0
							AND explicittxn = 0,
							1,
							0
						)
					)
					* execcnt
				)
					AS healthycnt,
				sum(execcnt) AS exectotal,
				sum(rowsmean * execcnt) AS lioperstmt
			FROM
				stmt_hr_calc
			GROUP BY
				statement_id
			ORDER BY
				lioperstmt
		),
stmt_eff as (
	SELECT
	sum(lioperstmt) AS liototal,
	sum(lioperstmt * (IF(fullcnt > 0, 1, 0))) AS fulllio,
	sum(lioperstmt * (IF(ijoincnt > 0, 1, 0))) AS ijoinlio,
	sum(lioperstmt * (IF(explicitcnt > 0, 1, 0))) AS explicitlio,
	sum(lioperstmt * (IF(healthycnt > 0, 1, 0))) AS healtylio
FROM
	sql_distinct_cnt )

SELECT *  FROM stmt_eff WHERE liototal IS NOT NULL
;
