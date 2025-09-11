CREATE TABLE IF NOT EXISTS notifications(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type VARCHAR(20) NOT NULL,  -- email, telegram
    status VARCHAR(20) NOT NULL DEFAULT 'scheduled',
    scheduled_at TIMESTAMPTZ NOT NULL,
    retries INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS email_notifications (
    id UUID PRIMARY KEY,
    notification_id UUID NOT NULL,
    email TEXT NOT NULL,
    subject TEXT NOT NULL,
    message TEXT NOT NULL,
    FOREIGN KEY (notification_id) REFERENCES notifications(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS telegram_notifications (
    id UUID PRIMARY KEY,
    notification_id UUID NOT NULL,
    chat_id TEXT NOT NULL,
    message TEXT NOT NULL,
    FOREIGN KEY (notification_id) REFERENCES notifications(id) ON DELETE CASCADE
);

-- Функция для обновления updated_at в таблице notifications
CREATE OR REPLACE FUNCTION update_notifications_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Триггер для автоматического обновления updated_at
CREATE OR REPLACE TRIGGER update_notification_modtime
BEFORE UPDATE ON notifications
FOR EACH ROW
EXECUTE FUNCTION update_notifications_updated_at();


-- Индексы (полезно для оптимизации запросов)
CREATE INDEX IF NOT EXISTS idx_notifications_type ON notifications (type);
CREATE INDEX IF NOT EXISTS idx_notifications_status ON notifications (status);
CREATE INDEX IF NOT EXISTS idx_email_notifications_email ON email_notifications (email);
CREATE INDEX IF NOT EXISTS idx_telegram_notifications_chat_id ON telegram_notifications (chat_id);