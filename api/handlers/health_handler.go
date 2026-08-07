package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jaybani/jb_cip/config"
	"github.com/jaybani/jb_cip/internal/helper"
	"github.com/jaybani/jb_cip/pkg/database"
	"github.com/jaybani/jb_cip/pkg/redis"
)

func HealthCheck(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		dbStatus := "connected"
		redisStatus := "connected"

		db, err := database.GetDB()
		if err != nil || db == nil {
			dbStatus = "disconnected"
		} else if err := db.Ping(); err != nil {
			dbStatus = "disconnected"
		}

		rdb, err := redis.GetClient()
		if err != nil || rdb == nil {
			redisStatus = "disconnected"
		} else if err := rdb.Ping(c.Context()).Err(); err != nil {
			redisStatus = "disconnected"
		}

		data := map[string]interface{}{
			"status":   "healthy",
			"version":  cfg.App.Version,
			"uptime":   time.Since(startTime).String(),
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"services": map[string]string{
				"database": dbStatus,
				"redis":    redisStatus,
			},
		}

		return helper.SendSuccess(c, "ok", data, nil)
	}
}

var startTime = time.Now()
