package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// DB represents the database connection
type DB struct {
	conn *sql.DB
}

// ServerRegistration represents a server that has registered with the aggregator
type ServerRegistration struct {
	ID           int64     `json:"id"`
	Hostname     string    `json:"hostname"`
	IPAddress    string    `json:"ip_address"`
	SystemInfo   string    `json:"system_info"` // JSON blob
	Bonds        string    `json:"bonds"`       // JSON blob of bond -> IPs mapping
	RegisteredAt time.Time `json:"registered_at"`
	LastSeen     time.Time `json:"last_seen"`
}

// TestResult represents the result of a connectivity test
type TestResult struct {
	ID             int64     `json:"id"`
	SourceHostname string    `json:"source_hostname"`
	TargetHostname string    `json:"target_hostname"`
	TargetIP       string    `json:"target_ip"`
	SourceIP       string    `json:"source_ip"`
	BondName       string    `json:"bond_name"`
	TestType       string    `json:"test_type"` // "arp" or "http"
	Success        bool      `json:"success"`
	ResponseTime   int64     `json:"response_time_ms"` // milliseconds
	ErrorMessage   string    `json:"error_message,omitempty"`
	TestedAt       time.Time `json:"tested_at"`
}

// NewDB creates a new database connection and initializes tables
func NewDB(dbPath string) (*DB, error) {
	// Add WAL mode, busy_timeout, and other optimizations to prevent database locking
	connStr := dbPath + "?_journal_mode=WAL&_foreign_keys=ON&_txlock=immediate&_busy_timeout=30000"
	conn, err := sql.Open("sqlite", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Limit to 1 connection to prevent database locking
	conn.SetMaxOpenConns(1)

	db := &DB{conn: conn}

	if err := db.initTables(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to initialize tables: %w", err)
	}

	return db, nil
}

// initTables creates the necessary database tables
func (db *DB) initTables() error {
	schemas := []string{
		`CREATE TABLE IF NOT EXISTS servers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			hostname TEXT NOT NULL UNIQUE,
			ip_address TEXT NOT NULL,
			system_info TEXT NOT NULL,
			bonds TEXT NOT NULL,
			registered_at DATETIME NOT NULL,
			last_seen DATETIME NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS test_results (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			source_hostname TEXT NOT NULL,
			target_hostname TEXT NOT NULL,
			target_ip TEXT NOT NULL,
			source_ip TEXT NOT NULL,
			bond_name TEXT NOT NULL,
			test_type TEXT NOT NULL,
			success INTEGER NOT NULL,
			response_time_ms INTEGER,
			error_message TEXT,
			tested_at DATETIME NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_servers_hostname ON servers(hostname)`,
		`CREATE INDEX IF NOT EXISTS idx_test_results_source ON test_results(source_hostname)`,
		`CREATE INDEX IF NOT EXISTS idx_test_results_target ON test_results(target_hostname)`,
		`CREATE INDEX IF NOT EXISTS idx_test_results_tested_at ON test_results(tested_at)`,
	}

	for _, schema := range schemas {
		if _, err := db.conn.Exec(schema); err != nil {
			return fmt.Errorf("failed to execute schema: %w", err)
		}
	}

	return nil
}

// RegisterServer registers or updates a server in the database
func (db *DB) RegisterServer(hostname, ipAddress string, systemInfo interface{}, bonds map[string][]string) error {
	systemInfoJSON, err := json.Marshal(systemInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal system info: %w", err)
	}

	bondsJSON, err := json.Marshal(bonds)
	if err != nil {
		return fmt.Errorf("failed to marshal bonds: %w", err)
	}

	now := time.Now()

	_, err = db.conn.Exec(`
		INSERT INTO servers (hostname, ip_address, system_info, bonds, registered_at, last_seen)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(hostname) DO UPDATE SET
			ip_address = excluded.ip_address,
			system_info = excluded.system_info,
			bonds = excluded.bonds,
			last_seen = excluded.last_seen
	`, hostname, ipAddress, string(systemInfoJSON), string(bondsJSON), now, now)

	if err != nil {
		return fmt.Errorf("failed to register server: %w", err)
	}

	return nil
}

// GetAllServers returns all registered servers
func (db *DB) GetAllServers() ([]ServerRegistration, error) {
	rows, err := db.conn.Query(`
		SELECT id, hostname, ip_address, system_info, bonds, registered_at, last_seen
		FROM servers
		ORDER BY hostname
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query servers: %w", err)
	}
	defer rows.Close()

	var servers []ServerRegistration
	for rows.Next() {
		var server ServerRegistration
		if err := rows.Scan(
			&server.ID,
			&server.Hostname,
			&server.IPAddress,
			&server.SystemInfo,
			&server.Bonds,
			&server.RegisteredAt,
			&server.LastSeen,
		); err != nil {
			return nil, fmt.Errorf("failed to scan server: %w", err)
		}
		servers = append(servers, server)
	}

	return servers, nil
}

// GetServer returns a specific server by hostname
func (db *DB) GetServer(hostname string) (*ServerRegistration, error) {
	var server ServerRegistration
	err := db.conn.QueryRow(`
		SELECT id, hostname, system_info, bonds, registered_at, last_seen
		FROM servers
		WHERE hostname = ?
	`, hostname).Scan(
		&server.ID,
		&server.Hostname,
		&server.SystemInfo,
		&server.Bonds,
		&server.RegisteredAt,
		&server.LastSeen,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get server: %w", err)
	}

	return &server, nil
}

// SaveTestResult saves a connectivity test result
func (db *DB) SaveTestResult(result TestResult) error {
	_, err := db.conn.Exec(`
		INSERT INTO test_results (
			source_hostname, target_hostname, target_ip, source_ip, bond_name, test_type,
			success, response_time_ms, error_message, tested_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		result.SourceHostname,
		result.TargetHostname,
		result.TargetIP,
		result.SourceIP,
		result.BondName,
		result.TestType,
		result.Success,
		result.ResponseTime,
		result.ErrorMessage,
		result.TestedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save test result: %w", err)
	}

	return nil
}

// GetTestResults returns recent test results
func (db *DB) GetTestResults(limit int) ([]TestResult, error) {
	query := `
		SELECT id, source_hostname, target_hostname, target_ip, source_ip, bond_name, test_type,
			   success, response_time_ms, error_message, tested_at
		FROM test_results
		ORDER BY tested_at DESC
	`

	var rows *sql.Rows
	var err error

	if limit > 0 {
		query += " LIMIT ?"
		rows, err = db.conn.Query(query, limit)
	} else {
		rows, err = db.conn.Query(query)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query test results: %w", err)
	}
	defer rows.Close()

	var results []TestResult
	for rows.Next() {
		var result TestResult
		if err := rows.Scan(
			&result.ID,
			&result.SourceHostname,
			&result.TargetHostname,
			&result.TargetIP,
			&result.SourceIP,
			&result.BondName,
			&result.TestType,
			&result.Success,
			&result.ResponseTime,
			&result.ErrorMessage,
			&result.TestedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan test result: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// GetTestResultsBySource returns test results for a specific source hostname
func (db *DB) GetTestResultsBySource(hostname string, limit int) ([]TestResult, error) {
	query := `
		SELECT id, source_hostname, target_hostname, target_ip, source_ip, bond_name, test_type,
			   success, response_time_ms, error_message, tested_at
		FROM test_results
		WHERE source_hostname = ?
		ORDER BY tested_at DESC
	`

	var rows *sql.Rows
	var err error

	if limit > 0 {
		query += " LIMIT ?"
		rows, err = db.conn.Query(query, hostname, limit)
	} else {
		rows, err = db.conn.Query(query, hostname)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query test results: %w", err)
	}
	defer rows.Close()

	var results []TestResult
	for rows.Next() {
		var result TestResult
		if err := rows.Scan(
			&result.ID,
			&result.SourceHostname,
			&result.TargetHostname,
			&result.TargetIP,
			&result.SourceIP,
			&result.BondName,
			&result.TestType,
			&result.Success,
			&result.ResponseTime,
			&result.ErrorMessage,
			&result.TestedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan test result: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// ClearTestResults deletes all test results from the database
func (db *DB) ClearTestResults() error {
	_, err := db.conn.Exec("DELETE FROM test_results")
	if err != nil {
		return fmt.Errorf("failed to clear test results: %w", err)
	}
	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}
