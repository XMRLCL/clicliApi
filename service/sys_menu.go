package service

import (
	"clicli/db/mysql"
	"clicli/domain/model"
)

func GetSysMenuList() []model.SysMenu {
	mysql := mysql.GetMysqlClient()
	mysql.AutoMigrate(&model.SysMenu{})
	var sysMenuList []model.SysMenu
	mysql.Find(&sysMenuList)
	return sysMenuList
}
