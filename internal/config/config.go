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

type RedisClusterConfig struct {
	Addrs    []string
	Password string
}
type MemcachedClusterConfig struct {
	Addrs    []string
	Username string
	Password string
}
type CacheConfig struct {
	Type             string
	RedisCluster     RedisClusterConfig
	MemcachedCluster MemcachedClusterConfig
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
	// Determine the environment and load the corresponding .env.local file
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "local" // Default to local if APP_ENV is not set
	}
	if loadErr := godotenv.Load(fmt.Sprintf(".env.%s", env)); loadErr != nil {
		log.Printf("Warning: .env.%s file not found or failed to load: %v", env, loadErr)
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
	serverPortStr := os.Getenv("NACOS_SERVER_PORT")
	if serverIP == "" {
		serverIP = "127.0.0.1" // Default IP
	}
	if serverPortStr == "" {
		serverPortStr = "8848" // Default port
	}
	serverPort, convErr := strconv.Atoi(serverPortStr)
	if convErr != nil {
		return MQConfigs{}, DBConfigs{}, CacheConfigs{}, fmt.Errorf("invalid NACOS_SERVER_PORT (%s): %w", serverPortStr, convErr)
	}

	serverConfigs := []constant.ServerConfig{
		{
			IpAddr: serverIP,
			Port:   uint64(serverPort),
		},
	}

	clientConfig := constant.ClientConfig{
		NamespaceId:         "de1eff7d-6809-4b59-a33e-25a2d6ae6e6c",
		TimeoutMs:           10000, // Increase timeout to 10 seconds
		NotLoadCacheAtStart: true,
		LogDir:              "nacos/log",
		CacheDir:            "nacos/cache",
		LogLevel:            "debug",                     // Enable detailed logging
		Username:            os.Getenv("NACOS_USERNAME"), // Set Nacos username
		Password:            os.Getenv("NACOS_PASSWORD"), // Set Nacos password
	}

	configClient, createErr := clients.CreateConfigClient(map[string]interface{}{
		"serverConfigs": serverConfigs,
		"clientConfig":  clientConfig,
	})
	if createErr != nil {
		return MQConfigs{}, DBConfigs{}, CacheConfigs{}, fmt.Errorf("create config client failed: %w", createErr)
	}

	if loadErr := loadConfigFromNacos(configClient, "kafka-config", &mqConfigs.Kafka); loadErr != nil {
		return MQConfigs{}, DBConfigs{}, CacheConfigs{}, fmt.Errorf("failed to load kafka-config: %w", loadErr)
	}
	if loadErr := loadConfigFromNacos(configClient, "mysql-config", &dbConfigs.MySQL); loadErr != nil {
		return MQConfigs{}, DBConfigs{}, CacheConfigs{}, fmt.Errorf("failed to load mysql-config: %w", loadErr)
	}
	if loadErr := loadConfigFromNacos(configClient, "nats-config", &mqConfigs.Nats); loadErr != nil {
		return MQConfigs{}, DBConfigs{}, CacheConfigs{}, fmt.Errorf("failed to load nats-config: %w", loadErr)
	}
	if loadErr := loadConfigFromNacos(configClient, "postgres-config", &dbConfigs.Postgres); loadErr != nil {
		return MQConfigs{}, DBConfigs{}, CacheConfigs{}, fmt.Errorf("failed to load postgres-config: %w", loadErr)
	}
	if loadErr := loadConfigFromNacos(configClient, "redis-cluster-config", &cacheConfigs.Cache.RedisCluster); loadErr != nil {
		return MQConfigs{}, DBConfigs{}, CacheConfigs{}, fmt.Errorf("failed to load redis-cluster-config: %w", loadErr)
	}

	configLoaded = true
	cacheExpiry = time.Now().Add(CacheTTL)
	return mqConfigs, dbConfigs, cacheConfigs, nil
}

func loadConfigFromNacos(client config_client.IConfigClient, dataId string, config interface{}) error {
	log.Printf("Fetching config for DataId: %s, Group: %s", dataId, DefaultGroup)
	content, fetchErr := client.GetConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  DefaultGroup,
	})
	if fetchErr != nil {
		return fmt.Errorf("get config failed for %s: %w", dataId, fetchErr)
	}

	log.Printf("Fetched content for %s: %s", dataId, content)
	if content == "" {
		return fmt.Errorf("config content for %s is empty", dataId)
	}

	log.Printf("Parsing content for %s", dataId)
	if parseErr := json.Unmarshal([]byte(content), config); parseErr != nil {
		log.Printf("Failed to parse config for %s: %v", dataId, parseErr)
		return fmt.Errorf("failed to parse config for %s: %w", dataId, parseErr)
	}

	log.Printf("Validating config for %s", dataId)
	if validateErr := validateConfig(config); validateErr != nil {
		log.Printf("Invalid configuration for %s: %v", dataId, validateErr)
		return fmt.Errorf("invalid configuration for %s: %w", dataId, validateErr)
	}

	// Listen for configuration changes
	listenErr := client.ListenConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  DefaultGroup,
		OnChange: func(namespace, group, dataId, data string) {
			log.Printf("Config changed for %s: %s\n", dataId, data)
			cacheMutex.Lock()
			defer cacheMutex.Unlock()
			if parseErr := json.Unmarshal([]byte(data), config); parseErr != nil {
				log.Printf("Failed to parse updated config for %s: %v", dataId, parseErr)
				return
			}
			if validateErr := validateConfig(config); validateErr != nil {
				log.Printf("Invalid updated configuration for %s: %v", dataId, validateErr)
				return
			}
			configLoaded = false // Invalidate cache
		},
	})
	if listenErr != nil {
		return fmt.Errorf("listen config failed for %s: %w", dataId, listenErr)
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
	case *RedisClusterConfig:
		if len(cfg.Addrs) == 0 {
			return errors.New("at least one address is required")
		}
	default:
		return errors.New("unknown config type")
	}
	return nil
}
