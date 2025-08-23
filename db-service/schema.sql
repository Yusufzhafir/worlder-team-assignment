-- Create tables for sensor ingestion
-- Database is created by env MYSQL_DATABASE=sensors (docker entrypoint)

-- Use strict SQL mode and UTC timestamps in your server config for consistency.

-- Drop & recreate table for idempotent local dev (comment DROP in prod)
-- DROP TABLE IF EXISTS sensor_readings;

CREATE TABLE IF NOT EXISTS sensor_readings (
  reading_id    BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  sensor_value  DOUBLE NOT NULL,
  sensor_type   VARCHAR(32) NOT NULL,
  id1           CHAR(8) NOT NULL,             -- enforce uppercase in app; see CHECK below
  id2           INT NOT NULL,
  ts            TIMESTAMP(6) NOT NULL,        -- event time (microsecond precision)
  created_at    TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),

  PRIMARY KEY (reading_id),

  -- Common query patterns
  KEY idx_ids_ts (id1, id2, ts),
  KEY idx_ts (ts),
  KEY idx_type_ts (sensor_type, ts),

  -- MySQL 8+ enforces CHECK
  CONSTRAINT chk_id1_uppercase CHECK (id1 REGEXP '^[A-Z0-9]{1,8}$')
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Optional helper view for quick counts per type per minute
-- CREATE OR REPLACE VIEW v_counts_per_minute AS
-- SELECT sensor_type, DATE_FORMAT(ts, '%Y-%m-%d %H:%i:00') AS minute, COUNT(*) AS cnt
-- FROM sensor_readings
-- GROUP BY sensor_type, minute;

-- Example user/grants are handled by MYSQL_USER/MYSQL_PASSWORD envs,
-- but this is how you'd do it manually:
-- CREATE USER IF NOT EXISTS 'app'@'%' IDENTIFIED BY 'app';
-- GRANT ALL PRIVILEGES ON sensors.* TO 'app'@'%';
-- FLUSH PRIVILEGES;
