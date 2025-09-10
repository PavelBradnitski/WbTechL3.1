package repository

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"github.com/PavelBradnitski/WbTechL3.1/internal/models"
)

func TestNotificationRepo_Create(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("failed to create mock database connection: %v", err)
		}
		defer db.Close()

		repo := NewNotificationRepo(db)

		req := &models.Notification{
			UserID:      "user123",
			Email:       "test@example.com",
			Type:        "email",
			Message:     "Test message",
			Subject:     "Test subject",
			ScheduledAt: time.Now().Add(time.Hour),
		}

		expectedID := "notification123"

		mock.ExpectQuery(regexp.QuoteMeta(`
   INSERT INTO notifications (user_id, email, type, message, subject, scheduled_at, status, retries, created_at, updated_at)
   VALUES ($1, $2, $3, $4, $5, $6, 'pending', 0, now(), now())
   RETURNING id
  `)).
			WithArgs(req.UserID, req.Email, req.Type, req.Message, req.Subject, req.ScheduledAt).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(expectedID))

		id, err := repo.Create(context.Background(), req)

		assert.NoError(t, err)
		assert.Equal(t, expectedID, id)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("DatabaseError", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("failed to create mock database connection: %v", err)
		}
		defer db.Close()

		repo := NewNotificationRepo(db)

		req := &models.Notification{
			UserID:      "user123",
			Email:       "test@example.com",
			Type:        "email",
			Message:     "Test message",
			Subject:     "Test subject",
			ScheduledAt: time.Now().Add(time.Hour),
		}

		mock.ExpectQuery(regexp.QuoteMeta(`
   INSERT INTO notifications (user_id, email, type, message, subject, scheduled_at, status, retries, created_at, updated_at)
   VALUES ($1, $2, $3, $4, $5, $6, 'pending', 0, now(), now())
   RETURNING id
  `)).
			WithArgs(req.UserID, req.Email, req.Type, req.Message, req.Subject, req.ScheduledAt).
			WillReturnError(errors.New("database error"))

		id, err := repo.Create(context.Background(), req)

		assert.Error(t, err)
		assert.Empty(t, id)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("ScanError", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("failed to create mock database connection: %v", err)
		}
		defer db.Close()

		repo := NewNotificationRepo(db)

		req := &models.Notification{
			UserID:      "user123",
			Email:       "test@example.com",
			Type:        "email",
			Message:     "Test message",
			Subject:     "Test subject",
			ScheduledAt: time.Now().Add(time.Hour),
		}

		mock.ExpectQuery(regexp.QuoteMeta(`
        INSERT INTO notifications (user_id, email, type, message, subject, scheduled_at, status, retries, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, 'pending', 0, now(), now())
        RETURNING id
    `)).
			WithArgs(req.UserID, req.Email, req.Type, req.Message, req.Subject, req.ScheduledAt).
			WillReturnError(sql.ErrNoRows)

		id, err := repo.Create(context.Background(), req)

		assert.Error(t, err) // Expect an error
		assert.Empty(t, id)  // ID should be empty

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
}

func TestNotificationRepo_GetByID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("failed to create mock database connection: %v", err)
		}
		defer db.Close()

		repo := NewNotificationRepo(db)

		notificationID := "notification123"
		expectedNotification := &models.Notification{
			ID:          notificationID,
			UserID:      "user123",
			Email:       "test@example.com",
			Type:        "email",
			Message:     "Test message",
			Subject:     "Test subject",
			Status:      "pending", // Or whatever status is
			ScheduledAt: time.Now().Add(time.Hour),
			Retries:     0,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		rows := sqlmock.NewRows([]string{"id", "user_id", "email", "type", "message", "subject", "status", "scheduled_at", "retries", "created_at", "updated_at"}).
			AddRow(
				expectedNotification.ID,
				expectedNotification.UserID,
				expectedNotification.Email,
				expectedNotification.Type,
				expectedNotification.Message,
				expectedNotification.Subject,
				expectedNotification.Status,
				expectedNotification.ScheduledAt,
				expectedNotification.Retries,
				expectedNotification.CreatedAt,
				expectedNotification.UpdatedAt,
			)

		mock.ExpectQuery(regexp.QuoteMeta(`
   SELECT id, user_id, email, type, message, subject, status, scheduled_at, retries, created_at, updated_at
   FROM notifications WHERE id=$1
  `)).
			WithArgs(notificationID).
			WillReturnRows(rows)

		notification, err := repo.GetByID(context.Background(), notificationID)

		assert.NoError(t, err)
		assert.NotNil(t, notification)
		assert.Equal(t, expectedNotification, notification)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("failed to create mock database connection: %v", err)
		}
		defer db.Close()

		repo := NewNotificationRepo(db)

		notificationID := "nonexistent_notification"

		mock.ExpectQuery(regexp.QuoteMeta(`
   SELECT id, user_id, email, type, message, subject, status, scheduled_at, retries, created_at, updated_at
   FROM notifications WHERE id=$1
  `)).
			WithArgs(notificationID).
			WillReturnError(sql.ErrNoRows)

		notification, err := repo.GetByID(context.Background(), notificationID)

		assert.Error(t, err)
		assert.Nil(t, notification)
		assert.Equal(t, errors.Is(err, sql.ErrNoRows), true) // Verify it's a sql.ErrNoRows error

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("DatabaseError", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("failed to create mock database connection: %v", err)
		}
		defer db.Close()

		repo := NewNotificationRepo(db)

		notificationID := "notification123"

		mock.ExpectQuery(regexp.QuoteMeta(`
   SELECT id, user_id, email, type, message, subject, status, scheduled_at, retries, created_at, updated_at
   FROM notifications WHERE id=$1
  `)).
			WithArgs(notificationID).
			WillReturnError(errors.New("database error"))

		notification, err := repo.GetByID(context.Background(), notificationID)

		assert.Error(t, err)
		assert.Nil(t, notification)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
}

func TestNotificationRepo_GetAll(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("failed to create mock database connection: %v", err)
		}
		defer db.Close()

		repo := NewNotificationRepo(db)

		expectedNotifications := []*models.Notification{
			{
				ID:          "notification1",
				UserID:      "user1",
				Email:       "test1@example.com",
				Type:        "email",
				Message:     "Message 1",
				Subject:     "Subject 1",
				Status:      "pending",
				ScheduledAt: time.Now().Add(time.Hour),
				Retries:     0,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			{
				ID:          "notification2",
				UserID:      "user2",
				Email:       "test2@example.com",
				Type:        "sms",
				Message:     "Message 2",
				Subject:     "Subject 2",
				Status:      "sent",
				ScheduledAt: time.Now().Add(2 * time.Hour),
				Retries:     1,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		}

		rows := sqlmock.NewRows([]string{"id", "user_id", "email", "type", "message", "subject", "status", "scheduled_at", "retries", "created_at", "updated_at"})
		for _, n := range expectedNotifications {
			rows.AddRow(
				n.ID,
				n.UserID,
				n.Email,
				n.Type,
				n.Message,
				n.Subject,
				n.Status,
				n.ScheduledAt,
				n.Retries,
				n.CreatedAt,
				n.UpdatedAt,
			)
		}

		mock.ExpectQuery(regexp.QuoteMeta(`
   SELECT id, user_id, email, type, message, subject, status, scheduled_at, retries, created_at, updated_at
   FROM notifications
  `)).
			WillReturnRows(rows)

		notifications, err := repo.GetAll(context.Background())

		assert.NoError(t, err)
		assert.NotNil(t, notifications)
		assert.Equal(t, expectedNotifications, notifications)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("NoNotifications", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("failed to create mock database connection: %v", err)
		}
		defer db.Close()

		repo := NewNotificationRepo(db)

		mock.ExpectQuery(regexp.QuoteMeta(`
        SELECT id, user_id, email, type, message, subject, status, scheduled_at, retries, created_at, updated_at
        FROM notifications
    `)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "email", "type", "message", "subject", "status", "scheduled_at", "retries", "created_at", "updated_at"})) // Empty rows

		notifications, err := repo.GetAll(context.Background())

		assert.NoError(t, err)
		assert.NotNil(t, notifications) // Verify the slice is not nil
		assert.Empty(t, notifications)  // Verify the slice is empty

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("DatabaseError", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("failed to create mock database connection: %v", err)
		}
		defer db.Close()

		repo := NewNotificationRepo(db)

		mock.ExpectQuery(regexp.QuoteMeta(`
   SELECT id, user_id, email, type, message, subject, status, scheduled_at, retries, created_at, updated_at
   FROM notifications
  `)).
			WillReturnError(errors.New("database error"))

		notifications, err := repo.GetAll(context.Background())

		assert.Error(t, err)
		assert.Nil(t, notifications)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("ScanError", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("failed to create mock database connection: %v", err)
		}
		defer db.Close()

		repo := NewNotificationRepo(db)

		rows := sqlmock.NewRows([]string{"id", "user_id", "email", "type", "message", "subject", "status", "scheduled_at", "retries", "created_at", "updated_at"}).
			AddRow("notification1", "user1", "test1@example.com", "email", "Message 1", "Subject 1", "pending", time.Now().Add(time.Hour), 0, time.Now(), "invalid_time") // Invalid UpdatedAt

		mock.ExpectQuery(regexp.QuoteMeta(`
   SELECT id, user_id, email, type, message, subject, status, scheduled_at, retries, created_at, updated_at
   FROM notifications
  `)).WillReturnRows(rows)

		notifications, err := repo.GetAll(context.Background())

		assert.Error(t, err) // Expect scan error
		assert.Nil(t, notifications)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
}
