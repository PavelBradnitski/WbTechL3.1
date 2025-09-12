package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/PavelBradnitski/WbTechL3.1/internal/models"
	"github.com/google/uuid"
)

// NotificationRepository определяет методы для работы с уведомлениями в базе данных.
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

// NewNotificationRepo создает новый экземпляр NotificationRepository.
func NewNotificationRepo(db *sql.DB) NotificationRepository {
	return &notificationRepo{db: db}
}

// Create создает новое уведомление в базе данных.
func (r *notificationRepo) Create(ctx context.Context, req *models.Notification) (string, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("error starting transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()
	// 1. Вставляем данные в таблицу notifications
	notificationQuery := `
  INSERT INTO notifications (type, status, scheduled_at, retries)
  VALUES ($1, $2, $3, $4)
  RETURNING id
 `
	var notificationID string
	err = tx.QueryRowContext(ctx, notificationQuery, req.Type, req.Status, req.ScheduledAt, req.Retries).Scan(&notificationID)
	if err != nil {
		return "", fmt.Errorf("error inserting into notifications: %w", err)
	}

	// 2. В зависимости от типа уведомления, вставляем данные в соответствующую таблицу
	switch req.Type {
	case "email":
		emailID := uuid.New().String()
		emailQuery := `
   INSERT INTO email_notifications (id, notification_id, email, subject, message)
   VALUES ($1, $2, $3, $4, $5)
  `
		_, err = tx.ExecContext(ctx, emailQuery, emailID, notificationID, req.EmailNotification.Email, req.EmailNotification.Subject, req.EmailNotification.Message)
		if err != nil {
			return "", fmt.Errorf("error inserting into email_notifications: %w", err)
		}
	case "telegram":
		telegramID := uuid.New().String()
		telegramQuery := `
   INSERT INTO telegram_notifications (id, notification_id, chat_id, message)
   VALUES ($1, $2, $3, $4)
  `
		_, err = tx.ExecContext(ctx, telegramQuery, telegramID, notificationID, req.TelegramNotification.ChatID, req.TelegramNotification.Message)
		if err != nil {
			return "", fmt.Errorf("error inserting into telegram_notifications: %w", err)
		}
	default:
		return "", fmt.Errorf("unknown notification type: %s", req.Type)
	}

	// Фиксируем транзакцию
	err = tx.Commit()
	if err != nil {
		return "", fmt.Errorf("error committing transaction: %w", err)
	}

	return notificationID, nil
}

// GetByID возвращает уведомление по его ID.
func (r *notificationRepo) GetByID(ctx context.Context, id string) (*models.Notification, error) {
	// 1. Получаем данные из таблицы notifications
	notificationQuery := `
        SELECT id, type, status, scheduled_at, retries
        FROM notifications
        WHERE id = $1
    `
	var n models.Notification
	err := r.db.QueryRowContext(ctx, notificationQuery, id).Scan(
		&n.ID, &n.Type, &n.Status, &n.ScheduledAt, &n.Retries,
	)
	if err != nil {
		return nil, fmt.Errorf("error getting notification by id: %w", err)
	}

	// 2. В зависимости от типа уведомления, получаем дополнительные данные из соответствующей таблицы
	switch n.Type {
	case "email":
		n.EmailNotification = &models.EmailNotification{}
		emailQuery := `
            SELECT email, subject, message
            FROM email_notifications
            WHERE notification_id = $1
        `
		err = r.db.QueryRowContext(ctx, emailQuery, id).Scan(
			&n.EmailNotification.Email, &n.EmailNotification.Subject, &n.EmailNotification.Message,
		)
		if err != nil {
			return nil, fmt.Errorf("error getting email notification details: %w", err)
		}
	case "telegram":
		n.TelegramNotification = &models.TelegramNotification{}
		telegramQuery := `
            SELECT chat_id, message
            FROM telegram_notifications
            WHERE notification_id = $1
        `
		err = r.db.QueryRowContext(ctx, telegramQuery, id).Scan(
			&n.TelegramNotification.ChatID, &n.TelegramNotification.Message,
		)
		if err != nil {
			return nil, fmt.Errorf("error getting telegram notification details: %w", err)
		}
	default:
		return nil, fmt.Errorf("unknown notification type: %s", n.Type)
	}

	return &n, nil
}

func (r *notificationRepo) GetAll(ctx context.Context) ([]*models.Notification, error) {
	notifications := []*models.Notification{}

	// 1. Получаем базовую информацию из таблицы notifications
	notificationQuery := `
        SELECT id, type, status, scheduled_at, retries
        FROM notifications
    `
	rows, err := r.db.QueryContext(ctx, notificationQuery)
	if err != nil {
		return nil, fmt.Errorf("error querying notifications: %w", err)
	}
	defer rows.Close()

	// Итерируемся по результатам запроса notifications
	for rows.Next() {
		var notif models.Notification
		err := rows.Scan(
			&notif.ID, &notif.Type, &notif.Status, &notif.ScheduledAt, &notif.Retries,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning notification: %w", err)
		}

		// 2. Получаем дополнительную информацию в зависимости от типа уведомления
		switch notif.Type {
		case "email":
			notif.EmailNotification = &models.EmailNotification{}
			emailQuery := `
                SELECT email, subject, message
                FROM email_notifications
                WHERE notification_id = $1
            `
			err = r.db.QueryRowContext(ctx, emailQuery, notif.ID).Scan(
				&notif.EmailNotification.Email, &notif.EmailNotification.Subject, &notif.EmailNotification.Message,
			)
			if err != nil {
				return nil, fmt.Errorf("error getting email notification details: %w", err)
			}
		case "telegram":
			notif.TelegramNotification = &models.TelegramNotification{}
			telegramQuery := `
                SELECT chat_id, message
                FROM telegram_notifications
                WHERE notification_id = $1
            `
			err = r.db.QueryRowContext(ctx, telegramQuery, notif.ID).Scan(
				&notif.TelegramNotification.ChatID, &notif.TelegramNotification.Message,
			)
			if err != nil {
				return nil, fmt.Errorf("error getting telegram notification details: %w", err)
			}
		default:
			log.Printf("Unknown notification type: %s", notif.Type)
		}

		notifications = append(notifications, &notif)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating through rows: %w", err)
	}

	return notifications, nil
}

func (r *notificationRepo) Cancel(ctx context.Context, id string) error {
	query := `UPDATE notifications SET status=$1, updated_at=now() WHERE id=$2`
	_, err := r.db.ExecContext(ctx, query, models.StatusCanceled, id)
	return err
}

func (r *notificationRepo) ReservePending(ctx context.Context, limit int) ([]*models.Notification, error) {
	query := `
  WITH selected_notifications AS (
   SELECT id
   FROM notifications
   WHERE status = $1 
   ORDER BY scheduled_at
   LIMIT $2
   FOR UPDATE SKIP LOCKED
  )
  UPDATE notifications
  SET status = $3,  
   updated_at = NOW()
  WHERE id IN (SELECT id FROM selected_notifications)
  RETURNING id, type, status, scheduled_at, retries, created_at, updated_at;
 `

	rows, err := r.db.QueryContext(ctx, query, models.StatusScheduled, limit, models.StatusProcessing)
	if err != nil {
		return nil, fmt.Errorf("failed to query for pending notifications: %w", err)
	}
	defer rows.Close()

	notifications := []*models.Notification{}
	for rows.Next() {
		n := &models.Notification{}
		if err := rows.Scan(
			&n.ID, &n.Type, &n.Status, &n.ScheduledAt, &n.Retries, &n.CreatedAt, &n.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}

		// Fetch additional details based on notification type.
		switch n.Type {
		case "email":
			emailNotification, err := r.getEmailNotificationDetails(ctx, n.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to get email notification details: %w", err)
			}
			log.Println("Email notification details:", emailNotification)
			n.EmailNotification = emailNotification
		case "telegram":
			telegramNotification, err := r.getTelegramNotificationDetails(ctx, n.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to get telegram notification details: %w", err)
			}
			log.Println("Telegram notification details:", telegramNotification)
			n.TelegramNotification = telegramNotification
		default:
			return nil, fmt.Errorf("unknown notification type: %s", n.Type)
		}

		notifications = append(notifications, n)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating through rows: %w", err)
	}

	return notifications, nil
}

func (r *notificationRepo) getEmailNotificationDetails(ctx context.Context, notificationID string) (*models.EmailNotification, error) {
	query := `
  SELECT id, notification_id, email, subject, message
  FROM email_notifications
  WHERE notification_id = $1;
 `

	row := r.db.QueryRowContext(ctx, query, notificationID)
	emailNotification := &models.EmailNotification{}
	err := row.Scan(&emailNotification.ID, &emailNotification.NotificationID, &emailNotification.Email, &emailNotification.Subject, &emailNotification.Message)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No email details found is not an error. Can happen if the type is not email.
		}
		return nil, fmt.Errorf("failed to scan email notification details: %w", err)
	}

	return emailNotification, nil
}

func (r *notificationRepo) getTelegramNotificationDetails(ctx context.Context, notificationID string) (*models.TelegramNotification, error) {
	query := `
  SELECT id, notification_id, chat_id, message
  FROM telegram_notifications
  WHERE notification_id = $1;
 `

	row := r.db.QueryRowContext(ctx, query, notificationID)
	telegramNotification := &models.TelegramNotification{}
	err := row.Scan(&telegramNotification.ID, &telegramNotification.NotificationID, &telegramNotification.ChatID, &telegramNotification.Message)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No telegram details found.
		}
		return nil, fmt.Errorf("failed to scan telegram notification details: %w", err)
	}

	return telegramNotification, nil
}

func (r *notificationRepo) IncrementRetries(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE notifications SET retries = retries+1, updated_at=now() WHERE id=$1`, id)
	return err
}

func (r *notificationRepo) UpdateStatus(ctx context.Context, id string, status models.Status) error {
	_, err := r.db.ExecContext(ctx, `UPDATE notifications SET status=$1, updated_at=now() WHERE id=$2`, status, id)
	return err
}
