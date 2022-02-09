package config

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"xa.com/manager/config"
	"xa.com/manager/server/life"
)

var (
	ProjectConfig *Project
)

//go:embed server.html
var serverHtml string

type MysqlConfiguration struct {
	Username  string            `json:"username"`  // 数据库账号
	Password  string            `json:"password"`  // 数据库密码
	Address   string            `json:"address"`   // mysql地址
	Databases map[string]string `json:"databases"` // 创建数据库 <数据库名称，数据库初始化文件>
}

type FileConfig struct {
	FileName string            `json:"fileName"` // 文件名
	Values   map[string]string `json:"values"`   // key/value值
}

type JavaItem struct {
	JavaClass string `json:"javaClass"` // 启动类
	Memory    string `json:"memory"`    // 启动内存配置
}

type JavaConfiguration struct {
	Name     string               `json:"name"`     // 使用签名配置Svn目录名称
	Commands map[string]*JavaItem `json:"commands"` // java命令
}

type Branch struct {
	Agent        string              `json:"agent"`   // 部署节点
	SvnResources map[string]string   `json:"svn"`     // 下载svn资源,<目录名,svn路径>
	Mysql        MysqlConfiguration  `json:"mysql"`   // 创建并初始化相关数据库
	Configs      []FileConfig        `json:"configs"` // 配置相关文件
	Java         []JavaConfiguration `json:"java"`
}

type Project struct {
	Project  string             `json:"project"`  // 项目名称
	BaseDir  string             `json:"baseDir"`  // 基础路径
	SvnUser  string             `json:"svnUser"`  // svn用户名
	SvnPass  string             `json:"svnPass"`  // svn密码
	Branches map[string]*Branch `json:"branches"` // 所有分支
}

type ProjectRes struct {
	Project  string   `json:"project"`
	Branches []string `json:"branches"`
}

func init() {
	life.AddServerInitial(func() {
		http.HandleFunc("/", html)
		http.HandleFunc("/list", list)

		applicationConfig := config.ApplicationConfig
		path := applicationConfig.ConfigPath
		file, err := os.Open(path)
		if err != nil {
			if os.IsNotExist(err) {
				logrus.Error(path, "配置文件不存在")
				fmt.Println("#################配置范例#####################")
				fmt.Println("#############################################")
				os.Exit(-1)
				return
			}
			logrus.Panic(path, "文件打开错误", err)
		}
		contentByte, err := ioutil.ReadAll(file)
		if err != nil {
			logrus.Panic(path, "文件读写错误", err)
		}
		var project *Project
		err = json.Unmarshal(contentByte, &project)
		if err != nil {
			logrus.Warn(path, "解析json失败", err)
			os.Exit(-2)
			return
		}
		ProjectConfig = project
	})
}

func html(writer http.ResponseWriter, request *http.Request) {
	_, _ = writer.Write([]byte(serverHtml))
}

func list(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("ContentType", "application/json")
	branchNames := make([]string, 0, len(ProjectConfig.Branches))
	for k, _ := range ProjectConfig.Branches {
		branchNames = append(branchNames, k)
	}
	projectRes := ProjectRes{Project: ProjectConfig.Project, Branches: branchNames}
	byteContent, _ := json.Marshal(&projectRes)
	_, _ = writer.Write(byteContent)
}
