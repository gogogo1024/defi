package db

import (
	"context"
	"database/sql"
	"defi/internal/config"
	"fmt"
	//"github.com/bradfitz/gomemcache/memcache"
	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	"github.com/rainycape/memcache"
	"log"
)

type DB struct {
	SQL       *sql.DB
	Redis     *redis.ClusterClient
	Memcached *memcache.Client
}

func InitDB(sqlCfg config.DBConfig, cacheCfg config.CacheConfig) *DB {
	// Initialize MySQL with multiple nodes for load balancing and failover
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", sqlCfg.User, sqlCfg.Password, sqlCfg.Host, sqlCfg.Port, sqlCfg.Database)
	sqlDB, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to open database connection: %v", err)
	}

	if err = sqlDB.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	var rdb *redis.ClusterClient
	var memcached *memcache.Client

	// Initialize cache based on type
	switch cacheCfg.Type {
	case "redis-cluster":
		rdb = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:    cacheCfg.RedisCluster.Addrs,
			Password: cacheCfg.RedisCluster.Password,
		})

		_, err = rdb.Ping(context.Background()).Result()
		if err != nil {
			log.Fatalf("Failed to connect to Redis cluster: %v", err)
		}
	case "memcached":
		memcached, err = memcache.New(cacheCfg.MemcachedCluster.Addrs...)
		if err != nil {
			log.Fatalf("Failed to connect to Memcached: %v", err)
		}
	default:
		log.Fatalf("Unsupported cache type: %s", cacheCfg.Type)
	}

	return &DB{
		SQL:       sqlDB,
		Redis:     rdb,
		Memcached: memcached,
	}
}
