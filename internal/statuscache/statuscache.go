package statuscache

import (
	"context"
	"fmt"
	"time"

	"github.com/PavelBradnitski/WbTechL3.1/internal/models"
	"github.com/wb-go/wbf/redis"
)

const (
	defaultTTL = 7 * 24 * time.Hour
)

type Cache struct {
	redis *redis.Client
}

func New(redisClient *redis.Client) *Cache {
	return &Cache{redis: redisClient}
}

func (c *Cache) key(id string) string {
	return fmt.Sprintf("notification:%s:status", id)
}

// SetStatus сохраняет статус уведомления в Redis с TTL.
func (c *Cache) SetStatus(ctx context.Context, id string, status models.Status) error {
	return c.redis.Client.Set(ctx, c.key(id), string(status), defaultTTL).Err()
}

// GetStatus получает статус уведомления из Redis (если есть).
func (c *Cache) GetStatus(ctx context.Context, id string) (models.Status, error) {
	val, err := c.redis.Client.Get(ctx, c.key(id)).Result()
	if err != nil {
		return "", err
	}
	return models.Status(val), nil
}
