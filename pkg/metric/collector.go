package metric

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
)

// RegisterDBStats registers SQL DB stats with Prometheus
func RegisterDBStats(db *sql.DB, dbName string) {
	labels := prometheus.Labels{"db_name": dbName}

	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name:        "go_sql_db_max_open_connections",
			Help:        "Maximum number of open connections to the database",
			ConstLabels: labels,
		},
		func() float64 {
			return float64(db.Stats().MaxOpenConnections)
		},
	))

	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name:        "go_sql_db_open_connections",
			Help:        "The number of established connections both in use and idle",
			ConstLabels: labels,
		},
		func() float64 {
			return float64(db.Stats().OpenConnections)
		},
	))

	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name:        "go_sql_db_in_use_connections",
			Help:        "The number of connections currently in use",
			ConstLabels: labels,
		},
		func() float64 {
			return float64(db.Stats().InUse)
		},
	))

	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name:        "go_sql_db_idle_connections",
			Help:        "The number of idle connections",
			ConstLabels: labels,
		},
		func() float64 {
			return float64(db.Stats().Idle)
		},
	))

	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name:        "go_sql_db_wait_count_total",
			Help:        "The total number of connections waited for",
			ConstLabels: labels,
		},
		func() float64 {
			return float64(db.Stats().WaitCount)
		},
	))

	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name:        "go_sql_db_wait_duration_seconds_total",
			Help:        "The total time blocked waiting for a new connection",
			ConstLabels: labels,
		},
		func() float64 {
			return db.Stats().WaitDuration.Seconds()
		},
	))
}

// RegisterRedisStats registers Redis Pool stats with Prometheus
func RegisterRedisStats(rdb *redis.Client, poolName string) {
	labels := prometheus.Labels{"pool_name": poolName}

	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name:        "go_redis_pool_hits_total",
			Help:        "Number of times free connection was found in the pool",
			ConstLabels: labels,
		},
		func() float64 {
			return float64(rdb.PoolStats().Hits)
		},
	))

	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name:        "go_redis_pool_misses_total",
			Help:        "Number of times free connection was NOT found in the pool",
			ConstLabels: labels,
		},
		func() float64 {
			return float64(rdb.PoolStats().Misses)
		},
	))

	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name:        "go_redis_pool_timeouts_total",
			Help:        "Number of times a wait timeout occurred",
			ConstLabels: labels,
		},
		func() float64 {
			return float64(rdb.PoolStats().Timeouts)
		},
	))

	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name:        "go_redis_pool_total_conns",
			Help:        "Number of total connections in the pool",
			ConstLabels: labels,
		},
		func() float64 {
			return float64(rdb.PoolStats().TotalConns)
		},
	))

	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name:        "go_redis_pool_idle_conns",
			Help:        "Number of idle connections in the pool",
			ConstLabels: labels,
		},
		func() float64 {
			return float64(rdb.PoolStats().IdleConns)
		},
	))

	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name:        "go_redis_pool_stale_conns",
			Help:        "Number of stale connections removed from the pool",
			ConstLabels: labels,
		},
		func() float64 {
			return float64(rdb.PoolStats().StaleConns)
		},
	))
}

// MonitorCronService monitors running jobs
type CronMonitor interface {
	RunningJobCount() int
}

func RegisterCronStats(monitor CronMonitor) {
	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "nuwa_cron_running_jobs",
			Help: "Number of currently running cron jobs",
		},
		func() float64 {
			return float64(monitor.RunningJobCount())
		},
	))
}
