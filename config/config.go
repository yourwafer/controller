package config

import "flag"

type applicationConfig struct {
	LogLevel      string
	ServerType    string
	ConsoleOutput bool
	Port          uint
}

var ApplicationConfig = new(applicationConfig)

func init() {
	logLevel := flag.String("log-level", "info", "配置log输出等级,trace/debug/info/warn/error")
	serverType := flag.String("type", "agent", "server:控制节点，agent:代理执行节点")
	consoleOutput := flag.Bool("console", true, "true:日志输出到文件同时，还会输出到控制台，false:仅输出到文件")
	port := flag.Uint("port", 11000, "agent绑定端口，或者server提供服务")
	flag.Parse()
	ApplicationConfig.LogLevel = *logLevel
	ApplicationConfig.ServerType = *serverType
	ApplicationConfig.ConsoleOutput = *consoleOutput
	ApplicationConfig.Port = *port
}
