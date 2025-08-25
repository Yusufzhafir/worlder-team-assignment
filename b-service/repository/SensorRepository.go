package repository

import (
	"context"
	"database/sql"

	model "github.com/Yusufzhafir/worlder-team-assignment/b-service/repository/model"

	"github.com/jmoiron/sqlx"
)

const insertReadingSQL = `
INSERT INTO sensor_readings (sensor_value, sensor_type, id1, id2, ts)
VALUES (:sensor_value, :sensor_type, :id1, :id2, :ts)
`

type SensorRepository interface {
	InsertReadingTx(ctx context.Context, db *sqlx.DB, r *model.SensorReadingInsert) (uint64, error)
	InsertReadingsBatchTx(ctx context.Context, db *sqlx.DB, rs []model.SensorReadingInsert) (int64, error)
}

type SensorRepositoryImpl struct {
}

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
