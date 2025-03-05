package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"log"
	"time"
)

type MQConfig struct {
	Type    string
	Brokers []string
	URL     string
}

var mqConfig MQConfig

func LoadConfig() MQConfig {
	// Server configuration
	serverConfigs := []constant.ServerConfig{
		{
			IpAddr: "127.0.0.1",
			Port:   8848,
		},
	}

	// Client configuration
	clientConfig := constant.ClientConfig{
		NamespaceId:         "public", // replace with your namespace
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              "nacos/log",
		CacheDir:            "nacos/cache",
		LogLevel:            "debug",
	}

	// Create config client
	configClient, err := clients.CreateConfigClient(map[string]interface{}{
		"serverConfigs": serverConfigs,
		"clientConfig":  clientConfig,
	})
	if err != nil {
		log.Fatalf("Create config client failed: %v", err)
	}

	// Get configuration based on type
	var dataId string
	cfgType := "kafka" // This should be dynamically set based on your requirements
	switch cfgType {
	case "kafka":
		dataId = "kafka-config"
	case "rabbitmq":
		dataId = "rabbitmq-config"
	default:
		log.Fatalf("Unsupported config type: %s", cfgType)
	}

	var content string
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		content, err = configClient.GetConfig(vo.ConfigParam{
			DataId: dataId,
			Group:  "DEFAULT_GROUP",
		})
		if err == nil {
			break
		}
		log.Printf("Get config failed (attempt %d/%d): %v", i+1, maxRetries, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("Get config failed after %d attempts: %v", maxRetries, err)
	}

	fmt.Printf("Config content: %s\n", content)

	// Parse the content into MQConfig (assuming JSON format)
	err = json.Unmarshal([]byte(content), &mqConfig)
	if err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	// Validate the configuration
	err = validateConfig(mqConfig)
	if err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Set up a listener for configuration changes
	err = configClient.ListenConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  "DEFAULT_GROUP",
		OnChange: func(namespace, group, dataId, data string) {
			fmt.Printf("Config changed: %s\n", data)
			err := json.Unmarshal([]byte(data), &mqConfig)
			if err != nil {
				log.Printf("Failed to parse updated config: %v", err)
				return
			}
			err = validateConfig(mqConfig)
			if err != nil {
				log.Printf("Invalid updated configuration: %v", err)
				return
			}
		},
	})
	if err != nil {
		log.Fatalf("Listen config failed: %v", err)
	}

	return mqConfig
}

func validateConfig(config MQConfig) error {
	if config.Type == "" {
		return errors.New("type is required")
	}
	if len(config.Brokers) == 0 {
		return errors.New("at least one broker is required")
	}
	if config.URL == "" {
		return errors.New("URL is required")
	}
	return nil
}
