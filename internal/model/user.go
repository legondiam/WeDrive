package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username   string `gorm:"unique"`
	Password   string
	TotalSpace int64 `gorm:"default:1073741824"`
	UsedSpace  int64 `gorm:"default:0"`
}
