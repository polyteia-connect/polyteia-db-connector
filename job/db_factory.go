package job

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"time"

	_ "github.com/go-sql-driver/mysql"  // MySQL driver
	_ "github.com/lib/pq"               // PostgreSQL driver
	_ "github.com/microsoft/go-mssqldb" // MS SQL Server driver
)

// OpenDatabase opens a database connection using the appropriate driver
// based on the database type. Returns a *sql.DB that implements the standard
// database/sql interface.
func OpenDatabase(ctx context.Context, dbConfig SourceDatabaseConfig) (*sql.DB, error) {
	var driverName string
	var connString string
	var err error

	switch dbConfig.Type {
	case "postgres", "postgresql":
		driverName = "postgres"
		connString = buildPostgresConnectionString(dbConfig)
	case "mysql", "mariadb":
		driverName = "mysql"
		connString = buildMySQLConnectionString(dbConfig)
	case "mssql", "sqlserver":
		driverName = "sqlserver"
		connString = buildMSSQLConnectionString(dbConfig)
	default:
		return nil, fmt.Errorf("unsupported database type: %s (supported: postgres, mysql, mssql)", dbConfig.Type)
	}

	db, err := sql.Open(driverName, connString)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s connection: %w", dbConfig.Type, err)
	}

	// Configure connection pool settings
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)

	// Test the connection with context timeout
	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		db.Close() //nolint
		return nil, fmt.Errorf("failed to ping %s database: %w", dbConfig.Type, err)
	}

	return db, nil
}

func buildPostgresConnectionString(dbConfig SourceDatabaseConfig) string {
	// PostgreSQL connection string format: postgres://user:password@host:port/dbname?sslmode=disable
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		url.QueryEscape(dbConfig.User),
		url.QueryEscape(dbConfig.Password),
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.Name,
	)
	return connStr
}

func buildMySQLConnectionString(dbConfig SourceDatabaseConfig) string {
	// MySQL connection string format: user:password@tcp(host:port)/dbname?parseTime=true
	// Note: MySQL driver uses DSN format. For passwords with special characters,
	// users should URL encode them in the configuration.
	connStr := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		url.QueryEscape(dbConfig.User),
		url.QueryEscape(dbConfig.Password),
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.Name,
	)
	return connStr
}

func buildMSSQLConnectionString(dbConfig SourceDatabaseConfig) string {
	// MS SQL Server connection string format: sqlserver://user:password@host:port?database=dbname&encrypt=true&trustServerCertificate=true
	connStr := fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=%s&encrypt=true&trustServerCertificate=true",
		url.QueryEscape(dbConfig.User),
		url.QueryEscape(dbConfig.Password),
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.Name,
	)
	return connStr
}
