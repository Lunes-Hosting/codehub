package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	
	_ "github.com/go-sql-driver/mysql"
)

// DB is the global database connection
var DB *sql.DB

// InitDB initializes the database connection
func InitDB() (*sql.DB, error) {
	// Get database connection details from environment variables directly
	dbUser := os.Getenv("MYSQL_USER")
	dbPassword := os.Getenv("MYSQL_PASSWORD")
	dbHost := os.Getenv("MYSQL_HOST")
	dbName := os.Getenv("MYSQL_DATABASE")
	
	// Log connection details (without password)
	log.Printf("Connecting to database: %s@tcp(%s)/%s", dbUser, dbHost, dbName)
	
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser,
		dbPassword,
		dbHost,
		dbName,
	)
	
	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	
	if err = DB.Ping(); err != nil {
		return nil, err
	}
	
	return DB, nil
}

// CloseDB closes the database connection
func CloseDB() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// GetDB returns the database connection
func GetDB() *sql.DB {
	return DB
}

// ExecuteQuery executes a query and returns the result
func ExecuteQuery(query string, args ...interface{}) (*sql.Rows, error) {
	return DB.Query(query, args...)
}

// ExecuteQueryRow executes a query and returns a single row
func ExecuteQueryRow(query string, args ...interface{}) *sql.Row {
	return DB.QueryRow(query, args...)
}

// ExecuteStatement executes a statement (INSERT, UPDATE, DELETE)
func ExecuteStatement(query string, args ...interface{}) (sql.Result, error) {
	return DB.Exec(query, args...)
}

// ScanRow scans a row into a map
func ScanRow(rows *sql.Rows) (map[string]interface{}, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	
	count := len(columns)
	values := make([]interface{}, count)
	valuePtrs := make([]interface{}, count)
	
	for i := range columns {
		valuePtrs[i] = &values[i]
	}
	
	if err := rows.Scan(valuePtrs...); err != nil {
		return nil, err
	}
	
	result := make(map[string]interface{})
	for i, col := range columns {
		val := values[i]
		
		b, ok := val.([]byte)
		if ok {
			result[col] = string(b)
		} else {
			result[col] = val
		}
	}
	
	return result, nil
}

// ScanRows scans all rows into a slice of maps
func ScanRows(rows *sql.Rows) ([]map[string]interface{}, error) {
	defer rows.Close()
	
	var result []map[string]interface{}
	
	for rows.Next() {
		row, err := ScanRow(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	
	return result, nil
}
