package utils

// Status responses contain measurements only, never DSNs or credentials.
type DatabaseStatus struct {
	Type      string              `json:"type"`
	Healthy   bool                `json:"healthy"`
	Message   string              `json:"message"`
	LatencyMS float64             `json:"latencyMs"`
	CheckedAt string              `json:"checkedAt"`
	Pool      *DatabasePoolStatus `json:"pool,omitempty"`
	MySQL     *MySQLStatus        `json:"mysql,omitempty"`
}
type DatabasePoolStatus struct {
	MaxOpenConnections int     `json:"maxOpenConnections"`
	OpenConnections    int     `json:"openConnections"`
	InUse              int     `json:"inUse"`
	Idle               int     `json:"idle"`
	WaitCount          int64   `json:"waitCount"`
	WaitDurationMS     float64 `json:"waitDurationMs"`
}
type MySQLStatus struct {
	Version          string   `json:"version,omitempty"`
	MaxConnections   *uint64  `json:"maxConnections"`
	ReadOnly         *bool    `json:"readOnly"`
	Uptime           *uint64  `json:"uptime"`
	ThreadsConnected *uint64  `json:"threadsConnected"`
	ThreadsRunning   *uint64  `json:"threadsRunning"`
	Questions        *uint64  `json:"questions"`
	SlowQueries      *uint64  `json:"slowQueries"`
	QPS              *float64 `json:"qps"`
	Warning          string   `json:"warning,omitempty"`
}
