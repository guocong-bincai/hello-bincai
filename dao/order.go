package dao

import (
	"github.com/jmoiron/sqlx"
	"time"
)

type Order struct {
	ID       int    `db:"id" json:"id"`
	UserID   int    `db:"user_id" json:"user_id"`
	Amount   int    `db:"amount" json:"amount"`
	Status   int    `db:"status" json:"status"`
	CreateAt string `db:"create_at" json:"create_at"`
}

// InsertOrder 插入订单信息
func InsertOrder(db *sqlx.DB, userID any, amount string) error {
	_, err := db.Exec("insert into test.order(user_id,amount,status,create_at) values (?,?,?,?)", userID, amount, 0, time.Now().Format("2006-01-02 15:04:05"))
	return err
}

// WaitingOrders 查询待推送的订单
func WaitingOrders(db *sqlx.DB) (orders []Order, err error) {
	err = db.Select(&orders, "select * from `order` where status = 0")
	return orders, err
}

// PushOrder 推送订单
func PushOrder(db *sqlx.DB, orderId int) error {
	_, err := db.Exec("update `order` set status =1 where id = ?", orderId)
	return err
}
