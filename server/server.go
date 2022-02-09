package server

import (
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"xa.com/manager/config"
	_ "xa.com/manager/server/command"
	_ "xa.com/manager/server/config"
	"xa.com/manager/server/life"
)

func Start() {
	go func() {
		err := http.ListenAndServe("0.0.0.0:"+strconv.Itoa(int(config.ApplicationConfig.Port)), nil)
		if err != nil {
			logrus.Panic("main Http 服务异常", err)
			return
		}
	}()
	life.CallInit()
	logrus.Info("Starting Main Server;bind port:", config.ApplicationConfig.Port)
}
