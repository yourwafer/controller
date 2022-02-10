package command

import (
	"bufio"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"xa.com/manager/agent/filter"
	"xa.com/manager/agent/life"
)

var runningJava = sync.Map{}

type javaParameters struct {
	BaseDir   string
	Project   string
	Branch    string
	Name      string
	Command   string
	JavaClass string
	SpaceMB   string
	Args      string
}

func init() {
	life.AddAgentInitial(func() {
		filter.RegisterHandler("/java", func(w http.ResponseWriter, r *http.Request) {
			err := r.ParseForm()
			if err != nil {
				logrus.Warn("必须使用Post Form")
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			form := r.PostForm
			params, msg := parseJavaParam(form)
			if len(msg) > 0 {
				messageLog(w, "请求参数异常"+msg, nil)
				return
			}
			sepString := string([]byte{os.PathSeparator})
			workDir := strings.Join([]string{params.BaseDir, params.Project, params.Branch, params.Name}, sepString)
			if params.Command == "start" {
				preProcess, _ := runningJava.Load(workDir)
				if preProcess != nil {
					messageLog(w, "已经有进程正在运行，请先终止"+strconv.Itoa(preProcess.(os.Process).Pid), nil)
					return
				}
			}
			err = os.Chdir(workDir)
			if err != nil {
				messageLog(w, "切换当前工作目录失败", err)
				return
			}
			messageLog(w, "切换java执行目录"+workDir, nil)
			files, err := ioutil.ReadDir(workDir + sepString + "lib" + sepString)
			if err != nil {
				messageLog(w, "读取文件夹失败", err)
				return
			}
			classPath := ".;resources;game-server.jar;"
			for _, file := range files {
				classPath = classPath + "lib" + sepString + file.Name() + ";"
			}
			messageLog(w, params.JavaClass, nil)
			switch params.Command {
			case "start", "reload":
				executeCommand(w, classPath, params, workDir)
			case "stop":
				preProcess, _ := runningJava.Load(workDir)
				if preProcess != nil {
					messageLog(w, "已经有进程正在运行:"+strconv.Itoa(preProcess.(*os.Process).Pid), nil)
					err := preProcess.(*os.Process).Signal(syscall.SIGTERM)
					if err == nil {
						if waitAndForceKill(w, workDir, preProcess) {
							return
						}
					} else {
						messageLog(w, "使用Signal失败，调用命令行关闭", err)
					}
				}
				executeCommand(w, classPath, params, workDir)
				if preProcess != nil {
					waitAndForceKill(w, workDir, preProcess)
				}
			}
		})
	})
}

func waitAndForceKill(w http.ResponseWriter, workDir string, preProcess interface{}) bool {
	// 最多等待5s
	from := time.Now()
	done := false
	for {
		preProcess, _ := runningJava.Load(workDir)
		if preProcess == nil {
			messageLog(w, "进程已经停止", nil)
			done = true
			return true
		}
		if time.Now().Sub(from) > 5*time.Second {
			messageLog(w, "进程:"+strconv.Itoa(preProcess.(*os.Process).Pid)+"使用Signal关闭失败，尝试使用命令关闭", nil)
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	if !done {
		err := preProcess.(*os.Process).Kill()
		messageLog(w, "Signal失败，使用kill", err)
		return true
	}
	return false
}

func executeCommand(w http.ResponseWriter, classPath string, params javaParameters, workDir string) {
	cmd := exec.Command("java", "-Dfile.encoding=UTF-8", "-Xms"+params.SpaceMB, "-Xmx"+params.SpaceMB, "-classpath", classPath, params.JavaClass, params.Args)
	pipReader, err := cmd.StdoutPipe()
	if err != nil {
		msg := "java执行失败"
		messageLog(w, msg, err)
		return
	}
	err = cmd.Start()
	if err != nil {
		msg := "java执行失败"
		messageLog(w, msg, err)
		return
	}
	done := make(chan error, 2)
	//utf8Reader := transform.NewReader(pipReader, simplifiedchinese.GBK.NewDecoder())
	msgBuilder := strings.Builder{}
	go func() {
		runningJava.Store(workDir, cmd.Process)
		logrus.Info("正在执行的java进程", workDir, cmd.Process.Pid)
		err := cmd.Wait()
		runningJava.Delete(workDir)
		logrus.Info("java进程停止", workDir, cmd.Process.Pid)
		if err != nil {
			msgBuilder.WriteString("java停止")
			msgBuilder.WriteString(err.Error())
			fmt.Println("wait done 结束")
			done <- err
			return
		}
		done <- nil
		fmt.Println("wait done 结束")
	}()
	go func() {
		scanner := bufio.NewScanner(pipReader)
		printConsole := true
		for scanner.Scan() {
			text := scanner.Text()
			if printConsole {
				fmt.Println(text)
			}
			msgBuilder.WriteString(scanner.Text())

			if strings.Contains(text, "com.xa.shennu.game.Start.main(Start.java:60)") || strings.Contains(text, "com.xa.shennu.center.Start.main(Start.java:67)") {
				printConsole = false
				done <- nil
			}
		}
		fmt.Println("scan done 结束")
	}()
	ret := <-done
	if ret != nil {
		msgBuilder.WriteString("java执行失败 " + ret.Error())
		_, _ = w.Write([]byte(msgBuilder.String()))
		return
	}
	msgBuilder.WriteString("java执行成功")
	_, _ = w.Write([]byte(msgBuilder.String()))
	return
}

func messageLog(w http.ResponseWriter, msg string, err error) {
	if err != nil {
		msg = msg + err.Error()
		w.WriteHeader(http.StatusInternalServerError)
	}
	_, _ = w.Write([]byte(msg))
	logrus.Info(msg)
}

func parseJavaParam(form url.Values) (javaParameters, string) {
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
	name := form.Get("name")
	if len(name) == 0 {
		msg.WriteString("name参数不可为空\n")
	}
	command := form.Get("command")
	if len(command) == 0 {
		msg.WriteString("command参数必须设置,start/stop\n")
	}
	javaClass := form.Get("javaClass")
	if len(javaClass) == 0 {
		msg.WriteString("javaClass参数必须设置\n")
	}
	spaceMB := form.Get("spaceMB")
	if len(spaceMB) == 0 {
		spaceMB = "2048M"
	}
	args := form.Get("args")
	return javaParameters{
		BaseDir:   baseDir,
		Project:   project,
		Branch:    branch,
		Name:      name,
		Command:   command,
		JavaClass: javaClass,
		Args:      args,
		SpaceMB:   spaceMB,
	}, msg.String()
}
