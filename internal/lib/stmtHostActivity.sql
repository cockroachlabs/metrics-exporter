
--  CREATE TABLE crdb_internal.node_statement_statistics (
--       node_id INT8 NOT NULL,
--       application_name STRING NOT NULL,
--       flags STRING NOT NULL,
--       statement_id STRING NOT NULL,
--       key STRING NOT NULL,
--       anonymized STRING NULL,
--       count INT8 NOT NULL,
--       first_attempt_count INT8 NOT NULL,
--       max_retries INT8 NOT NULL,
--       last_error STRING NULL,
--       rows_avg FLOAT8 NOT NULL,
--       rows_var FLOAT8 NOT NULL,
--       parse_lat_avg FLOAT8 NOT NULL,
--       parse_lat_var FLOAT8 NOT NULL,
--       plan_lat_avg FLOAT8 NOT NULL,
--       plan_lat_var FLOAT8 NOT NULL,
--       run_lat_avg FLOAT8 NOT NULL,
--       run_lat_var FLOAT8 NOT NULL,
--       service_lat_avg FLOAT8 NOT NULL,
--       service_lat_var FLOAT8 NOT NULL,
--       overhead_lat_avg FLOAT8 NOT NULL,
--       overhead_lat_var FLOAT8 NOT NULL,
--       bytes_read_avg FLOAT8 NOT NULL,
--       bytes_read_var FLOAT8 NOT NULL,
--       rows_read_avg FLOAT8 NOT NULL,
--       rows_read_var FLOAT8 NOT NULL,
--       network_bytes_avg FLOAT8 NULL,
--       network_bytes_var FLOAT8 NULL,
--       network_msgs_avg FLOAT8 NULL,
--       network_msgs_var FLOAT8 NULL,
--       max_mem_usage_avg FLOAT8 NULL,
--       max_mem_usage_var FLOAT8 NULL,
--       max_disk_usage_avg FLOAT8 NULL,
--       max_disk_usage_var FLOAT8 NULL,
--       contention_time_avg FLOAT8 NULL,
--       contention_time_var FLOAT8 NULL,
--       implicit_txn BOOL NOT NULL,
--       full_scan BOOL NOT NULL,
--       sample_plan JSONB NULL,
--       database_name STRING NOT NULL,
--       exec_node_ids INT8[] NOT NULL
--   )


WITH tmp as (SELECT
    extract(epoch from now())::int64 as ts,
    statement_id,
	application_name,
	database_name,
	sum(count) as c,
	max(max_disk_usage_avg) as max_disk,
	sum(network_bytes_avg * count::float) / sum(count)::float as net_bytes,
    sum(run_lat_avg * count::float) / sum(count)::float as run_lat,
    sum(rows_read_avg * count::float) / sum(count)::float as rows_read,
    sum(rows_avg * count::float) / sum(count)::float as rows_avg,
    sum(bytes_read_avg * count::float) / sum(count)::float as bytes_avg,
    sum(run_lat_avg * count::float) as total_lat,
    sum(max_retries) as max_retries,
    max(max_mem_usage_avg) as max_mem,
    sum(contention_time_avg * count::float) / sum(count)::float as cont_time

FROM
  crdb_internal.node_statement_statistics
WHERE
  application_name not like '$ internal-%'  
GROUP BY
 statement_id,
 application_name,
 database_name)
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
FROM tmp
ORDER BY total_lat desc
LIMIT $1


