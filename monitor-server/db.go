package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

var MYSQL_HOST = "10.15.97.43"
var MYSQL_PORT = 3306
var MYSQL_USER = "app"
var MYSQL_PASS = "app"
var dsn = ""
var dsnWithDB = ""

// initDB 初始化 MySQL 数据库连接和表结构
func initDB() {
	dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/", MYSQL_USER, MYSQL_PASS, MYSQL_HOST, MYSQL_PORT)
	var err error
	// MySQL 连接不指定数据库，首先连接到 MySQL 服务器
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error connecting to the MySQL server: %v", err)
	}

	// 检查指定数据库是否存在
	dbName := "test"
	if !checkDatabaseExists(dbName) {
		// 如果数据库不存在，创建数据库
		err = createDatabase(dbName)
		if err != nil {
			log.Fatalf("Failed to create database: %v", err)
		}
	}

	// 切换到指定的数据库
	err = db.Close() // 关闭旧的数据库连接
	if err != nil {
		log.Fatalf("Failed to close connection: %v", err)
	}

	// 使用指定数据库进行连接
	dsnWithDB = dsn + dbName
	db, err = sql.Open("mysql", dsnWithDB)
	if err != nil {
		log.Fatalf("Error connecting to the specified database: %v", err)
	}

	// 初始化表结构
	err = initializeTables("resources/mysql.sql")
	if err != nil {
		log.Fatalf("Failed to initialize tables: %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("Error pinging the database: %v", err)
	}

	log.Println("Successfully connected to MySQL database")

}

// checkDatabaseExists 检查指定数据库是否存在
func checkDatabaseExists(dbName string) bool {
	var exists bool
	query := "SELECT COUNT(*) FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = ?"
	err := db.QueryRow(query, dbName).Scan(&exists)
	if err != nil {
		log.Fatalf("Error checking database existence: %v", err)
	}
	return exists
}

// createDatabase 创建指定的数据库
func createDatabase(dbName string) error {
	_, err := db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName))
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}
	log.Printf("Database %s created successfully", dbName)
	return nil
}

// initializeTables 从指定的 SQL 文件中读取并执行初始化表结构
func initializeTables(sqlFilePath string) error {
	// 读取 SQL 文件内容
	sqlBytes, err := ioutil.ReadFile(sqlFilePath)
	if err != nil {
		return fmt.Errorf("failed to read SQL file: %w", err)
	}

	// 将文件内容转换为字符串
	sqlStatements := string(sqlBytes)

	// 使用分号拆分 SQL 语句
	statements := strings.Split(sqlStatements, ";")

	// 逐条执行 SQL 语句
	for _, statement := range statements {
		// 移除前后的空白字符
		statement = strings.TrimSpace(statement)
		if statement == "" {
			continue // 跳过空的语句
		}

		// 执行 SQL 语句
		_, err := db.Exec(statement)
		if err != nil {
			return fmt.Errorf("failed to execute SQL statement: %w\nStatement: %s", err, statement)
		}
	}

	log.Println("Tables initialized successfully")
	return nil
}
