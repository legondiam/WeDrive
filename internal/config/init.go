package config

import (
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type AppConf struct {
	Port int `mapstructure:"port"`
}
type MysqlConf struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
}
type RedisConf struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
}
type DbConf struct {
	Mysql MysqlConf `mapstructure:"mysql"`
	Redis RedisConf `mapstructure:"redis"`
}
type JwtConf struct {
	SecretKey              string        `mapstructure:"SecretKey"`
	Issuer                 string        `mapstructure:"Issuer"`
	AccessTokenExpiration  time.Duration `mapstructure:"AccessTokenExpiration"`
	RefreshTokenExpiration time.Duration `mapstructure:"RefreshTokenExpiration"`
}
type MinioConf struct {
	Endpoint      string        `mapstructure:"endpoint"`
	AccessKey     string        `mapstructure:"access_key"`
	SecretKey     string        `mapstructure:"secret_key"`
	UseSSL        bool          `mapstructure:"use_ssl"`
	BucketName    string        `mapstructure:"bucket_name"`
	Location      string        `mapstructure:"location"`
	UploadTimeout time.Duration `mapstructure:"upload_timeout"`
}
type DownloadConf struct {
	PublicBaseURL string `mapstructure:"public_base_url"`
	SignSecret    string `mapstructure:"sign_secret"`
}
type Conf struct {
	App      AppConf      `mapstructure:"app"`
	DB       DbConf       `mapstructure:"database"`
	Jwt      JwtConf      `mapstructure:"jwt"`
	Minio    MinioConf    `mapstructure:"minio"`
	Download DownloadConf `mapstructure:"download"`
}

var GlobalConf Conf

func Init() error {
	viper.SetConfigFile("config/config.yaml")
	//读取配置
	err := viper.ReadInConfig()
	if err != nil {
		return errors.WithStack(err)
	}
	//解析配置
	err = viper.Unmarshal(&GlobalConf)
	if err != nil {
		return errors.WithStack(err)
	}
	//fmt.Printf("Debug: 读取到的端口是 = %d\n", GlobalConf.DB.Mysql.Port)

	return nil
}
