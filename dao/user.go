package dao

import (
	"github.com/jmoiron/sqlx"
)

type User struct {
	ID       int    `db:"id" json:"id"`
	Username string `db:"username" json:"username"`
	Password string `db:"password" json:"password"`
}

// ExistUser 通过username查询数据库
func ExistUser(db *sqlx.DB, username string) (user User, err error) {
	err = db.Get(&user, "select * from user where username = ?", username)
	return user, err
}

// Register 注册用户
func Register(db *sqlx.DB, username string, password string) error {
	_, err := db.Exec("insert into user (username,password) values (?,?)", username, password)
	return err
}

// UpdatePassword 修改用户密码
func UpdatePassword(db *sqlx.DB, username string, password string) error {
	_, err := db.Exec("update user set password =? where username = ? ", password, username)
	return err
}
