package command

import (
	"encoding/base64"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"xa.com/manager/server/config"
)

func buildSvn(branch string, name string, resource string, command string) (string, string) {
	project := config.ProjectConfig
	curBranch := project.Branches[branch]
	if curBranch == nil {
		return "", ""
	}

	params := "baseDir=" + project.BaseDir + "&" +
		"project=" + project.Project + "&" +
		"branch=" + branch + "&" +
		"svnPath=" + resource + "&" +
		"svnUser=" + project.SvnUser + "&" +
		"svnPass=" + base64.URLEncoding.EncodeToString([]byte(project.SvnPass)) + "&" +
		"name=" + name + "&" +
		"command=" + command

	return "http://" + curBranch.Agent + "/svn", params
}

func buildMysql(agent string, username string, password string, address string, name string, init string, command string) (string, string) {
	params := "address=" + address + "&" +
		"userName=" + username + "&" +
		"password=" + password + "&" +
		"database=" + name + "&" +
		"database=" + name + "&" +
		"command=" + command + "&" +
		"initScript=" + init

	return "http://" + agent + "/mysql", params
}

func buildConfig(branch string, fileConfig config.FileConfig) (string, string) {
	project := config.ProjectConfig
	curBranch := project.Branches[branch]
	if curBranch == nil {
		return "", ""
	}
	params := "baseDir=" + project.BaseDir + "&" +
		"project=" + project.Project + "&" +
		"branch=" + branch + "&" +
		"config=" + fileConfig.FileName + "&"
	for key, val := range fileConfig.Values {
		params += key + "=" + val + "&"
	}

	return "http://" + curBranch.Agent + "/config", params
}

func buildJava(branch string, serverName string, command string, args string, item *config.JavaItem) (string, string) {
	project := config.ProjectConfig
	curBranch := project.Branches[branch]
	if curBranch == nil {
		return "", ""
	}
	params := "baseDir=" + project.BaseDir + "&" +
		"project=" + project.Project + "&" +
		"branch=" + branch + "&" +
		"name=" + serverName + "&" +
		"command=" + command + "&" +
		"javaClass=" + item.JavaClass + "&" +
		"spaceMB=" + item.Memory + "&" +
		"args=" + args

	return "http://" + curBranch.Agent + "/java", params
}

func httpPost(url, param string, msgBuilder *strings.Builder) {
	client := &http.Client{}
	var data = strings.NewReader(param)
	req, err := http.NewRequest("POST", url, data)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("User-Agent", "Apipost client Runtime/+https://www.apipost.cn/")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	msg, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		msgBuilder.WriteString(err.Error())
		msgBuilder.WriteString("\n")
		return
	}
	msgBuilder.Write(msg)
	msgBuilder.WriteString("\n")
}
