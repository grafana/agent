// Copyright  Grafana Labs
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package collector

const (
	// https://docs.snowflake.com/en/sql-reference/account-usage/storage_usage.html
	storageMetricQuery = `SELECT STORAGE_BYTES, STAGE_BYTES, FAILSAFE_BYTES 
	FROM ACCOUNT_USAGE.STORAGE_USAGE 
	ORDER BY USAGE_DATE DESC LIMIT 1;`

	// https://docs.snowflake.com/en/sql-reference/account-usage/database_storage_usage_history.html
	databaseStorageMetricQuery = `SELECT DATABASE_NAME, DATABASE_ID, AVERAGE_DATABASE_BYTES, AVERAGE_FAILSAFE_BYTES
	FROM ACCOUNT_USAGE.DATABASE_STORAGE_USAGE_HISTORY
	WHERE USAGE_DATE >= dateadd(hour, -24, current_timestamp());`

	// https://docs.snowflake.com/en/sql-reference/account-usage/metering_history.html
	creditMetricQuery = `SELECT SERVICE_TYPE, NAME, avg(CREDITS_USED_COMPUTE), avg(CREDITS_USED_CLOUD_SERVICES)
	FROM ACCOUNT_USAGE.METERING_HISTORY
	WHERE START_TIME >= dateadd(hour, -24, current_timestamp())
	GROUP BY SERVICE_TYPE, NAME;`

	// https://docs.snowflake.com/en/sql-reference/account-usage/warehouse_metering_history.html
	warehouseCreditMetricQuery = `SELECT WAREHOUSE_NAME, WAREHOUSE_ID, avg(CREDITS_USED_COMPUTE), avg(CREDITS_USED_CLOUD_SERVICES)
	FROM ACCOUNT_USAGE.WAREHOUSE_METERING_HISTORY
	WHERE START_TIME >= dateadd(hour, -24, current_timestamp())
	GROUP BY WAREHOUSE_NAME, WAREHOUSE_ID;`

	// https://docs.snowflake.com/en/sql-reference/account-usage/login_history.html
	loginMetricQuery = `SELECT REPORTED_CLIENT_TYPE, REPORTED_CLIENT_VERSION, sum(iff(IS_SUCCESS = 'NO', 1, 0)), 
		sum(iff(IS_SUCCESS = 'YES', 1, 0)), count(*)
	FROM ACCOUNT_USAGE.LOGIN_HISTORY
	WHERE EVENT_TIMESTAMP >= dateadd(hour, -24, current_timestamp())
	GROUP BY REPORTED_CLIENT_TYPE, REPORTED_CLIENT_VERSION;`

	// https://docs.snowflake.com/en/sql-reference/account-usage/warehouse_load_history.html
	warehouseLoadMetricQuery = `SELECT WAREHOUSE_NAME, WAREHOUSE_ID, avg(AVG_RUNNING), avg(AVG_QUEUED_LOAD), avg(AVG_QUEUED_PROVISIONING),  avg(AVG_BLOCKED)
	FROM ACCOUNT_USAGE.WAREHOUSE_LOAD_HISTORY
	WHERE START_TIME >= dateadd(hour, -24, current_timestamp()) 
	GROUP BY WAREHOUSE_NAME, WAREHOUSE_ID;`

	// https://docs.snowflake.com/en/sql-reference/account-usage/automatic_clustering_history.html
	autoClusteringMetricQuery = `SELECT TABLE_NAME, TABLE_ID, SCHEMA_NAME, SCHEMA_ID, DATABASE_NAME, DATABASE_ID, 
		sum(CREDITS_USED), sum(NUM_BYTES_RECLUSTERED), sum(NUM_ROWS_RECLUSTERED)
	FROM ACCOUNT_USAGE.AUTOMATIC_CLUSTERING_HISTORY
	WHERE START_TIME >= dateadd(hour, -24, current_timestamp())
	GROUP BY TABLE_NAME, TABLE_ID, DATABASE_NAME, DATABASE_ID, SCHEMA_NAME, SCHEMA_ID;`

	// https://docs.snowflake.com/en/sql-reference/account-usage/table_storage_metrics.html
	tableStorageMetricQuery = `SELECT TABLE_NAME, ID, TABLE_SCHEMA, TABLE_SCHEMA_ID, TABLE_CATALOG, TABLE_CATALOG_ID, 
		sum(ACTIVE_BYTES), sum(TIME_TRAVEL_BYTES), sum(FAILSAFE_BYTES), sum(RETAINED_FOR_CLONE_BYTES)
	FROM ACCOUNT_USAGE.TABLE_STORAGE_METRICS
	GROUP BY TABLE_NAME, ID, TABLE_CATALOG, TABLE_CATALOG_ID, TABLE_SCHEMA, TABLE_SCHEMA_ID;`

	// https://docs.snowflake.com/en/sql-reference/account-usage/replication_usage_history.html
	replicationMetricQuery = `SELECT DATABASE_NAME, DATABASE_ID, sum(CREDITS_USED), sum(BYTES_TRANSFERRED) 
	FROM ACCOUNT_USAGE.REPLICATION_USAGE_HISTORY
	WHERE START_TIME >= dateadd(hour, -24, current_timestamp())
	GROUP BY DATABASE_NAME, DATABASE_ID;`
)
