package command

import (
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"strings"
	"time"
	"xa.com/manager/server/config"
	"xa.com/manager/server/life"
	"xa.com/manager/server/socket/client"
)

func init() {
	life.AddServerInitial(func() {
		http.HandleFunc("/socket", socketFunc)
	})
}

func socketFunc(writer http.ResponseWriter, request *http.Request) {
	query := request.URL.Query()
	params := make(map[string]string)
	var branchName string
	var module, command uint16
	for key, value := range query {
		switch key {
		case "_branch":
			branchName = value[0]
		case "_module":
			ii, _ := strconv.Atoi(value[0])
			module = uint16(int16(ii))
		case "_command":
			ii, _ := strconv.Atoi(value[0])
			command = uint16(int16(ii))
		default:
			params[key] = value[0]
		}
	}
	branch := config.ProjectConfig.Branches[branchName]
	if branch == nil {
		_, _ = writer.Write([]byte("服务器没有配置此地址相关设置"))
		return
	}
	ip := strings.Split(branch.Agent, ":")[0]
	address := ip + ":" + strconv.Itoa(branch.ServerPort)
	connect, err := client.Connect(address, 5*time.Second)
	if err != nil {
		logrus.Error(address, "连接失败", err)
		_, _ = writer.Write([]byte(address + " 连接失败 " + err.Error()))
		return
	}
	defer func() {
		_ = connect.Connect.Close()
	}()
	msgChan, err := connect.Write(module, command, params)
	if err != nil {
		logrus.Error(address, "发送请求失败", err)
		_, _ = writer.Write([]byte(address + " 发送请求失败 " + err.Error()))
		return
	}
	msg := <-msgChan
	if msg.Error() > 0 {
		_, _ = writer.Write([]byte(address + " 请求已经处理,但响应错误码" + strconv.Itoa(int(msg.Error()))))
		return
	}
	_, _ = writer.Write(msg.Body)
}
