CREATE TABLE IF NOT EXISTS alert_config (
                              id INT AUTO_INCREMENT PRIMARY KEY,
                              cpu_threshold FLOAT NOT NULL,
                              memory_threshold FLOAT NOT NULL,
                              disk_threshold FLOAT NOT NULL,
                              cpu_duration FLOAT NOT NULL,
                              memory_duration FLOAT NOT NULL
);

CREATE TABLE IF NOT EXISTS hosts (
                       id INT AUTO_INCREMENT PRIMARY KEY,
                       ip_address VARCHAR(255) NOT NULL UNIQUE,
                       label VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
                       alert_enabled BOOLEAN NOT NULL
);
