package main

import (
	"database/sql"
	"fmt"
	"log"
)

// 从数据库加载报警配置
func loadAlertConfigFromDB() {
	query := `SELECT cpu_threshold, memory_threshold, disk_threshold, cpu_duration, memory_duration FROM alert_config WHERE id = 1`
	row := db.QueryRow(query)

	err := row.Scan(&alertConfig.CPUThreshold, &alertConfig.MemoryThreshold, &alertConfig.DiskThreshold, &alertConfig.CPUDuration, &alertConfig.MemoryDuration)
	if err != nil {
		if err == sql.ErrNoRows {
			saveAlertConfigToDB(alertConfig)
			fmt.Println("No alert config found, using default config")
		} else {
			log.Fatalf("Failed to load alert config: %v", err)
		}
	} else {
		fmt.Println("Successfully loaded alert config from database")
	}
}

// 将报警配置保存到数据库
func saveAlertConfigToDB(config AlertConfig) error {
	query := `REPLACE INTO alert_config (id, cpu_threshold, memory_threshold, disk_threshold, cpu_duration, memory_duration) VALUES (1, ?, ?, ?, ?, ?)`
	_, err := db.Exec(query, config.CPUThreshold, config.MemoryThreshold, config.DiskThreshold, config.CPUDuration, config.MemoryDuration)
	return err
}
