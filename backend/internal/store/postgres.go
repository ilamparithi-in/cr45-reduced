package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/CR45-NITT/cr45-reduced/backend/internal/domain"
)

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

func (s *PostgresStore) GetResolvedTimetable(ctx context.Context, classID string) (domain.Timetable, error) {
	const query = `
SELECT ds.slot_index,
       COALESCE(o.course_code, ds.course_code) AS course_code,
       COALESCE(o.start_time, ds.start_time) AS start_time,
       COALESCE(o.end_time, ds.end_time) AS end_time,
       COALESCE(o.venue, ds.venue) AS venue,
       COALESCE(o.status, ds.status) AS status
FROM default_slots ds
LEFT JOIN overrides o
  ON o.class_id = ds.class_id AND o.slot_index = ds.slot_index
WHERE ds.class_id = $1
ORDER BY ds.slot_index
`

	rows, err := s.db.QueryContext(ctx, query, classID)
	if err != nil {
		return domain.Timetable{}, err
	}
	defer rows.Close()

	var slots []domain.Slot
	for rows.Next() {
		var slot domain.Slot
		if err := rows.Scan(&slot.SlotIndex, &slot.CourseCode, &slot.StartTime, &slot.EndTime, &slot.Venue, &slot.Status); err != nil {
			return domain.Timetable{}, err
		}
		slots = append(slots, slot)
	}
	if err := rows.Err(); err != nil {
		return domain.Timetable{}, err
	}
	if len(slots) == 0 {
		return domain.Timetable{}, ErrClassNotFound
	}

	return domain.Timetable{
		ClassID: classID,
		Date:    time.Now().Format("2006-01-02"),
		Slots:   slots,
	}, nil
}

func (s *PostgresStore) UpsertOverride(ctx context.Context, req domain.UpdateOverrideRequest) error {
	if req.SlotIndex <= 0 {
		return ErrInvalidSlotIdx
	}

	var maxSlot sql.NullInt64
	if err := s.db.QueryRowContext(ctx, `SELECT MAX(slot_index) FROM default_slots WHERE class_id = $1`, req.ClassID).Scan(&maxSlot); err != nil {
		return ErrClassNotFound
	}
	if !maxSlot.Valid {
		return ErrClassNotFound
	}
	if req.SlotIndex > int(maxSlot.Int64)+1 {
		return ErrInvalidSlotIdx
	}

	if req.SlotIndex == int(maxSlot.Int64)+1 {
		const insertQuery = `
INSERT INTO default_slots (class_id, slot_index, course_code, start_time, end_time, venue, status)
VALUES ($1, $2, $3, $4, $5, $6, $7)
`
		_, err := s.db.ExecContext(ctx, insertQuery, req.ClassID, req.SlotIndex, req.CourseCode, req.StartTime, req.EndTime, req.Venue, req.Status)
		return err
	}

	const query = `
INSERT INTO overrides (class_id, slot_index, course_code, start_time, end_time, venue, status, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
ON CONFLICT (class_id, slot_index)
DO UPDATE SET
  course_code = EXCLUDED.course_code,
  start_time = EXCLUDED.start_time,
  end_time = EXCLUDED.end_time,
  venue = EXCLUDED.venue,
  status = EXCLUDED.status,
  updated_at = NOW()
`

	_, err := s.db.ExecContext(ctx, query, req.ClassID, req.SlotIndex, req.CourseCode, req.StartTime, req.EndTime, req.Venue, req.Status)
	return err
}

func (s *PostgresStore) DeleteSlot(ctx context.Context, req domain.DeleteSlotRequest) error {
	if req.SlotIndex <= 0 {
		return ErrInvalidSlotIdx
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	var exists bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM default_slots WHERE class_id = $1 AND slot_index = $2)`, req.ClassID, req.SlotIndex).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return ErrClassNotFound
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM overrides WHERE class_id = $1 AND slot_index = $2`, req.ClassID, req.SlotIndex); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM default_slots WHERE class_id = $1 AND slot_index = $2`, req.ClassID, req.SlotIndex); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `UPDATE overrides SET slot_index = -slot_index WHERE class_id = $1 AND slot_index > $2`, req.ClassID, req.SlotIndex); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE default_slots SET slot_index = -slot_index WHERE class_id = $1 AND slot_index > $2`, req.ClassID, req.SlotIndex); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE overrides SET slot_index = -slot_index - 1 WHERE class_id = $1 AND slot_index < 0`, req.ClassID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE default_slots SET slot_index = -slot_index - 1 WHERE class_id = $1 AND slot_index < 0`, req.ClassID); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (s *PostgresStore) GetPasswordHash(ctx context.Context, username string) (string, error) {
	var hash string
	err := s.db.QueryRowContext(ctx, `SELECT password_hash FROM app_users WHERE username = $1`, username).Scan(&hash)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrInvalidLogin
	}
	if err != nil {
		return "", err
	}
	return hash, nil
}
