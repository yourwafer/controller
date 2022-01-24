package agent

import (
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	_ "xa.com/manager/agent/command"
	_ "xa.com/manager/agent/filter"
	"xa.com/manager/config"
)

func Start() {
	go func() {
		err := http.ListenAndServe("0.0.0.0:"+strconv.Itoa(int(config.ApplicationConfig.Port)), nil)
		if err != nil {
			logrus.Panic("agent Http 服务异常", err)
			return
		}
	}()
	logrus.Info("Starting Agent;bind port:", config.ApplicationConfig.Port)
}
