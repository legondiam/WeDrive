package config

import (
	"os"
	"strings"
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
type AdminConf struct {
	UserIDs []uint `mapstructure:"userIDs"`
}
type CookieConf struct {
	Domain   string `mapstructure:"domain"`
	Path     string `mapstructure:"path"`
	Secure   bool   `mapstructure:"secure"`
	HttpOnly bool   `mapstructure:"http_only"`
	SameSite string `mapstructure:"same_site"`
}
type Conf struct {
	App      AppConf      `mapstructure:"app"`
	DB       DbConf       `mapstructure:"database"`
	Jwt      JwtConf      `mapstructure:"jwt"`
	Minio    MinioConf    `mapstructure:"minio"`
	Download DownloadConf `mapstructure:"download"`
	Admin    AdminConf    `mapstructure:"admin"`
	Cookie   CookieConf   `mapstructure:"cookie"`
}

var GlobalConf Conf

func Init() error {
	confPath := os.Getenv("CONFIG_PATH")
	if confPath != "" {
		viper.SetConfigFile(confPath)
	} else {
		viper.SetConfigFile("config/config.yaml")
	}
	//读取配置
	err := viper.ReadInConfig()
	if err != nil {
		return errors.WithStack(err)
	}

	// 允许环境变量覆盖配置文件
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// 显式绑定关键密钥
	viper.MustBindEnv("database.mysql.password", "DATABASE_MYSQL_PASSWORD")
	viper.MustBindEnv("database.redis.password", "DATABASE_REDIS_PASSWORD")
	viper.MustBindEnv("minio.access_key", "MINIO_ACCESS_KEY")
	viper.MustBindEnv("minio.secret_key", "MINIO_SECRET_KEY")
	viper.MustBindEnv("jwt.SecretKey", "JWT_SECRET_KEY")
	viper.MustBindEnv("download.sign_secret", "DOWNLOAD_SIGN_SECRET")

	//解析配置
	err = viper.Unmarshal(&GlobalConf)
	if err != nil {
		return errors.WithStack(err)
	}

	if isProd() {
		if err := validateSecrets(GlobalConf); err != nil {
			return errors.WithMessage(err, "密钥验证失败")
		}
	}

	return nil
}

// isProd 判断是否是生产环境
func isProd() bool {
	return strings.EqualFold(os.Getenv("APP_ENV"), "prod") ||
		strings.EqualFold(os.Getenv("APP_ENV"), "production")
}

// validateSecrets 验证密钥
func validateSecrets(c Conf) error {
	if c.DB.Mysql.Password == "" {
		return errors.New("缺少 DATABASE_MYSQL_PASSWORD")
	}
	if c.Minio.AccessKey == "" || c.Minio.SecretKey == "" {
		return errors.New("缺少 MINIO_ACCESS_KEY / MINIO_SECRET_KEY")
	}
	if c.Jwt.SecretKey == "" {
		return errors.New("缺少 JWT_SECRET_KEY")
	}
	if c.Download.SignSecret == "" {
		return errors.New("缺少 DOWNLOAD_SIGN_SECRET")
	}
	return nil
}
