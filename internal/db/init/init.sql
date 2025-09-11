-- создаём пользователя
DO
$do$
BEGIN
   IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'notify_user') THEN
      CREATE ROLE notify_user LOGIN PASSWORD 'notify_pass';
   END IF;
END
$do$;
-- даём права на базу
GRANT ALL PRIVILEGES ON DATABASE notifications TO notify_user;
GRANT ALL PRIVILEGES ON TABLE notifications TO notify_user;
GRANT ALL PRIVILEGES ON TABLE email_notifications TO notify_user;
GRANT ALL PRIVILEGES ON TABLE telegram_notifications TO notify_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO notify_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO notify_user;
ALTER SCHEMA public OWNER TO notify_user;