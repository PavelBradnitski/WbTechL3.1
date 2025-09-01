package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/PavelBradnitski/WbTechL3.1/internal/models"
)

type NotificationRepository interface {
	Create(ctx context.Context, n *models.Notification) (string, error)
	GetByID(ctx context.Context, id string) (*models.Notification, error)
	UpdateStatus(ctx context.Context, id, status string) error
	Cancel(ctx context.Context, id string) error
	FindReady(ctx context.Context, t time.Time, limit int) ([]*models.Notification, error)
	Reschedule(ctx context.Context, id string, next time.Time, retries int) error
	GetPending(ctx context.Context, limit int) ([]*models.Notification, error)
	IncrementRetries(ctx context.Context, id string) error
}

type notificationRepo struct {
	db *sql.DB
}

func NewNotificationRepo(db *sql.DB) NotificationRepository {
	return &notificationRepo{db: db}
}

// Create возвращает только ID новой записи
func (r *notificationRepo) Create(ctx context.Context, req *models.Notification) (string, error) {
	query := `
		INSERT INTO notifications (user_id, message, scheduled_at, status, retries, created_at, updated_at)
		VALUES ($1, $2, $3, 'pending', 0, now(), now())
		RETURNING id
	`
	var id string
	err := r.db.QueryRowContext(ctx, query, req.UserID, req.Message, req.ScheduledAt).Scan(&id)
	return id, err
}

func (r *notificationRepo) GetByID(ctx context.Context, id string) (*models.Notification, error) {
	query := `SELECT id,user_id, message, status, scheduled_at, retries, created_at, updated_at
	          FROM notifications WHERE id=$1`
	var n models.Notification
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&n.ID, &n.UserID, &n.Message, &n.Status, &n.ScheduledAt,
		&n.Retries, &n.CreatedAt, &n.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *notificationRepo) Cancel(ctx context.Context, id string) error {
	query := `UPDATE notifications SET status='canceled', updated_at=now() WHERE id=$1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *notificationRepo) GetPending(ctx context.Context, limit int) ([]*models.Notification, error) {
	query := `
		SELECT id, user_id, message, scheduled_at, status, retries, created_at, updated_at
		FROM notifications
		WHERE status='pending' AND scheduled_at <= now()
		ORDER BY scheduled_at ASC
		LIMIT $1
	`
	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*models.Notification
	for rows.Next() {
		var n models.Notification
		err := rows.Scan(
			&n.ID, &n.UserID, &n.Message, &n.ScheduledAt,
			&n.Status, &n.Retries, &n.CreatedAt, &n.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, &n)
	}
	return result, nil
}

func (r *notificationRepo) IncrementRetries(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE notifications SET retries = retries+1, updated_at=now() WHERE id=$1`, id)
	return err
}

func (r *notificationRepo) UpdateStatus(ctx context.Context, id, status string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE notifications SET status=$1, updated_at=now() WHERE id=$2`, status, id)
	return err
}

func (r *notificationRepo) FindReady(ctx context.Context, t time.Time, limit int) ([]*models.Notification, error) {
	query := `SELECT id, message, send_at, status, retries, created_at, updated_at
	          FROM notifications
	          WHERE send_at <= $1 AND status='scheduled'
	          ORDER BY send_at ASC
	          LIMIT $2`
	rows, err := r.db.QueryContext(ctx, query, t, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*models.Notification
	for rows.Next() {
		n := &models.Notification{}
		if err := rows.Scan(&n.ID, &n.Message, &n.ScheduledAt, &n.Status, &n.Retries, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, n)
	}
	return list, nil
}

func (r *notificationRepo) Reschedule(ctx context.Context, id string, next time.Time, retries int) error {
	query := `UPDATE notifications
	          SET status='scheduled', retries=$2, send_at=$3, updated_at=now()
	          WHERE id=$1`
	_, err := r.db.ExecContext(ctx, query, id, retries, next)
	return err
}
