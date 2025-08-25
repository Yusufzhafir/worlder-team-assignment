package model

import "time"

// Full row as read from DB
type SensorReading struct {
	ReadingID   uint64    `db:"reading_id"`
	SensorValue float64   `db:"sensor_value"`
	SensorType  string    `db:"sensor_type"`
	ID1         string    `db:"id1"`
	ID2         int       `db:"id2"`
	TS          time.Time `db:"ts"`         // TIMESTAMP(6)
	CreatedAt   time.Time `db:"created_at"` // TIMESTAMP(6)
}

// Insert DTO (omit auto fields)
type SensorReadingInsert struct {
	SensorValue float64   `db:"sensor_value"`
	SensorType  string    `db:"sensor_type"`
	ID1         string    `db:"id1"`
	ID2         int       `db:"id2"`
	TS          time.Time `db:"ts"`
}
