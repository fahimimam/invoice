package config

import (
	"log"
	"sync"
	"time"

	"github.com/spf13/viper"
)

// Redis holds postgres config
type Redis struct {
	URL          string        `yaml:"url"`
	RedisTimeOut time.Duration `yaml:"time_out"`
}

var redisOnce = sync.Once{}
var redisConfig *Redis

// loadRedis loads config from path
func loadRedis(fileName string) error {
	viper.SetConfigFile(fileName)
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}
	viper.AutomaticEnv()

	redisConfig = &Redis{
		URL:          viper.GetString("redis.url"),
		RedisTimeOut: viper.GetDuration("redis.time_out") * time.Second,
	}

	return nil
}

// GetRedis returns redis config
func GetRedis(fileName string) *Redis {
	redisOnce.Do(func() {
		err := loadRedis(fileName)
		if err != nil {
			log.Fatalf("unable to read config file: %v", err)
		}
	})

	return redisConfig
}
