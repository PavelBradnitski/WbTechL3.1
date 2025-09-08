package repository

import (
	"context"
	"database/sql"
	"log"

	"github.com/PavelBradnitski/WbTechL3.1/internal/models"
)

type NotificationRepository interface {
	Create(ctx context.Context, n *models.Notification) (string, error)
	GetByID(ctx context.Context, id string) (*models.Notification, error)
	GetAll(ctx context.Context) ([]*models.Notification, error)
	UpdateStatus(ctx context.Context, id string, status models.Status) error
	Cancel(ctx context.Context, id string) error
	ReservePending(ctx context.Context, limit int) ([]*models.Notification, error)
	IncrementRetries(ctx context.Context, id string) error
}

type notificationRepo struct {
	db *sql.DB
}

func NewNotificationRepo(db *sql.DB) NotificationRepository {
	return &notificationRepo{db: db}
}

// Create создает новое уведомление в базе данных.
func (r *notificationRepo) Create(ctx context.Context, req *models.Notification) (string, error) {
	query := `
		INSERT INTO notifications (user_id, email, type, message, subject, scheduled_at, status, retries, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, 'pending', 0, now(), now())
		RETURNING id
	`
	var id string
	err := r.db.QueryRowContext(
		ctx,
		query,
		req.UserID,
		req.Email,
		req.Type,
		req.Message,
		req.Subject,
		req.ScheduledAt,
	).Scan(&id)
	return id, err
}

// GetByID возвращает уведомление по его ID.
func (r *notificationRepo) GetByID(ctx context.Context, id string) (*models.Notification, error) {
	query := `
		SELECT id, user_id, email, type, message, subject, status, scheduled_at, retries, created_at, updated_at
		FROM notifications WHERE id=$1
	`
	var n models.Notification
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&n.ID, &n.UserID, &n.Email, &n.Type, &n.Message, &n.Subject, &n.Status,
		&n.ScheduledAt, &n.Retries, &n.CreatedAt, &n.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *notificationRepo) GetAll(ctx context.Context) ([]*models.Notification, error) {
	query := `
		SELECT id, user_id, email, type, message, subject, status, scheduled_at, retries, created_at, updated_at
		FROM notifications
	`
	var n []*models.Notification
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		log.Printf("error querying notifications: %v", err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var notif models.Notification
		if err := rows.Scan(
			&notif.ID, &notif.UserID, &notif.Email, &notif.Type, &notif.Message, &notif.Subject, &notif.Status,
			&notif.ScheduledAt, &notif.Retries, &notif.CreatedAt, &notif.UpdatedAt,
		); err != nil {
			log.Printf("error scanning notification: %v", err)
			return nil, err
		}
		n = append(n, &notif)
	}

	return n, nil
}

func (r *notificationRepo) Cancel(ctx context.Context, id string) error {
	query := `UPDATE notifications SET status=$1, updated_at=now() WHERE id=$2`
	_, err := r.db.ExecContext(ctx, query, models.StatusCanceled, id)
	return err
}

func (r *notificationRepo) ReservePending(ctx context.Context, limit int) ([]*models.Notification, error) {
	query := `
        WITH cte AS (
            SELECT id
            FROM notifications
            WHERE status = 'pending'
            ORDER BY scheduled_at
            LIMIT $1
            FOR UPDATE SKIP LOCKED
        )
        UPDATE notifications n
        SET status = $2,
            updated_at = now()
        FROM cte
        WHERE n.id = cte.id
        RETURNING n.id, n.user_id, n.email, n.type, n.message, n.subject, n.status, n.scheduled_at, n.retries, n.created_at, n.updated_at;
    `

	rows, err := r.db.QueryContext(ctx, query, limit, models.StatusProcessing)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*models.Notification
	for rows.Next() {
		var n models.Notification
		if err := rows.Scan(
			&n.ID, &n.UserID, &n.Email, &n.Type, &n.Message, &n.Subject, &n.Status,
			&n.ScheduledAt, &n.Retries, &n.CreatedAt, &n.UpdatedAt,
		); err != nil {
			return nil, err
		}
		notifications = append(notifications, &n)
	}
	return notifications, nil
}

func (r *notificationRepo) IncrementRetries(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE notifications SET retries = retries+1, updated_at=now() WHERE id=$1`, id)
	return err
}

func (r *notificationRepo) UpdateStatus(ctx context.Context, id string, status models.Status) error {
	_, err := r.db.ExecContext(ctx, `UPDATE notifications SET status=$1, updated_at=now() WHERE id=$2`, status, id)
	return err
}
