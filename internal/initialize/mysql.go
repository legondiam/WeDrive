package initialize

import (
	"WeDrive/internal/config"
	"WeDrive/internal/model"
	"fmt"

	"github.com/pkg/errors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func MysqlInit() (*gorm.DB, error) {
	//fmt.Println(config.GlobalConf.DB.Mysql.User, config.GlobalConf.DB.Mysql.Password, config.GlobalConf.DB.Mysql.Host, config.GlobalConf.DB.Mysql.Port)
	DSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.GlobalConf.DB.Mysql.User,
		config.GlobalConf.DB.Mysql.Password,
		config.GlobalConf.DB.Mysql.Host,
		config.GlobalConf.DB.Mysql.Port, "wedrive")
	dbconn, err := gorm.Open(mysql.Open(DSN), &gorm.Config{})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	db, _ := dbconn.DB()
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(10)
	_ = dbconn.AutoMigrate(&model.User{})
	err = dbconn.AutoMigrate(&model.FileStore{})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	err = dbconn.AutoMigrate(&model.UserFile{})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	err = dbconn.AutoMigrate(&model.ShareFile{})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	err = dbconn.AutoMigrate(&model.UploadSession{})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return dbconn, nil
}
