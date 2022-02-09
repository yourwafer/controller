package command

import (
	"net/http"
	"strings"
	"xa.com/manager/server/config"
	"xa.com/manager/server/life"
)

func init() {
	life.AddServerInitial(func() {
		http.HandleFunc("/init", initBranch)
	})
}

func initBranch(writer http.ResponseWriter, request *http.Request) {
	query := request.URL.Query()
	branchName := query.Get("branch")
	branch := config.ProjectConfig.Branches[branchName]
	if branch == nil {
		_, _ = writer.Write([]byte("服务器没有配置此地址相关设置"))
		return
	}
	msgBuilder := strings.Builder{}
	// 下载svn资源
	for name, resource := range branch.SvnResources {
		url, param := buildSvn(branchName, name, resource, "checkout")
		httpPost(url, param, &msgBuilder)
	}
	// 创建数据库
	mysql := branch.Mysql
	for dbName, dbInit := range mysql.Databases {
		url, param := buildMysql(branch.Agent, mysql.Username, mysql.Password, mysql.Address, dbName, dbInit, "create")
		httpPost(url, param, &msgBuilder)
	}
	// 修改配置
	for _, fileConfig := range branch.Configs {
		url, param := buildConfig(branchName, fileConfig)
		httpPost(url, param, &msgBuilder)
	}
	_, _ = writer.Write([]byte(msgBuilder.String()))
}
