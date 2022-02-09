package command

import (
	"net/http"
	"strings"
	"xa.com/manager/server/config"
	"xa.com/manager/server/life"
)

func init() {
	life.AddServerInitial(func() {
		http.HandleFunc("/start", start)
	})
}

func start(writer http.ResponseWriter, request *http.Request) {
	query := request.URL.Query()
	branchName := query.Get("branch")
	branch := config.ProjectConfig.Branches[branchName]
	if branch == nil {
		_, _ = writer.Write([]byte("服务器没有配置此地址相关设置"))
		return
	}
	command := query.Get("command")
	args := query.Get("args")
	msgBuilder := strings.Builder{}
	for _, java := range branch.Java {
		commandItem := java.Commands[command]
		if commandItem == nil {
			msgBuilder.WriteString(java.Name + "不存在command[" + command + "]")
			continue
		}
		url, param := buildJava(branchName, java.Name, command, args, commandItem)
		httpPost(url, param, &msgBuilder)
	}
	_, _ = writer.Write([]byte(msgBuilder.String()))
}
