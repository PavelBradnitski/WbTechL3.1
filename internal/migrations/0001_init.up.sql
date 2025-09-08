CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type VARCHAR(20) NOT NULL,             -- канал доставки: email, telegram, sms...
    user_id TEXT NOT NULL,                 -- идентификатор пользователя (например chat_id в Telegram)
    email TEXT NOT NULL,                            -- email получателя
    message TEXT NOT NULL,
    subject TEXT NOT NULL,                        -- тема сообщения (для email) 
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    scheduled_at TIMESTAMPTZ NOT NULL,
    retries INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);


-- индекс для поиска уведомлений по времени отправки
CREATE INDEX IF NOT EXISTS idx_notifications_scheduled_at
    ON notifications (scheduled_at);
