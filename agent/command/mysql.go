package command

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"
	"xa.com/manager/agent/filter"
)

type mysqlParameters struct {
	Address    string
	Username   string
	Password   string
	Database   string
	InitScript string
	Command    string
}

func init() {
	filter.RegisterHandler("/mysql", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				if e, ok := err.(error); ok {
					messageLog(w, "系统错误", e)
				} else {
					logrus.Info(err)
					messageLog(w, "系统错误", errors.New("unknow"))
				}
			}
		}()
		err := r.ParseForm()
		if err != nil {
			logrus.Warn("必须使用Post Form")
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		form := r.PostForm
		params, msg := parseMysqlParam(form)
		if len(msg) != 0 {
			messageLog(w, msg, nil)
			return
		}
		dsn := fmt.Sprintf("%s:%s@%s(%s)/?charset=utf8&parseTime=True", params.Username, params.Password, "tcp", params.Address)
		db, err := sql.Open("mysql", dsn)
		if err != nil {
			panic("mysql初始化失败" + dsn + ";" + err.Error())
		}
		defer func(db *sql.DB) {
			_ = db.Close()
		}(db)
		switch params.Command {
		case "drop":
			_, err := db.Exec("DROP DATABASE IF EXISTS `" + params.Database + "`")
			if err != nil {
				messageLog(w, "数据库删除失败"+params.Database, err)
				return
			}
			messageLog(w, "数据库已经删除"+params.Database, nil)
			return
		case "create":
			createdRes, err := db.Exec("CREATE DATABASE IF NOT EXISTS `" + params.Database + "`")
			if err != nil {
				messageLog(w, "数据库写入失败", err)
				return
			}
			createSuc, _ := createdRes.RowsAffected()
			if createSuc == 0 {
				messageLog(w, "数据库已经存在"+params.Database, nil)
				return
			}
			if len(params.InitScript) == 0 {
				messageLog(w, "创建数据库成功"+params.Database, nil)
				return
			}
			_, _ = db.Exec("USE `" + params.Database + "`")
			cmd := exec.Command("mysql", "--reconnect", "--default-character-set=utf8", "-u", params.Username, "-p"+params.Password, params.Database /*, "-e", "source "+params.InitScript+";"*/)
			outPip, _ := cmd.StdoutPipe()
			go func() {
				scanner := bufio.NewScanner(outPip)
				for scanner.Scan() {
					text := scanner.Text()
					fmt.Println(text)
				}
				fmt.Println("scan done 结束")
			}()
			errPip, _ := cmd.StderrPipe()
			go func() {
				scanner := bufio.NewScanner(errPip)
				for scanner.Scan() {
					text := scanner.Text()
					fmt.Println(text)
				}
				fmt.Println("scan done 结束")
			}()
			sqlScript, err := os.Open(params.InitScript)
			if err != nil {
				logrus.Warn("文件路径错误", params.InitScript, err)
				_, _ = w.Write([]byte("文件路径错误" + params.InitScript))
				return
			}
			defer func() {
				_ = sqlScript.Close()
			}()
			writerPip, _ := cmd.StdinPipe()
			go func() {
				buf := make([]byte, 1024*1024)
				state, _ := sqlScript.Stat()
				fileSize := state.Size()
				var sum int64 = 0
				begin := time.Now()
				logrus.Info("开始读取文件", params.InitScript, begin)
				for {
					count, err := sqlScript.Read(buf)
					readDuration := time.Since(begin)
					if count > 0 {
						wc, we := writerPip.Write(buf[:count])
						sum += int64(wc)
						if count != wc {
							fmt.Println(readDuration, "写入sql脚本失败", count, wc)
						} else {
							percent := sum * 1000 / fileSize
							fmt.Println(readDuration, "读取进度,total:", fileSize, ",current:", sum, ">", percent, "‰")
						}
						if we != nil {
							fmt.Println("写入mysql错误", we)
							_ = writerPip.Close()
							fmt.Println("关闭write pip")
							break
						}
					}
					if err != nil {
						if err == io.EOF {
							fmt.Println(readDuration, "写入完成，总大小:", fileSize)
						} else {
							fmt.Println(readDuration, "读取文件错误", err)
						}
						_ = writerPip.Close()
						break
					}
				}
			}()
			err = cmd.Run()
			if err != nil {
				printMsg := "初始化数据库失败\n" + err.Error()
				logrus.Info(printMsg)
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(printMsg))
			} else {
				printMsg := "初始化数据库成功"
				logrus.Info(printMsg)
				_, _ = w.Write([]byte(printMsg))
			}
		}
	})
}

func parseMysqlParam(form url.Values) (mysqlParameters, string) {
	msg := strings.Builder{}
	address := form.Get("address")
	if len(address) == 0 {
		msg.WriteString("address参数必须设置为绝对路径\r\n")
	}
	userName := form.Get("userName")
	if len(userName) == 0 {
		msg.WriteString("userName参数必须设置为绝对路径\r\n")
	}
	password := form.Get("password")
	if len(password) == 0 {
		msg.WriteString("password参数不可为空\n")
	}
	database := form.Get("database")
	if len(database) == 0 {
		msg.WriteString("database不可为空\n")
	}
	initScript := form.Get("initScript")
	command := form.Get("command")
	if len(command) == 0 {
		msg.WriteString("command不可为空/支持create、drop")
	}
	return mysqlParameters{
		Address:    address,
		Username:   userName,
		Password:   password,
		Database:   database,
		InitScript: initScript,
		Command:    command,
	}, msg.String()
}
