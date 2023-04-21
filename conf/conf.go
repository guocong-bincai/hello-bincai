package conf

import (
	"github.com/go-redis/redis"
	"github.com/jmoiron/sqlx"
	"log"
)

//var db *sqlx.DB
//var redisClient *redis.Client

func Init() (db *sqlx.DB, redisClient *redis.Client) {
	var err error
	db, err = sqlx.Open("mysql", "root:password@tcp(127.0.0.1:3306)/test")
	if err != nil {
		log.Fatalln("连接Mysql数据库失败：", err)
		return
	}

	//连接redis缓存
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,
	})
	_, err = redisClient.Ping().Result()
	if err != nil {
		log.Fatalln("连接Redis数据库失败：", err)
		return
	}
	return db, redisClient
}
