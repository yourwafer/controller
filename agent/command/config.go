package command

import (
	"bufio"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"os"
	"strings"
	"xa.com/manager/agent/filter"
	"xa.com/manager/agent/life"
)

type configParameters struct {
	BaseDir string
	Project string
	Branch  string
	Config  string
	Values  map[string]string
}

func init() {
	life.AddAgentInitial(func() {
		filter.RegisterHandler("/config", func(w http.ResponseWriter, r *http.Request) {
			err := r.ParseForm()
			if err != nil {
				logrus.Warn("必须使用Post Form")
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			form := r.PostForm
			params, msg := parseConfigParam(form)
			if len(msg) != 0 {
				messageLog(w, msg, nil)
				return
			}
			sepString := string([]byte{os.PathSeparator})
			filePath := strings.Join([]string{params.BaseDir, params.Project, params.Branch, params.Config}, sepString)
			fileBak := filePath + ".bak"
			err = os.Rename(filePath, fileBak)
			if err != nil {
				messageLog(w, "重命名失败"+fileBak, err)
				return
			}
			bakFile, err := os.Open(fileBak)
			if err != nil {
				messageLog(w, "打开备份配置失败", err)
				return
			}
			defer func() {
				_ = bakFile.Close()
				_ = os.Remove(fileBak)
			}()
			scanner := bufio.NewScanner(bakFile)
			newFile, err := os.Create(filePath)
			if err != nil {
				messageLog(w, "创建新文件失败", err)
				return
			}
			defer func() {
				_ = newFile.Close()
			}()
			values := params.Values
			for scanner.Scan() {
				text := scanner.Text()
				if strings.HasPrefix(text, "#") || strings.HasPrefix(text, "//") {
					_, _ = fmt.Fprintln(newFile, text)
					continue
				}
				keyVal := strings.SplitN(text, "=", 2)
				if len(keyVal) < 2 {
					_, _ = fmt.Fprintln(newFile, text)
					continue
				}
				key := strings.TrimSpace(keyVal[0])
				replaceVal, exit := values[key]
				if !exit {
					_, _ = fmt.Fprintln(newFile, text)
					continue
				}
				value := strings.TrimSpace(keyVal[1])
				_, _ = fmt.Fprintf(newFile, "%s=%s", key, replaceVal)
				_, _ = fmt.Fprintln(newFile)
				messageLog(w, filePath+";"+key+":"+value+"->"+replaceVal, nil)
			}
			messageLog(w, "替换文件成功"+filePath, nil)
		})
	})
}

func parseConfigParam(form url.Values) (configParameters, string) {
	msg := strings.Builder{}
	baseDir := form.Get("baseDir")
	if len(baseDir) == 0 {
		msg.WriteString("baseDir参数必须设置为绝对路径\r\n")
	}
	project := form.Get("project")
	if len(project) == 0 {
		msg.WriteString("project参数不可为空\n")
	}
	branch := form.Get("branch")
	if len(branch) == 0 {
		msg.WriteString("branch分支不可为空\n")
	}
	config := form.Get("config")
	if len(config) == 0 {
		msg.WriteString("config文件名参数不可为空\n")
	}
	values := make(map[string]string)
	for key, value := range form {
		if !strings.HasPrefix(key, "-") {
			continue
		}
		values[key[1:]] = value[0]
	}
	return configParameters{
		BaseDir: baseDir,
		Project: project,
		Branch:  branch,
		Config:  config,
		Values:  values,
	}, msg.String()
}
