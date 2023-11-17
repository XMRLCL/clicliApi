package service

import (
	"clicli/db/mongodb"
	"clicli/db/mysql"
	"gorm.io/gorm"
)

var mysqlClient *gorm.DB
var mongoClient *mongodb.MongoClient

func InitMysqlClient() {
	mysqlClient = mysql.GetMysqlClient()
}

func InitMongoClient() {
	mongoClient = mongodb.GetMongoClient()
}
