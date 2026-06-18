package cache

import (
	"lockbox/internal/config"
	appLogger "lockbox/internal/logger"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(
	cfg *config.RedisConfig,
	appCfg *config.AppConfig,
	log appLogger.Logger,
) *redis.Client {
	opt, err := redis.ParseURL(cfg.URL)
	if err != nil {
		log.Error("failed to connect redis", err, map[string]any{})
		return nil
	}

	// if appCfg.Env == "production" {
	// 	opt.TLSConfig = &tls.Config{}
	// }

	client := redis.NewClient(opt)

	return client
}
