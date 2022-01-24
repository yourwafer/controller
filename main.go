package main

import (
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
	"time"
	"xa.com/manager/agent"
	"xa.com/manager/config"
	_ "xa.com/manager/config"
	"xa.com/manager/server"
)

func main() {
	switch config.ApplicationConfig.ServerType {
	case "server":
		logrus.Debug("服务以Server方式启动")
		server.Start()
	case "agent":
		agent.Start()
	default:
		logrus.Warn("启动参数[type]配置错误,使用")
		os.Exit(-1)
	}
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM)
	sig := <-signals
	logrus.Info("收到信号,准备关闭所有任务", sig.String())
	time.Sleep(2 * time.Second)
	logrus.Info("进程已经正确停止")
}
