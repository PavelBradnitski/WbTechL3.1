package repository

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"github.com/PavelBradnitski/WbTechL3.1/internal/models"
)

// helper для инициализации моков
func newTestRepo(t *testing.T) (NotificationRepository, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}

	repo := NewNotificationRepo(db)

	cleanup := func() {
		db.Close()
	}

	return repo, mock, cleanup
}

func TestNotificationRepo_Create(t *testing.T) {
	t.Run("Success_Email", func(t *testing.T) {
		repo, mock, cleanup := newTestRepo(t)
		defer cleanup()

		req := &models.Notification{
			Type:        models.NotificationType("email"),
			Status:      models.Status("scheduled"),
			ScheduledAt: time.Now().Add(time.Hour),
			Retries:     0, // Добавлено поле Retries
			EmailNotification: &models.EmailNotification{
				Email:   "test@example.com",
				Subject: "Test subject",
				Message: "Test message",
			},
		}

		expectedNotificationID := "notif-123"

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`
			INSERT INTO notifications (type, status, scheduled_at, retries)
			VALUES ($1, $2, $3, $4)
			RETURNING id
		`)).WithArgs(req.Type, req.Status, req.ScheduledAt, req.Retries).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(expectedNotificationID))
		mock.ExpectExec(regexp.QuoteMeta(`
			INSERT INTO email_notifications (id, notification_id, email, subject, message)
			VALUES ($1, $2, $3, $4, $5)
		`)).WithArgs(sqlmock.AnyArg(), expectedNotificationID, req.EmailNotification.Email, req.EmailNotification.Subject, req.EmailNotification.Message).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		id, err := repo.Create(context.Background(), req)
		assert.NoError(t, err)
		assert.Equal(t, expectedNotificationID, id)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
	t.Run("Success_Telegram", func(t *testing.T) {
		repo, mock, cleanup := newTestRepo(t)
		defer cleanup()

		req := &models.Notification{
			Type:        models.NotificationType("telegram"),
			Status:      models.Status("scheduled"),
			ScheduledAt: time.Now().Add(time.Hour),
			Retries:     0, // Добавлено поле Retries
			TelegramNotification: &models.TelegramNotification{
				ChatID:  "123456789",
				Message: "Test message",
			},
		}

		expectedNotificationID := "notif-456"

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`
			INSERT INTO notifications (type, status, scheduled_at, retries)
			VALUES ($1, $2, $3, $4)
			RETURNING id
		`)).WithArgs(req.Type, req.Status, req.ScheduledAt, req.Retries).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(expectedNotificationID))
		mock.ExpectExec(regexp.QuoteMeta(`
			INSERT INTO telegram_notifications (id, notification_id, chat_id, message)
			VALUES ($1, $2, $3, $4)
		`)).WithArgs(sqlmock.AnyArg(), expectedNotificationID, req.TelegramNotification.ChatID, req.TelegramNotification.Message).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		id, err := repo.Create(context.Background(), req)
		assert.NoError(t, err)
		assert.Equal(t, expectedNotificationID, id)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error_BeginTransaction", func(t *testing.T) {
		repo, mock, cleanup := newTestRepo(t)
		defer cleanup()

		req := &models.Notification{
			Type:        models.NotificationType("email"),
			Status:      models.Status("scheduled"),
			ScheduledAt: time.Now().Add(time.Hour),
			Retries:     0,
			EmailNotification: &models.EmailNotification{
				Email:   "test@example.com",
				Subject: "Test subject",
				Message: "Test message",
			},
		}

		mock.ExpectBegin().WillReturnError(fmt.Errorf("transaction error"))

		_, err := repo.Create(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error starting transaction")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error_InsertNotifications", func(t *testing.T) {
		repo, mock, cleanup := newTestRepo(t)
		defer cleanup()

		req := &models.Notification{
			Type:        models.NotificationType("email"),
			Status:      models.Status("scheduled"),
			ScheduledAt: time.Now().Add(time.Hour),
			Retries:     0,
			EmailNotification: &models.EmailNotification{
				Email:   "test@example.com",
				Subject: "Test subject",
				Message: "Test message",
			},
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`
			INSERT INTO notifications (type, status, scheduled_at, retries)
			VALUES ($1, $2, $3, $4)
			RETURNING id
		`)).WithArgs(req.Type, req.Status, req.ScheduledAt, req.Retries).
			WillReturnError(fmt.Errorf("insert error"))
		mock.ExpectRollback()

		_, err := repo.Create(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error inserting into notifications")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error_InsertEmailNotifications", func(t *testing.T) {
		repo, mock, cleanup := newTestRepo(t)
		defer cleanup()

		req := &models.Notification{
			Type:        models.NotificationType("email"),
			Status:      models.Status("scheduled"),
			ScheduledAt: time.Now().Add(time.Hour),
			Retries:     0,
			EmailNotification: &models.EmailNotification{
				Email:   "test@example.com",
				Subject: "Test subject",
				Message: "Test message",
			},
		}

		expectedNotificationID := "notif-123"

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`
			INSERT INTO notifications (type, status, scheduled_at, retries)
			VALUES ($1, $2, $3, $4)
			RETURNING id
		`)).WithArgs(req.Type, req.Status, req.ScheduledAt, req.Retries).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(expectedNotificationID))
		mock.ExpectExec(regexp.QuoteMeta(`
			INSERT INTO email_notifications (id, notification_id, email, subject, message)
			VALUES ($1, $2, $3, $4, $5)
		`)).WithArgs(sqlmock.AnyArg(), expectedNotificationID, req.EmailNotification.Email, req.EmailNotification.Subject, req.EmailNotification.Message).
			WillReturnError(fmt.Errorf("email insert error"))
		mock.ExpectRollback()

		_, err := repo.Create(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error inserting into email_notifications")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestNotificationRepo_GetByID(t *testing.T) {
	t.Run("Success_Email", func(t *testing.T) {
		repo, mock, cleanup := newTestRepo(t)
		defer cleanup()

		notificationID := "notif-123"

		// Ожидаемые значения из БД
		expectedNotification := &models.Notification{
			ID:          notificationID,
			Type:        "email",
			Status:      "scheduled",
			ScheduledAt: time.Now().Add(time.Hour),
			Retries:     0,
			EmailNotification: &models.EmailNotification{
				Email:   "test@example.com",
				Subject: "Test Subject",
				Message: "Test Message",
			},
		}
		scheduledAt := expectedNotification.ScheduledAt

		// Описываем ожидаемые запросы и возвращаемые значения
		mock.ExpectQuery(regexp.QuoteMeta(`
			SELECT id, type, status, scheduled_at, retries
			FROM notifications
			WHERE id = $1
		`)).WithArgs(notificationID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "type", "status", "scheduled_at", "retries"}).
				AddRow(expectedNotification.ID, expectedNotification.Type, expectedNotification.Status, scheduledAt, expectedNotification.Retries))

		mock.ExpectQuery(regexp.QuoteMeta(`
			SELECT email, subject, message
			FROM email_notifications
			WHERE notification_id = $1
		`)).WithArgs(notificationID).
			WillReturnRows(sqlmock.NewRows([]string{"email", "subject", "message"}).
				AddRow(expectedNotification.EmailNotification.Email, expectedNotification.EmailNotification.Subject, expectedNotification.EmailNotification.Message))

		// Вызываем тестируемую функцию
		notification, err := repo.GetByID(context.Background(), notificationID)

		// Проверяем результаты
		assert.NoError(t, err)
		assert.Equal(t, expectedNotification, notification)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Success_Telegram", func(t *testing.T) {
		repo, mock, cleanup := newTestRepo(t)
		defer cleanup()

		notificationID := "notif-456"

		// Ожидаемые значения из БД
		expectedNotification := &models.Notification{
			ID:          notificationID,
			Type:        "telegram",
			Status:      "scheduled",
			ScheduledAt: time.Now().Add(time.Hour),
			Retries:     0,
			TelegramNotification: &models.TelegramNotification{
				ChatID:  "123456789",
				Message: "Telegram Message",
			},
		}
		scheduledAt := expectedNotification.ScheduledAt

		// Описываем ожидаемые запросы и возвращаемые значения
		mock.ExpectQuery(regexp.QuoteMeta(`
			SELECT id, type, status, scheduled_at, retries
			FROM notifications
			WHERE id = $1
		`)).WithArgs(notificationID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "type", "status", "scheduled_at", "retries"}).
				AddRow(expectedNotification.ID, expectedNotification.Type, expectedNotification.Status, scheduledAt, expectedNotification.Retries))

		mock.ExpectQuery(regexp.QuoteMeta(`
			SELECT chat_id, message
			FROM telegram_notifications
			WHERE notification_id = $1
		`)).WithArgs(notificationID).
			WillReturnRows(sqlmock.NewRows([]string{"chat_id", "message"}).
				AddRow(expectedNotification.TelegramNotification.ChatID, expectedNotification.TelegramNotification.Message))

		// Вызываем тестируемую функцию
		notification, err := repo.GetByID(context.Background(), notificationID)

		// Проверяем результаты
		assert.NoError(t, err)
		assert.Equal(t, expectedNotification, notification)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("NotFound", func(t *testing.T) {
		repo, mock, cleanup := newTestRepo(t)
		defer cleanup()

		notificationID := "non-existent-id"

		// Ожидаем, что QueryRowContext вернет sql.ErrNoRows
		mock.ExpectQuery(regexp.QuoteMeta(`
			SELECT id, type, status, scheduled_at, retries
			FROM notifications
			WHERE id = $1
		`)).WithArgs(notificationID).
			WillReturnError(sql.ErrNoRows)

		// Вызываем тестируемую функцию
		notification, err := repo.GetByID(context.Background(), notificationID)

		// Проверяем результаты
		assert.Error(t, err)
		assert.Nil(t, notification)
		assert.ErrorContains(t, err, "error getting notification by id") // Проверяем сообщение об ошибке
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DBError", func(t *testing.T) {
		repo, mock, cleanup := newTestRepo(t)
		defer cleanup()

		notificationID := "some-id"

		// Мокируем ошибку БД
		mock.ExpectQuery(regexp.QuoteMeta(`
			SELECT id, type, status, scheduled_at, retries
			FROM notifications
			WHERE id = $1
		`)).WithArgs(notificationID).
			WillReturnError(fmt.Errorf("database connection error"))

		// Вызываем тестируемую функцию
		notification, err := repo.GetByID(context.Background(), notificationID)

		// Проверяем результаты
		assert.Error(t, err)
		assert.Nil(t, notification)
		assert.ErrorContains(t, err, "error getting notification by id") // Проверяем сообщение об ошибке
		assert.ErrorContains(t, err, "database connection error")        // Проверяем вложенную ошибку
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("UnknownType", func(t *testing.T) {
		repo, mock, cleanup := newTestRepo(t)
		defer cleanup()

		notificationID := "unknown-type-id"

		// Мокируем возврат уведомления с неизвестным типом
		mock.ExpectQuery(regexp.QuoteMeta(`
			SELECT id, type, status, scheduled_at, retries
			FROM notifications
			WHERE id = $1
		`)).WithArgs(notificationID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "type", "status", "scheduled_at", "retries"}).
				AddRow(notificationID, "unknown", "scheduled", time.Now(), 0)) // "unknown" тип

		// Вызываем тестируемую функцию
		notification, err := repo.GetByID(context.Background(), notificationID)

		// Проверяем результаты
		assert.Error(t, err)
		assert.Nil(t, notification)
		assert.ErrorContains(t, err, "unknown notification type") // Проверяем сообщение об ошибке
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestNotificationRepo_GetAll(t *testing.T) {
	t.Run("Success_MultipleNotifications", func(t *testing.T) {
		repo, mock, cleanup := newTestRepo(t)
		defer cleanup()

		// Ожидаемые уведомления
		expectedNotifications := []*models.Notification{
			{
				ID:          "email-1",
				Type:        "email",
				Status:      "scheduled",
				ScheduledAt: time.Now().Add(time.Hour),
				Retries:     0,
				EmailNotification: &models.EmailNotification{
					Email:   "test1@example.com",
					Subject: "Subject 1",
					Message: "Message 1",
				},
			},
			{
				ID:          "telegram-1",
				Type:        "telegram",
				Status:      "pending",
				ScheduledAt: time.Now().Add(2 * time.Hour),
				Retries:     1,
				TelegramNotification: &models.TelegramNotification{
					ChatID:  "12345",
					Message: "Telegram Message 1",
				},
			},
		}

		// Мокируем первый запрос (получение базовой информации)
		rows := sqlmock.NewRows([]string{"id", "type", "status", "scheduled_at", "retries"})
		scheduledAtEmail := expectedNotifications[0].ScheduledAt
		scheduledAtTelegram := expectedNotifications[1].ScheduledAt

		rows.AddRow(expectedNotifications[0].ID, expectedNotifications[0].Type, expectedNotifications[0].Status, scheduledAtEmail, expectedNotifications[0].Retries)
		rows.AddRow(expectedNotifications[1].ID, expectedNotifications[1].Type, expectedNotifications[1].Status, scheduledAtTelegram, expectedNotifications[1].Retries)

		mock.ExpectQuery(regexp.QuoteMeta(`
			SELECT id, type, status, scheduled_at, retries
			FROM notifications
		`)).WillReturnRows(rows)

		// Мокируем запросы для email уведомления
		mock.ExpectQuery(regexp.QuoteMeta(`
			SELECT email, subject, message
			FROM email_notifications
			WHERE notification_id = $1
		`)).WithArgs(expectedNotifications[0].ID).
			WillReturnRows(sqlmock.NewRows([]string{"email", "subject", "message"}).
				AddRow(expectedNotifications[0].EmailNotification.Email, expectedNotifications[0].EmailNotification.Subject, expectedNotifications[0].EmailNotification.Message))

		// Мокируем запросы для telegram уведомления
		mock.ExpectQuery(regexp.QuoteMeta(`
			SELECT chat_id, message
			FROM telegram_notifications
			WHERE notification_id = $1
		`)).WithArgs(expectedNotifications[1].ID).
			WillReturnRows(sqlmock.NewRows([]string{"chat_id", "message"}).
				AddRow(expectedNotifications[1].TelegramNotification.ChatID, expectedNotifications[1].TelegramNotification.Message))

		// Вызываем тестируемую функцию
		notifications, err := repo.GetAll(context.Background())

		// Проверяем результаты
		assert.NoError(t, err)
		assert.Equal(t, expectedNotifications, notifications)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Success_NoNotifications", func(t *testing.T) {
		repo, mock, cleanup := newTestRepo(t)
		defer cleanup()

		// Мокируем запрос, возвращающий пустой результат
		mock.ExpectQuery(regexp.QuoteMeta(`
			SELECT id, type, status, scheduled_at, retries
			FROM notifications
		`)).WillReturnRows(sqlmock.NewRows([]string{"id", "type", "status", "scheduled_at", "retries"}))

		// Вызываем тестируемую функцию
		notifications, err := repo.GetAll(context.Background())

		// Проверяем результаты
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrNotFound)
		assert.Nil(t, notifications)
		assert.NoError(t, mock.ExpectationsWereMet())

	})

	t.Run("DBError_QueryNotifications", func(t *testing.T) {
		repo, mock, cleanup := newTestRepo(t)
		defer cleanup()

		// Мокируем ошибку при запросе notifications
		mock.ExpectQuery(regexp.QuoteMeta(`
			SELECT id, type, status, scheduled_at, retries
			FROM notifications
		`)).WillReturnError(fmt.Errorf("database connection error"))

		// Вызываем тестируемую функцию
		notifications, err := repo.GetAll(context.Background())

		// Проверяем результаты
		assert.Error(t, err)
		assert.Nil(t, notifications)
		assert.ErrorContains(t, err, "error querying notifications")
		assert.ErrorContains(t, err, "database connection error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DBError_QueryEmailDetails", func(t *testing.T) {
		repo, mock, cleanup := newTestRepo(t)
		defer cleanup()

		// Мокируем основной запрос
		rows := sqlmock.NewRows([]string{"id", "type", "status", "scheduled_at", "retries"})
		rows.AddRow("email-1", "email", "scheduled", time.Now(), 0) // Тип "email"
		mock.ExpectQuery(regexp.QuoteMeta(`
			SELECT id, type, status, scheduled_at, retries
			FROM notifications
		`)).WillReturnRows(rows)

		// Мокируем ошибку при запросе деталей email
		mock.ExpectQuery(regexp.QuoteMeta(`
			SELECT email, subject, message
			FROM email_notifications
			WHERE notification_id = $1
		`)).WithArgs("email-1").WillReturnError(fmt.Errorf("database connection error"))

		// Вызываем тестируемую функцию
		notifications, err := repo.GetAll(context.Background())

		// Проверяем результаты
		assert.Error(t, err)
		assert.Nil(t, notifications)
		assert.ErrorContains(t, err, "error getting email notification details")
		assert.ErrorContains(t, err, "database connection error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DBError_ScanNotifications", func(t *testing.T) {
		repo, mock, cleanup := newTestRepo(t)
		defer cleanup()

		// Мокируем основной запрос, но с ошибкой при сканировании
		rows := sqlmock.NewRows([]string{"id", "type", "status", "scheduled_at", "retries"}).AddRow("id", "type", "status", "not-a-time", 0) // Некорректный тип для scheduled_at
		mock.ExpectQuery(regexp.QuoteMeta(`
			SELECT id, type, status, scheduled_at, retries
			FROM notifications
		`)).WillReturnRows(rows)

		// Вызываем тестируемую функцию
		notifications, err := repo.GetAll(context.Background())

		// Проверяем результаты
		assert.Error(t, err)
		assert.Nil(t, notifications)
		assert.ErrorContains(t, err, "error scanning notification")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("UnknownNotificationType", func(t *testing.T) {
		repo, mock, cleanup := newTestRepo(t)
		defer cleanup()

		// Мокируем основной запрос
		rows := sqlmock.NewRows([]string{"id", "type", "status", "scheduled_at", "retries"})
		rows.AddRow("unknown-1", "unknown", "scheduled", time.Now(), 0) // Неизвестный тип
		mock.ExpectQuery(regexp.QuoteMeta(`
			SELECT id, type, status, scheduled_at, retries
			FROM notifications
		`)).WillReturnRows(rows)

		// Вызываем тестируемую функцию
		notifications, err := repo.GetAll(context.Background())

		// Проверяем результаты
		assert.NoError(t, err)                            //  Функция не должна возвращать ошибку, она логирует
		assert.Len(t, notifications, 1)                   // Возвращает уведомление с заполненными базовыми полями
		assert.Nil(t, notifications[0].EmailNotification) // Убедимся, что дополнительные поля не заполнены
		assert.Nil(t, notifications[0].TelegramNotification)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestNotificationRepo_Cancel(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		repo, mock, cleanup := newTestRepo(t)
		defer cleanup()

		notificationID := "notif-123"

		// Ожидаем, что ExecContext будет вызван с правильными аргументами
		mock.ExpectExec(regexp.QuoteMeta(`
   UPDATE notifications SET status=$1, updated_at=now() WHERE id=$2
  `)).WithArgs(models.StatusCanceled, notificationID).
			WillReturnResult(sqlmock.NewResult(1, 1)) // 1 row affected

		// Вызываем тестируемую функцию
		err := repo.Cancel(context.Background(), notificationID)

		// Проверяем результаты
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("NotFound", func(t *testing.T) {
		repo, mock, cleanup := newTestRepo(t)
		defer cleanup()

		notificationID := "non-existent-id"

		// Ожидаем, что ExecContext будет вызван, но не затронет ни одной строки
		mock.ExpectExec(regexp.QuoteMeta(`
   UPDATE notifications SET status=$1, updated_at=now() WHERE id=$2
  `)).WithArgs(models.StatusCanceled, notificationID).
			WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows affected

		// Вызываем тестируемую функцию
		err := repo.Cancel(context.Background(), notificationID)

		// В данном случае, отсутствие обновления строки не является ошибкой, поэтому возвращаем nil
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DBError", func(t *testing.T) {
		repo, mock, cleanup := newTestRepo(t)
		defer cleanup()

		notificationID := "some-id"

		// Мокируем ошибку БД
		mock.ExpectExec(regexp.QuoteMeta(`
   UPDATE notifications SET status=$1, updated_at=now() WHERE id=$2
  `)).WithArgs(models.StatusCanceled, notificationID).
			WillReturnError(fmt.Errorf("database connection error"))

		// Вызываем тестируемую функцию
		err := repo.Cancel(context.Background(), notificationID)

		// Проверяем результаты
		assert.Error(t, err)
		assert.ErrorContains(t, err, "database connection error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("InvalidStatus", func(t *testing.T) {
		repo, mock, cleanup := newTestRepo(t)
		defer cleanup()

		notificationID := "invalid-status-id"

		//  Мокируем ситуацию, когда база данных возвращает ошибку, говорящую об ограничении (constraint)
		//  Например, если попытаться установить status в значение, которое не разрешено
		mock.ExpectExec(regexp.QuoteMeta(`
   UPDATE notifications SET status=$1, updated_at=now() WHERE id=$2
  `)).WithArgs(models.StatusCanceled, notificationID).
			WillReturnError(fmt.Errorf("pq: invalid input value for enum \"notification_status\": \"cancelled\"")) // Замените "pq" на драйвер вашей БД, если он другой.

		// Вызываем тестируемую функцию
		err := repo.Cancel(context.Background(), notificationID)

		// Проверяем результаты
		assert.Error(t, err)
		assert.ErrorContains(t, err, "invalid input value for enum") // Проверяем, что ошибка связана с некорректным статусом
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
