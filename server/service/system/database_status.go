package system

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	"github.com/isgvto/gin-vue-admin-gblog/server/utils"
)

const mysqlBasicSQL = "SELECT VERSION(), @@GLOBAL.max_connections, @@GLOBAL.read_only"
const mysqlStatusSQL = "SHOW GLOBAL STATUS WHERE Variable_name IN ('Uptime','Threads_connected','Threads_running','Questions','Slow_queries')"

type databaseQuery interface {
	PingContext(context.Context) error
	Stats() sql.DBStats
	QueryRowContext(context.Context, string, ...any) *sql.Row
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}
type qpsSample struct {
	at                time.Time
	uptime, questions uint64
}
type qpsTracker struct {
	sync.Mutex
	db     databaseQuery
	sample *qpsSample
}

var databaseRates qpsTracker

func (tracker *qpsTracker) rate(db databaseQuery, now time.Time, uptime, questions uint64) *float64 {
	tracker.Lock()
	defer tracker.Unlock()
	previous := tracker.sample
	if tracker.db != db {
		previous = nil
	}
	if previous != nil && !now.After(previous.at) {
		return nil
	}
	tracker.db = db
	tracker.sample = &qpsSample{at: now, uptime: uptime, questions: questions}
	if previous == nil || uptime < previous.uptime || questions < previous.questions {
		return nil
	}
	seconds := now.Sub(previous.at).Seconds()
	if seconds <= 0 {
		return nil
	}
	value := float64(questions-previous.questions) / seconds
	return &value
}

func collectDatabaseStatus(parent context.Context) utils.DatabaseStatus {
	if global.GVA_DB == nil {
		return utils.DatabaseStatus{Type: global.GVA_CONFIG.System.DbType, Message: "数据库未初始化", CheckedAt: time.Now().UTC().Format(time.RFC3339)}
	}
	db, err := global.GVA_DB.DB()
	if err != nil {
		return utils.DatabaseStatus{Type: global.GVA_DB.Dialector.Name(), Message: "无法读取数据库连接池", CheckedAt: time.Now().UTC().Format(time.RFC3339)}
	}
	return inspectDatabase(parent, db, global.GVA_DB.Dialector.Name(), &databaseRates)
}

func inspectDatabase(parent context.Context, db databaseQuery, kind string, tracker *qpsTracker) utils.DatabaseStatus {
	ctx, cancel := context.WithTimeout(parent, 3*time.Second)
	defer cancel()
	status := utils.DatabaseStatus{Type: kind, CheckedAt: time.Now().UTC().Format(time.RFC3339)}
	start := time.Now()
	err := db.PingContext(ctx)
	status.LatencyMS = float64(time.Since(start)) / float64(time.Millisecond)
	pool := db.Stats()
	status.Pool = &utils.DatabasePoolStatus{MaxOpenConnections: pool.MaxOpenConnections, OpenConnections: pool.OpenConnections, InUse: pool.InUse, Idle: pool.Idle, WaitCount: pool.WaitCount, WaitDurationMS: float64(pool.WaitDuration) / float64(time.Millisecond)}
	if err != nil {
		status.Message = "数据库连接失败"
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
			status.Message = "数据库连接检测超时"
		}
		return status
	}
	status.Healthy = true
	status.Message = "连接正常"
	if kind != "mysql" {
		return status
	}
	mysql := &utils.MySQLStatus{}
	status.MySQL = mysql
	var version string
	var max uint64
	var readOnly bool
	if err = db.QueryRowContext(ctx, mysqlBasicSQL).Scan(&version, &max, &readOnly); err == nil {
		mysql.Version = version
		mysql.MaxConnections = &max
		mysql.ReadOnly = &readOnly
	} else {
		mysql.Warning = "部分基础信息无法读取，请检查数据库权限或版本兼容性"
	}
	rows, err := db.QueryContext(ctx, mysqlStatusSQL)
	if err != nil {
		mysql.Warning = "MySQL 全局指标无法读取，请检查数据库权限或稍后重试"
		return status
	}
	defer rows.Close()
	values := map[string]*uint64{}
	for rows.Next() {
		var key, value string
		if err = rows.Scan(&key, &value); err != nil {
			mysql.Warning = "MySQL 全局指标读取不完整"
			continue
		}
		number, parseErr := strconv.ParseUint(value, 10, 64)
		if parseErr == nil {
			values[key] = &number
		}
	}
	mysql.Uptime = values["Uptime"]
	mysql.ThreadsConnected = values["Threads_connected"]
	mysql.ThreadsRunning = values["Threads_running"]
	mysql.Questions = values["Questions"]
	mysql.SlowQueries = values["Slow_queries"]
	if rows.Err() != nil {
		mysql.Warning = "MySQL 全局指标读取不完整"
		return status
	}
	if len(values) < 5 && mysql.Warning == "" {
		mysql.Warning = "部分全局指标未返回"
	}
	if mysql.Uptime != nil && mysql.Questions != nil {
		mysql.QPS = tracker.rate(db, time.Now(), *mysql.Uptime, *mysql.Questions)
	}
	return status
}
