package db

import (
	"context"
	"database/sql"
	"defi/internal/config"
	"fmt"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

type DB struct {
	SQL       *sql.DB
	Redis     *redis.Client
	Memcached *memcache.Client
}

func InitDB(sqlCfg config.DBConfig, cacheCfg config.CacheConfig) *DB {
	// Initialize MySQL
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", sqlCfg.User, sqlCfg.Password, sqlCfg.Host, sqlCfg.Port, sqlCfg.Database)
	sqlDB, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to open database connection: %v", err)
	}

	if err = sqlDB.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	var rdb *redis.Client
	var mcache *memcache.Client

	// Initialize cache based on type
	switch cacheCfg.Type {
	case "redis":
		rdb = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", cacheCfg.Host, cacheCfg.Port),
			Password: cacheCfg.Password,
			DB:       cacheCfg.DB,
		})

		_, err = rdb.Ping(context.Background()).Result()
		if err != nil {
			log.Fatalf("Failed to connect to Redis: %v", err)
		}
	case "memcache":
		mcache = memcache.New(fmt.Sprintf("%s:%d", cacheCfg.Host, cacheCfg.Port))
		err = mcache.Ping()
		if err != nil {
			log.Fatalf("Failed to connect to Memcached: %v", err)
		}
	default:
		log.Fatalf("Unsupported cache type: %s", cacheCfg.Type)
	}

	return &DB{
		SQL:       sqlDB,
		Redis:     rdb,
		Memcached: mcache,
	}
}
