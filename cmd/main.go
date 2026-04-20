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

	if err := config.Init(); err != nil {
		logger.S.Fatalf("加载配置失败:%+v", err)
		return
	}

	application, err := app.Init()
	if err != nil {
		logger.S.Fatalf("依赖初始化失败:%+v", err)
		return
	}
	application.StartBackgroundJobs()

	logger.S.Info("服务启动成功")
	addr := fmt.Sprintf("0.0.0.0:%d", config.GlobalConf.App.Port)
	application.Engine.Run(addr)
}
