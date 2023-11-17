package service

import "clicli/domain/model"

func InsertManyAt(messages []model.AtMessage) error {
	return mysqlClient.Create(&messages).Error
}

func SelectAtMessage(userId uint, page, pageSize int) (messages []model.AtMessage) {
	mysqlClient.Where("uid = ?", userId).Limit(pageSize).Offset((page - 1) * pageSize).Find(&messages)
	return
}
