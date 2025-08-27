package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	model "github.com/Yusufzhafir/worlder-team-assignment/b-service/repository/model"

	"github.com/jmoiron/sqlx"
)

type IDCombination struct {
	ID1 string `db:"id1"`
	ID2 int    `db:"id2"`
}

type SensorRepository interface {
	InsertReadingTx(ctx context.Context, db *sqlx.DB, r *model.SensorReadingInsert) (uint64, error)
	InsertReadingsBatchTx(ctx context.Context, db *sqlx.DB, rs []model.SensorReadingInsert) (int64, error)

	// Select by time (already implemented)
	SelectByTime(ctx context.Context, db *sqlx.DB, startTime, stopTime time.Time, limit int, offset int) ([]model.SensorReading, error)
	SelectCountByTime(ctx context.Context, db *sqlx.DB, startTime, stopTime time.Time) (int64, error)

	// Select by ID combinations
	SelectByIDs(ctx context.Context, db *sqlx.DB, ids []IDCombination, limit int, offset int) ([]model.SensorReading, error)
	SelectCountByIDs(ctx context.Context, db *sqlx.DB, ids []IDCombination) (int64, error)

	// Select by ID combinations and time
	SelectByIDsAndTime(ctx context.Context, db *sqlx.DB, ids []IDCombination, startTime, stopTime time.Time, limit int, offset int) ([]model.SensorReading, error)
	SelectCountByIDsAndTime(ctx context.Context, db *sqlx.DB, ids []IDCombination, startTime, stopTime time.Time) (int64, error)

	// Delete operations
	DeleteByTime(ctx context.Context, db *sqlx.DB, startTime, stopTime time.Time) (int64, error)
	DeleteByIDs(ctx context.Context, db *sqlx.DB, ids []IDCombination) (int64, error)
	DeleteByIDsAndTime(ctx context.Context, db *sqlx.DB, ids []IDCombination, startTime, stopTime time.Time) (int64, error)

	// Update operations
	UpdateByTime(ctx context.Context, db *sqlx.DB, startTime, stopTime time.Time, sensorValue float64, sensorType string) (int64, error)
	UpdateByIDs(ctx context.Context, db *sqlx.DB, ids []IDCombination, sensorValue float64, sensorType string) (int64, error)
	UpdateByIDsAndTime(ctx context.Context, db *sqlx.DB, ids []IDCombination, startTime, stopTime time.Time, sensorValue float64, sensorType string) (int64, error)

	//default pagination
	SelectSensorDataPaginated(ctx context.Context, db *sqlx.DB, limit, offset int) ([]model.SensorReading, error)
	SelectCountPagination(ctx context.Context, db *sqlx.DB) (int64, error)
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

type selectSensorPageArgs struct {
	Limit  int `db:"limit"`
	Offset int `db:"offset"`
}

const selectSensorsDataPaginated = `
SELECT sensor_value, sensor_type, id1, id2, ts
FROM sensor_readings
ORDER BY ts
LIMIT :limit OFFSET :offset
`

func (repo *SensorRepositoryImpl) SelectSensorDataPaginated(
	ctx context.Context,
	db *sqlx.DB,
	limit, offset int,
) ([]model.SensorReading, error) {
	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	args := selectSensorPageArgs{
		Limit:  limit,
		Offset: offset,
	}

	rows, err := db.NamedQuery(selectSensorsDataPaginated, args)
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

const selectCountPaginated = `
SELECT count(id1) AS cnt
FROM sensor_readings
`

func (repo *SensorRepositoryImpl) SelectCountPagination(
	ctx context.Context,
	db *sqlx.DB,
) (int64, error) {
	var result CountResult
	if err := db.GetContext(ctx, &result, selectCountPaginated); err != nil {
		return 0, err
	}

	return result.Cnt, nil
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

// Helper function to build ID condition for SQL queries
func buildIDCondition(ids []IDCombination) (string, []interface{}) {
	if len(ids) == 0 {
		return "1=1", []interface{}{}
	}

	var conditions []string
	var args []interface{}

	for _, id := range ids {
		conditions = append(conditions, "(id1 = ? AND id2 = ?)")
		args = append(args, id.ID1, id.ID2)
	}

	return "(" + strings.Join(conditions, " OR ") + ")", args
}

// Select by ID combinations
func (repo *SensorRepositoryImpl) SelectByIDs(
	ctx context.Context,
	db *sqlx.DB,
	ids []IDCombination,
	limit, offset int,
) ([]model.SensorReading, error) {
	if len(ids) == 0 {
		return []model.SensorReading{}, nil
	}

	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	idCondition, args := buildIDCondition(ids)
	query := fmt.Sprintf(`
SELECT sensor_value, sensor_type, id1, id2, ts
FROM sensor_readings
WHERE %s
ORDER BY ts
LIMIT ? OFFSET ?
`, idCondition)

	args = append(args, limit, offset)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var readings []model.SensorReading
	for rows.Next() {
		var r model.SensorReading
		if err := rows.Scan(&r.SensorValue, &r.SensorType, &r.ID1, &r.ID2, &r.TS); err != nil {
			return nil, err
		}
		readings = append(readings, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return readings, nil
}

func (repo *SensorRepositoryImpl) SelectCountByIDs(
	ctx context.Context,
	db *sqlx.DB,
	ids []IDCombination,
) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}

	idCondition, args := buildIDCondition(ids)
	query := fmt.Sprintf(`
SELECT count(*) AS cnt
FROM sensor_readings
WHERE %s
`, idCondition)

	var result CountResult
	if err := db.GetContext(ctx, &result, query, args...); err != nil {
		return 0, err
	}

	return result.Cnt, nil
}

// Select by ID combinations and time
func (repo *SensorRepositoryImpl) SelectByIDsAndTime(
	ctx context.Context,
	db *sqlx.DB,
	ids []IDCombination,
	startTime, stopTime time.Time,
	limit, offset int,
) ([]model.SensorReading, error) {
	if len(ids) == 0 {
		return []model.SensorReading{}, nil
	}

	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	idCondition, args := buildIDCondition(ids)
	query := fmt.Sprintf(`
SELECT sensor_value, sensor_type, id1, id2, ts
FROM sensor_readings
WHERE %s AND ts >= ? AND ts < ?
ORDER BY ts
LIMIT ? OFFSET ?
`, idCondition)

	args = append(args, startTime, stopTime, limit, offset)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var readings []model.SensorReading
	for rows.Next() {
		var r model.SensorReading
		if err := rows.Scan(&r.SensorValue, &r.SensorType, &r.ID1, &r.ID2, &r.TS); err != nil {
			return nil, err
		}
		readings = append(readings, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return readings, nil
}

func (repo *SensorRepositoryImpl) SelectCountByIDsAndTime(
	ctx context.Context,
	db *sqlx.DB,
	ids []IDCombination,
	startTime, stopTime time.Time,
) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}

	idCondition, args := buildIDCondition(ids)
	query := fmt.Sprintf(`
SELECT count(*) AS cnt
FROM sensor_readings
WHERE %s AND ts >= ? AND ts < ?
`, idCondition)

	args = append(args, startTime, stopTime)

	var result CountResult
	if err := db.GetContext(ctx, &result, query, args...); err != nil {
		return 0, err
	}

	return result.Cnt, nil
}

// Delete operations
func (repo *SensorRepositoryImpl) DeleteByTime(
	ctx context.Context,
	db *sqlx.DB,
	startTime, stopTime time.Time,
) (int64, error) {
	query := `DELETE FROM sensor_readings WHERE ts >= ? AND ts < ?`

	result, err := db.ExecContext(ctx, query, startTime, stopTime)
	if err != nil {
		return 0, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return affected, nil
}

func (repo *SensorRepositoryImpl) DeleteByIDs(
	ctx context.Context,
	db *sqlx.DB,
	ids []IDCombination,
) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}

	idCondition, args := buildIDCondition(ids)
	query := fmt.Sprintf(`DELETE FROM sensor_readings WHERE %s`, idCondition)

	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return affected, nil
}

func (repo *SensorRepositoryImpl) DeleteByIDsAndTime(
	ctx context.Context,
	db *sqlx.DB,
	ids []IDCombination,
	startTime, stopTime time.Time,
) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}

	idCondition, args := buildIDCondition(ids)
	query := fmt.Sprintf(`DELETE FROM sensor_readings WHERE %s AND ts >= ? AND ts < ?`, idCondition)

	args = append(args, startTime, stopTime)

	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return affected, nil
}

// Update operations
func (repo *SensorRepositoryImpl) UpdateByTime(
	ctx context.Context,
	db *sqlx.DB,
	startTime, stopTime time.Time,
	sensorValue float64,
	sensorType string,
) (int64, error) {
	query := `UPDATE sensor_readings SET sensor_value = ?, sensor_type = ? WHERE ts >= ? AND ts < ?`

	result, err := db.ExecContext(ctx, query, sensorValue, sensorType, startTime, stopTime)
	if err != nil {
		return 0, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return affected, nil
}

func (repo *SensorRepositoryImpl) UpdateByIDs(
	ctx context.Context,
	db *sqlx.DB,
	ids []IDCombination,
	sensorValue float64,
	sensorType string,
) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}

	idCondition, args := buildIDCondition(ids)
	query := fmt.Sprintf(`UPDATE sensor_readings SET sensor_value = ?, sensor_type = ? WHERE %s`, idCondition)

	// Prepend the update values to the args
	args = append([]interface{}{sensorValue, sensorType}, args...)

	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return affected, nil
}

func (repo *SensorRepositoryImpl) UpdateByIDsAndTime(
	ctx context.Context,
	db *sqlx.DB,
	ids []IDCombination,
	startTime, stopTime time.Time,
	sensorValue float64,
	sensorType string,
) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}

	idCondition, args := buildIDCondition(ids)
	query := fmt.Sprintf(`UPDATE sensor_readings SET sensor_value = ?, sensor_type = ? WHERE %s AND ts >= ? AND ts < ?`, idCondition)

	// Prepend the update values and append time values
	args = append([]interface{}{sensorValue, sensorType}, args...)
	args = append(args, startTime, stopTime)

	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return affected, nil
}
