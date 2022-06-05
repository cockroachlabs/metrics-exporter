WITH
	tmp
		AS (
			SELECT
				extract('epoch', now())::INT8 AS ts,
				statement_id,
				application_name,
				database_name,
				sum(count) AS c,
				max(max_disk_usage_avg) AS max_disk,
				sum(network_bytes_avg * count::FLOAT8) / sum(count)::FLOAT8
					AS net_bytes,
				sum(run_lat_avg * count::FLOAT8) / sum(count)::FLOAT8
					AS run_lat,
				sum(rows_read_avg * count::FLOAT8) / sum(count)::FLOAT8
					AS rows_read,
				sum(rows_avg * count::FLOAT8) / sum(count)::FLOAT8 
          AS rows_avg,
				sum(bytes_read_avg * count::FLOAT8) / sum(count)::FLOAT8
					AS bytes_avg,
				sum(run_lat_avg * count::FLOAT8) AS total_lat,
				sum(max_retries) AS max_retries,
				max(max_mem_usage_avg) AS max_mem,
				sum(contention_time_avg * count::FLOAT8) / sum(count)::FLOAT8
					AS cont_time
			FROM
				crdb_internal.node_statement_statistics
			WHERE
				application_name NOT LIKE '$ internal-%'
			GROUP BY
				statement_id, application_name, database_name
		)
SELECT
	ts,
	statement_id,
	application_name,
	database_name,
	c,
	max_disk,
	net_bytes,
	run_lat,
	rows_read,
	rows_avg,
	bytes_avg,
	max_retries,
	max_mem,
	cont_time
FROM
	tmp
ORDER BY
	total_lat DESC
LIMIT
	$1;
