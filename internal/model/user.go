package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username    string `gorm:"unique"`
	Password    string
	TotalSpace  int64 `gorm:"default:314572800"`
	UsedSpace   int64 `gorm:"default:0"`
	MemberLevel int8  `gorm:"default:0"`
	VipExpireAt *time.Time
}
