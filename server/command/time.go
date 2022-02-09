package command

import (
	"encoding/base64"
	"net/http"
	"strings"
	"xa.com/manager/server/config"
	"xa.com/manager/server/life"
)

func init() {
	life.AddServerInitial(func() {
		http.HandleFunc("/time", timeFunc)
	})
}

func timeFunc(writer http.ResponseWriter, request *http.Request) {
	query := request.URL.Query()
	branchName := query.Get("branch")
	branch := config.ProjectConfig.Branches[branchName]
	if branch == nil {
		_, _ = writer.Write([]byte("服务器没有配置此地址相关设置"))
		return
	}
	time := query.Get("time")
	time = base64.URLEncoding.EncodeToString([]byte(time))
	url := buildTime(branch.Agent, time)
	msgBuilder := strings.Builder{}
	httpGet(url, &msgBuilder)
	_, _ = writer.Write([]byte(msgBuilder.String()))
}
