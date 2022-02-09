package command

import (
	"encoding/base64"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"xa.com/manager/agent/life"
	"xa.com/manager/config"
)

type initParameters struct {
	BaseDir string
	Project string
	Branch  string
	SvnPath string
	Name    string
	Command string
	SvnUser string
	SvnPass string
}

func init() {
	life.AddAgentInitial(func() {
		http.HandleFunc("/svn", func(w http.ResponseWriter, r *http.Request) {
			err := r.ParseForm()
			if err != nil {
				logrus.Warn("必须使用Post Form")
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			form := r.PostForm
			params, msg := parseSvnParam(form)
			if len(msg) > 0 {
				_, _ = w.Write([]byte(msg))
				logrus.Info("请求参数异常", msg)
				return
			}
			if !strings.Contains(params.SvnPath, "svn://192.168.11.200/") && !strings.Contains(params.SvnPath, "https://svn.h5.xaigame.com/") {
				logrus.Warn("非法svn路径", params.SvnPath)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			workDir := strings.Join([]string{params.BaseDir, params.Project, params.Branch}, string([]byte{os.PathSeparator}))
			dstPath := workDir + string([]byte{os.PathSeparator}) + params.Name
			switch params.Command {
			case "checkout":
				args := []string{"co", "--username", params.SvnUser,
					"--password", params.SvnPass,
					"--non-interactive", "--trust-server-cert",
					params.SvnPath, dstPath}
				execSvn(w, params, args)
			case "update":
				args := []string{"up",
					"--accept", "mc", params.SvnPass,
					"--non-interactive", "--trust-server-cert",
					dstPath}
				execSvn(w, params, args)
			}
		})
	})
}

func execSvn(w http.ResponseWriter, params initParameters, args []string) {
	workDir := strings.Join([]string{params.BaseDir, params.Project, params.Branch}, string([]byte{os.PathSeparator}))
	err := os.MkdirAll(workDir, os.ModeDir)
	if err != nil {
		msg := "创建父文件夹失败" + workDir
		logrus.Warn(msg)
		_, _ = w.Write([]byte(msg))
		return
	}
	cmd := exec.Command("svn", args...)
	writers := io.MultiWriter(w, config.LogrusWriter)
	cmd.Stdout = writers
	cmd.Stderr = writers
	err = cmd.Run()
	if err != nil {
		msg := "svn命令失败失败" + strings.Join(args, " ") + err.Error()
		_, _ = w.Write([]byte(msg))
		logrus.Warn(msg)
		return
	}
	msg := "svn 命令成功" + strings.Join(args, " ")
	_, _ = w.Write([]byte(msg))
	logrus.Warn(msg)
}

func parseSvnParam(form url.Values) (initParameters, string) {
	msg := strings.Builder{}
	baseDir := form.Get("baseDir")
	if len(baseDir) == 0 {
		msg.WriteString("baseDir参数必须设置为绝对路径\r\n")
	}
	project := form.Get("project")
	if len(project) == 0 {
		msg.WriteString("project参数不可为空")
	}
	branch := form.Get("branch")
	if len(branch) == 0 {
		msg.WriteString("branch分支不可为空")
	}
	svnPath := form.Get("svnPath")
	if len(svnPath) == 0 {
		msg.WriteString("svnPath路径不可为空")
	}
	svnUser := form.Get("svnUser")
	if len(svnUser) == 0 {
		msg.WriteString("svnUser用户名不可为空")
	}
	svnPass := form.Get("svnPass")
	if len(svnPass) == 0 {
		msg.WriteString("svnPass密码不可为空")
	}
	decodeString, _ := base64.URLEncoding.DecodeString(svnPass)
	svnPass = string(decodeString)
	name := form.Get("name")
	if len(name) == 0 {
		msg.WriteString("name参数不可为空")
	}
	command := form.Get("command")
	if len(command) == 0 {
		msg.WriteString("command参数必须设置,execSvn/update")
	}
	return initParameters{
		BaseDir: baseDir,
		Project: project,
		Branch:  branch,
		SvnPath: svnPath,
		SvnUser: svnUser,
		SvnPass: svnPass,
		Name:    name,
		Command: command,
	}, msg.String()
}
