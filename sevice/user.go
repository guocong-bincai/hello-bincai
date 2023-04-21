package sevice

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

var (
	db *sqlx.DB
	c  *gin.Context
)
