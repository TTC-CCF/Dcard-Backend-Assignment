package cache

import (
	"context"
	"encoding/json"
	"main/utils"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func Init() {
	if os.Getenv("APP_ENV") == "test" {
		RedisClient = redis.NewClient(&redis.Options{
			Addr:     os.Getenv("TEST_REDIS_HOST") + ":" + os.Getenv("TEST_REDIS_PORT"),
			Password: os.Getenv("TEST_REDIS_PASSWORD"),
		})
	} else {
		RedisClient = redis.NewClient(&redis.Options{
			Addr:     os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
			Password: os.Getenv("REDIS_PASSWORD"),
		})
	}
}

func CacheMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.Request.URL.Path + "?" + c.Request.URL.RawQuery
		data, err, _ := utils.Sfg.Do(key, func() (interface{}, error) {
			return RedisClient.Get(c, key).Result()
		})

		if err != nil {
			c.Next()
		} else {
			var jsondata []utils.Item
			err := json.Unmarshal([]byte(data.(string)), &jsondata)
			if err != nil {
				c.Next()
			}
			c.JSON(200, jsondata)
			c.Abort()
		}
	}
}

// key: url path with query parameters, value: the corresponding response
func SetCache(ctx context.Context, key string, data []utils.Item) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if _, err := RedisClient.Set(ctx, key, string(jsonData), 5*time.Minute).Result(); err != nil {
		return err
	}

	return nil
}

// key: condition kind (age | country | gender | platform), value: list of cached url path with query parmeters
func AddConditionCache(ctx context.Context, conditionKind, newKey string) error {
	_, err := RedisClient.LPush(ctx, conditionKind, newKey).Result()
	return err
}

func DeleteConditionCache(ctx context.Context, key string) error {
	// get the cached keys from all kinds of conditions
	keys, err := RedisClient.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return err
	}

	// delete the cached data
	for _, k := range keys {
		_, err = RedisClient.Del(ctx, k).Result()
		if err != nil {
			return err
		}
	}

	// delete the cached keys
	_, err = RedisClient.Del(ctx, key).Result()
	return err
}
