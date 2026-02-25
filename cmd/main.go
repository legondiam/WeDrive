package main

import (
	"WeDrive/internal/app"
	"WeDrive/internal/config"
	"WeDrive/pkg/logger"
	"fmt"
)

func main() {
	logger.Init()
	defer logger.S.Sync()

	err := config.Init()
	if err != nil {
		logger.S.Fatalf("加载配置失败:%+v", err)
		return
	}
	err, r := app.Init()
	if err != nil {
		logger.S.Fatalf("依赖初始化失败:%+v", err)
		return
	}

	logger.S.Info("服务启动成功")
	addr := fmt.Sprintf(":%d", config.GlobalConf.App.Port)
	r.Run(addr)

}
