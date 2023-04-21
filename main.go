package main

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"hello/conf"
	"hello/dao"
	"log"
	"net/http"
	"time"
)

/*
练手项目：
模拟实现一个用户登录+提交订单的功能
要求：
	1.使用缓存
	2.接口鉴权使用middleware实现
	3.定时任务并发推送订单
	4.gin+mysql+redis+sqlx
*/

var db *sqlx.DB
var redisClient *redis.Client

func main() {

	db, redisClient = conf.Init()
	//创建gin框架
	r := gin.Default()
	//注册接口
	r.POST("/register", registerHandler)
	//登陆接口
	r.POST("/login", loginHandler)
	//注册提交订单接口
	r.POST("/order", authMiddleWare(), orderHandler)

	//定时修改订单状态
	go pushOrders()

	r.Run(":8080")

}

// registerHandler 注册功能模块
func registerHandler(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	var user dao.User
	//先查询用户在数据库存不存在
	user, err := dao.ExistUser(db, username)

	//这是一个处理的方式，sqlx没办法直接映射到一个空的结构体上。所以只能借助sql包去判断。
	if err == sql.ErrNoRows {
		//2.用户名如果在数据库不存在->注册用户
		err = dao.Register(db, username, password)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"新用户，注册账号失败 error": "用户名"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"新用户，注册账号成功 success": "用户名"})
		return

	}

	if user.ID > 0 {
		//1.用户名如果在数据库存在->修改密码
		err = dao.UpdatePassword(db, username, password)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"已经有用户，修改密码失败 error": "用户名"})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{"已经有用户，修改密码成功 success": "用户名"})
		return
	}

}

// loginHandler 登陆接口处理函数
func loginHandler(c *gin.Context) {
	//获取请求参数
	username := c.PostForm("username")
	password := c.PostForm("password")

	//查询用户信息
	var user dao.User
	user, err := dao.ExistUser(db, username)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名"})
		return
	}

	//验证密码
	if user.Password != password {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名或密码错误"})
		return
	}

	//生成token并存储到redis缓存中
	token := fmt.Sprintf("%d_%d", user.ID, time.Now())
	err = redisClient.Set(token, user.ID, time.Hour*24).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成token失败"})
		return
	}

	//返回token
	c.JSON(http.StatusOK, gin.H{"token": token})
}

// 提交订单接口处理函数
func orderHandler(c *gin.Context) {
	//请求参数amount
	amount := c.PostForm("amount")
	//获取用户id
	userID, exsits := c.Get("userID")
	if !exsits {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	//插入订单记录
	err := dao.InsertOrder(db, userID, amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "提交订单失败"})
		return
	}

	//返回成功响应
	c.JSON(http.StatusOK, gin.H{"message": "提交订单成功"})
}

// 鉴权中间件
func authMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		//获取token
		token := c.GetHeader("Authorization")
		log.Printf("输入的token 为 %s", token)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权的访问"})
			c.Abort()
			return
		}
		//验证token
		userID, err := redisClient.Get(token).Int()
		log.Printf("token专成的uid 为 %d", userID)

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权的访问的token"})
			c.Abort()
			return
		}
		c.Set("userID", userID)
		c.Next()
		log.Printf("authMiddleWare 中间件校验成功")
	}

}

// 定时任务：并发推送订单
func pushOrders() {
	for {
		//查询待推送的订单
		var orders []dao.Order
		orders, err := dao.WaitingOrders(db)
		if err != nil {
			log.Fatalln("查询待推送的订单失败：", err)
			continue
		}
		//并发推送订单
		for _, order := range orders {
			go pushOrder(order)
		}
		time.Sleep(time.Second * 5)
	}
}

// 推送订单
func pushOrder(order dao.Order) {
	//模拟推送订单的过程，
	time.Sleep(time.Second * 1)
	err := dao.PushOrder(db, order.ID)
	if err != nil {
		log.Println("更新订单状态失败：", err)
		return
	}
	log.Printf("订单%d推送成功\n", order.ID)
}
