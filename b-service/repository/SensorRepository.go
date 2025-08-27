package repository

import (
	"context"
	"database/sql"
	"time"

	model "github.com/Yusufzhafir/worlder-team-assignment/b-service/repository/model"

	"github.com/jmoiron/sqlx"
)

type SensorRepository interface {
	InsertReadingTx(ctx context.Context, db *sqlx.DB, r *model.SensorReadingInsert) (uint64, error)
	InsertReadingsBatchTx(ctx context.Context, db *sqlx.DB, rs []model.SensorReadingInsert) (int64, error)
	SelectByTime(ctx context.Context, db *sqlx.DB, startTime, stopTime time.Time, limit int, offset int) ([]model.SensorReading, error)
	SelectCountByTime(ctx context.Context, db *sqlx.DB, startTime, stopTime time.Time) (int64, error)
}

type SensorRepositoryImpl struct {
}

func NewSensorRepository() SensorRepository {
	return &SensorRepositoryImpl{}
}

const insertReadingSQL = `
INSERT INTO sensor_readings (sensor_value, sensor_type, id1, id2, ts)
VALUES (:sensor_value, :sensor_type, :id1, :id2, :ts)
`

func (sensorRepo *SensorRepositoryImpl) InsertReadingTx(ctx context.Context, db *sqlx.DB, r *model.SensorReadingInsert) (uint64, error) {
	tx, err := db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return 0, err
	}
	// Roll back on any error
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	res, err := tx.NamedExecContext(ctx, insertReadingSQL, r)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return uint64(id), nil
}

func (sensorRepo *SensorRepositoryImpl) InsertReadingsBatchTx(ctx context.Context, db *sqlx.DB, rs []model.SensorReadingInsert) (int64, error) {
	if len(rs) == 0 {
		return 0, nil
	}

	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	stmt, err := tx.PrepareNamedContext(ctx, insertReadingSQL)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	var affected int64
	for _, r := range rs {
		res, execErr := stmt.ExecContext(ctx, r)
		if execErr != nil {
			err = execErr
			return affected, err
		}
		if n, _ := res.RowsAffected(); n > 0 {
			affected += n
		}
	}

	if err = tx.Commit(); err != nil {
		return affected, err
	}
	return affected, nil
}

const selectSensorsDataByTimePaginated = `
SELECT sensor_value, sensor_type, id1, id2, ts
FROM sensor_readings
WHERE ts >= :ts_start AND ts < :ts_stop
ORDER BY ts
LIMIT :limit OFFSET :offset
`

type SelectByTimePageArgs struct {
	TsStart time.Time `db:"ts_start"`
	TsStop  time.Time `db:"ts_stop"`
	Limit   int       `db:"limit"`
	Offset  int       `db:"offset"`
}

func (repo *SensorRepositoryImpl) SelectByTime(
	ctx context.Context,
	db *sqlx.DB,
	startTime, stopTime time.Time,
	limit, offset int,
) ([]model.SensorReading, error) {
	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	args := SelectByTimePageArgs{
		TsStart: startTime,
		TsStop:  stopTime,
		Limit:   limit,
		Offset:  offset,
	}

	rows, err := db.NamedQuery(selectSensorsDataByTimePaginated, args)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var readings []model.SensorReading
	for rows.Next() {
		var r model.SensorReading
		if err := rows.StructScan(&r); err != nil {
			return nil, err
		}
		readings = append(readings, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return readings, nil
}

const selectCountSensorsDataByTimePaginated = `
SELECT count(id1) AS cnt
FROM sensor_readings
WHERE ts >= ? AND ts < ?
`

type CountResult struct {
	Cnt int64 `db:"cnt"`
}

func (repo *SensorRepositoryImpl) SelectCountByTime(
	ctx context.Context,
	db *sqlx.DB,
	startTime, stopTime time.Time,
) (int64, error) {
	var result CountResult
	if err := db.GetContext(ctx, &result, selectCountSensorsDataByTimePaginated, startTime, stopTime); err != nil {
		return 0, err
	}

	return result.Cnt, nil
}
