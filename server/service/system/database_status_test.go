package system

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
	"time"
)

type statusDriver struct{}

func (statusDriver) Open(mode string) (driver.Conn, error) { return &statusConn{mode: mode}, nil }

type statusConn struct{ mode string }

func (*statusConn) Prepare(string) (driver.Stmt, error) { return nil, errors.New("unexpected prepare") }
func (*statusConn) Close() error                        { return nil }
func (*statusConn) Begin() (driver.Tx, error)           { return nil, errors.New("unexpected transaction") }
func (c *statusConn) Ping(ctx context.Context) error {
	if c.mode == "down" {
		return errors.New("password=secret db-host")
	}
	if c.mode == "timeout" {
		<-ctx.Done()
		return ctx.Err()
	}
	return nil
}
func (c *statusConn) QueryContext(ctx context.Context, query string, _ []driver.NamedValue) (driver.Rows, error) {
	if c.mode == "denied" {
		return nil, errors.New("permission denied for secret user")
	}
	if query == mysqlBasicSQL {
		return &statusRows{columns: []string{"version", "max", "readonly"}, values: [][]driver.Value{{"8.4-test", int64(151), int64(0)}}}, nil
	}
	if query == mysqlStatusSQL {
		return &statusRows{columns: []string{"Variable_name", "Value"}, values: [][]driver.Value{{"Uptime", "90000"}, {"Threads_connected", "12"}, {"Threads_running", "2"}, {"Questions", "100"}, {"Slow_queries", "3"}}}, nil
	}
	return nil, errors.New("unexpected SQL")
}

type statusRows struct {
	columns []string
	values  [][]driver.Value
	index   int
}

func (r *statusRows) Columns() []string { return r.columns }
func (*statusRows) Close() error        { return nil }
func (r *statusRows) Next(values []driver.Value) error {
	if r.index == len(r.values) {
		return io.EOF
	}
	copy(values, r.values[r.index])
	r.index++
	return nil
}
func init() { sql.Register("gblog-status-test", statusDriver{}) }
func testStatusDB(t *testing.T, mode string) *sql.DB {
	t.Helper()
	db, err := sql.Open("gblog-status-test", mode)
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(5)
	t.Cleanup(func() { db.Close() })
	return db
}

func TestDatabaseStatusHealthyAndPartialFailure(t *testing.T) {
	for _, mode := range []string{"healthy", "down", "denied"} {
		t.Run(mode, func(t *testing.T) {
			db := testStatusDB(t, mode)
			status := inspectDatabase(context.Background(), db, "mysql", &qpsTracker{})
			if status.Pool == nil || status.Pool.MaxOpenConnections != 5 {
				t.Fatal("pool missing")
			}
			raw, _ := json.Marshal(status)
			if strings.Contains(string(raw), "secret") || strings.Contains(string(raw), "db-host") {
				t.Fatal("credentials leaked")
			}
			if mode == "down" {
				if status.Healthy || status.MySQL != nil {
					t.Fatal("failed connection reported healthy")
				}
				return
			}
			if !status.Healthy || status.MySQL == nil || status.Pool.OpenConnections != 1 {
				t.Fatal("health or current pool snapshot missing")
			}
			if mode == "denied" {
				if status.MySQL.Warning == "" || status.MySQL.ThreadsConnected != nil {
					t.Fatal("missing metrics represented as zero")
				}
				return
			}
			if status.MySQL.Version != "8.4-test" || *status.MySQL.ThreadsConnected != 12 || *status.MySQL.SlowQueries != 3 || *status.MySQL.ReadOnly || status.MySQL.QPS != nil {
				t.Fatal("wrong metrics or initial QPS")
			}
			if db.Stats().InUse != 0 {
				t.Fatal("query leaked connection")
			}
		})
	}
}
func TestDatabaseStatusContextTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	status := inspectDatabase(ctx, testStatusDB(t, "timeout"), "mysql", &qpsTracker{})
	if status.Healthy || !strings.Contains(status.Message, "超时") {
		t.Fatal("timeout not isolated")
	}
}
func TestDatabaseQPSResets(t *testing.T) {
	db := testStatusDB(t, "healthy")
	tracker := &qpsTracker{}
	now := time.Now()
	if tracker.rate(db, now, 100, 100) != nil {
		t.Fatal("first sample returned rate")
	}
	rate := tracker.rate(db, now.Add(10*time.Second), 110, 150)
	if rate == nil || *rate != 5 {
		t.Fatal("wrong sample rate")
	}
	if tracker.rate(db, now.Add(20*time.Second), 2, 160) != nil {
		t.Fatal("restart not detected")
	}
	if tracker.rate(db, now.Add(30*time.Second), 12, 1) != nil {
		t.Fatal("counter reset not detected")
	}
	if tracker.rate(testStatusDB(t, "healthy"), now.Add(40*time.Second), 20, 100) != nil {
		t.Fatal("different database reused rate")
	}
}
