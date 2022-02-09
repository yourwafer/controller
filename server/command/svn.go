package command

import (
	"net/http"
	"strings"
	"xa.com/manager/server/config"
	"xa.com/manager/server/life"
)

func init() {
	life.AddServerInitial(func() {
		http.HandleFunc("/svn", svnFunc)
	})
}

func svnFunc(writer http.ResponseWriter, request *http.Request) {
	query := request.URL.Query()
	branchName := query.Get("branch")
	branch := config.ProjectConfig.Branches[branchName]
	if branch == nil {
		_, _ = writer.Write([]byte("服务器没有配置此地址相关设置"))
		return
	}
	name := query.Get("name")
	msgBuilder := strings.Builder{}
	svnResource := branch.SvnResources[name]
	url, param := buildSvn(branchName, name, svnResource, "update")
	httpPost(url, param, &msgBuilder)
	_, _ = writer.Write([]byte(msgBuilder.String()))
}
