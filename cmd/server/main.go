package main

import (
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"

	"github.com/wb-go/wbf/ginext"

	"github.com/PavelBradnitski/WbTechL3.1/internal/handler"
	"github.com/PavelBradnitski/WbTechL3.1/internal/repository"
	"github.com/PavelBradnitski/WbTechL3.1/internal/service"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/wb-go/wbf/dbpg"
)

func runMigrations(dsn string) {
	m, err := migrate.New(
		"file://migrations",
		dsn,
	)
	if err != nil {
		log.Fatalf("migration init failed: %v", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("migration failed: %v", err)
	}
	log.Println("migrations applied")
}

func main() {
	// читаем параметры подключения
	masterDSN := os.Getenv("POSTGRES_MASTER_DSN")
	if masterDSN == "" {
		masterDSN = "postgres://notify_user:notify_pass@postgres:5432/notifications?sslmode=disable"
	}
	slaveDSNs := []string{}
	if s := os.Getenv("POSTGRES_SLAVE_DSN"); s != "" {
		slaveDSNs = append(slaveDSNs, s)
	}
	runMigrations(masterDSN)
	// создаём подключение через нашу обёртку
	db, err := dbpg.New(masterDSN, slaveDSNs, &dbpg.Options{
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
	})
	if err != nil {
		log.Fatalf("failed to init db: %v", err)
	}

	// репозиторий
	repo := repository.NewNotificationRepo(db.Master)

	// сервис
	svc := service.NewNotificationService(repo)

	// http engine
	r := ginext.New()

	// хендлеры
	handler.NewNotificationHandler(r, svc)
	// // создаём планировщик
	// url := os.Getenv("RABBITMQ_URL")
	// if url == "" {
	// 	url = "amqp://guest:guest@localhost:5672/"
	// }
	// rabbit, err := rabbitmq.Connect(url, 5, 5*time.Second)
	// if err != nil {
	// 	panic(err)
	// }
	// scheduler := service.NewScheduler(svc, rabbit, "notifications", 5*time.Second)
	// scheduler.Start()
	// defer scheduler.Stop()
	// запуск сервера
	addr := ":8080"
	if envAddr := os.Getenv("HTTP_ADDR"); envAddr != "" {
		addr = envAddr
	}
	if err := r.Run(addr); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
