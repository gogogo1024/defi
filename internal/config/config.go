package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	DefaultGroup = "DEFAULT_GROUP"
	CacheTTL     = 5 * time.Minute // Set TTL for cached configurations
)

type MQConfig struct {
	Type    string
	Brokers []string
	URL     string
}

type DBConfig struct {
	Type     string
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

type CacheConfig struct {
	Type     string
	Host     string
	Port     int
	Password string
	DB       int
}

type MQConfigs struct {
	Kafka MQConfig
	Nats  MQConfig
}

type DBConfigs struct {
	MySQL    DBConfig
	Postgres DBConfig
}

type CacheConfigs struct {
	Cache CacheConfig
}

var (
	mqConfigs    MQConfigs
	dbConfigs    DBConfigs
	cacheConfigs CacheConfigs
	configLoaded bool
	cacheMutex   sync.RWMutex
	cacheExpiry  time.Time
)

func LoadConfig() (MQConfigs, DBConfigs, CacheConfigs, error) {
	// Determine the environment and load the corresponding .env file
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "local" // Default to local if APP_ENV is not set
	}
	err := godotenv.Load(fmt.Sprintf(".env.%s", env))
	if err != nil {
		log.Fatalf("Error loading .env.%s file: %v", env, err)
	}

	cacheMutex.RLock()
	if configLoaded && time.Now().Before(cacheExpiry) {
		cacheMutex.RUnlock()
		return mqConfigs, dbConfigs, cacheConfigs, nil
	}
	cacheMutex.RUnlock()

	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	if configLoaded && time.Now().Before(cacheExpiry) {
		return mqConfigs, dbConfigs, cacheConfigs, nil
	}

	serverIP := os.Getenv("NACOS_SERVER_IP")
	serverPort, err := strconv.Atoi(os.Getenv("NACOS_SERVER_PORT"))
	if err != nil {
		return MQConfigs{}, DBConfigs{}, CacheConfigs{}, fmt.Errorf("invalid NACOS_SERVER_PORT: %w", err)
	}

	serverConfigs := []constant.ServerConfig{
		{
			IpAddr: serverIP,
			Port:   uint64(serverPort),
		},
	}

	clientConfig := constant.ClientConfig{
		NamespaceId:         "public",
		TimeoutMs:           10000, // Increase timeout to 10 seconds
		NotLoadCacheAtStart: true,
		LogDir:              "nacos/log",
		CacheDir:            "nacos/cache",
		LogLevel:            "debug", // Enable detailed logging
	}

	configClient, err := clients.CreateConfigClient(map[string]interface{}{
		"serverConfigs": serverConfigs,
		"clientConfig":  clientConfig,
	})
	if err != nil {
		return MQConfigs{}, DBConfigs{}, CacheConfigs{}, fmt.Errorf("create config client failed: %w", err)
	}

	if err := loadConfigFromNacos(configClient, "kafka-config", &mqConfigs.Kafka); err != nil {
		return MQConfigs{}, DBConfigs{}, CacheConfigs{}, err
	}
	if err := loadConfigFromNacos(configClient, "mysql-config", &dbConfigs.MySQL); err != nil {
		return MQConfigs{}, DBConfigs{}, CacheConfigs{}, err
	}
	if err := loadConfigFromNacos(configClient, "nats-config", &mqConfigs.Nats); err != nil {
		return MQConfigs{}, DBConfigs{}, CacheConfigs{}, err
	}
	if err := loadConfigFromNacos(configClient, "postgres-config", &dbConfigs.Postgres); err != nil {
		return MQConfigs{}, DBConfigs{}, CacheConfigs{}, err
	}
	if err := loadConfigFromNacos(configClient, "cache-config", &cacheConfigs.Cache); err != nil {
		return MQConfigs{}, DBConfigs{}, CacheConfigs{}, err
	}

	configLoaded = true
	cacheExpiry = time.Now().Add(CacheTTL)
	return mqConfigs, dbConfigs, cacheConfigs, nil
}

func loadConfigFromNacos(client config_client.IConfigClient, dataId string, config interface{}) error {
	content, err := client.GetConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  DefaultGroup,
	})
	if err != nil {
		return fmt.Errorf("get config failed for %s: %w", dataId, err)
	}

	fmt.Printf("Config content for %s: %s\n", dataId, content)

	if err := json.Unmarshal([]byte(content), config); err != nil {
		return fmt.Errorf("failed to parse config for %s: %w", dataId, err)
	}

	if err := validateConfig(config); err != nil {
		return fmt.Errorf("invalid configuration for %s: %w", dataId, err)
	}

	// Listen for configuration changes
	err = client.ListenConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  DefaultGroup,
		OnChange: func(namespace, group, dataId, data string) {
			log.Printf("Config changed for %s: %s\n", dataId, data)
			cacheMutex.Lock()
			defer cacheMutex.Unlock()
			if err := json.Unmarshal([]byte(data), config); err != nil {
				log.Printf("Failed to parse updated config for %s: %v", dataId, err)
				return
			}
			if err := validateConfig(config); err != nil {
				log.Printf("Invalid updated configuration for %s: %v", dataId, err)
				return
			}
			configLoaded = false // Invalidate cache
		},
	})
	if err != nil {
		return fmt.Errorf("listen config failed for %s: %w", dataId, err)
	}

	return nil
}

func validateConfig(config interface{}) error {
	switch cfg := config.(type) {
	case *MQConfig:
		if cfg.Type == "" {
			return errors.New("type is required")
		}
		if len(cfg.Brokers) == 0 {
			return errors.New("at least one broker is required")
		}
		if cfg.URL == "" {
			return errors.New("URL is required")
		}
	case *DBConfig:
		if cfg.Type == "" {
			return errors.New("type is required")
		}
		if cfg.Host == "" {
			return errors.New("host is required")
		}
		if cfg.Port == 0 {
			return errors.New("port is required")
		}
		if cfg.User == "" {
			return errors.New("user is required")
		}
		if cfg.Password == "" {
			return errors.New("password is required")
		}
		if cfg.Database == "" {
			return errors.New("database is required")
		}
	case *CacheConfig:
		if cfg.Type == "" {
			return errors.New("type is required")
		}
		if cfg.Host == "" {
			return errors.New("host is required")
		}
		if cfg.Port == 0 {
			return errors.New("port is required")
		}
	default:
		return errors.New("unknown config type")
	}
	return nil
}
