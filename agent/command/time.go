package command

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"os/exec"
	"runtime"
	"time"
	"xa.com/manager/agent/filter"
)

func init() {
	filter.RegisterHandler("/time", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		update := query.Get("update")
		layout := "2006-01-02 15:04:05"
		if "" == update {
			now := time.Now()
			_, err := w.Write([]byte(now.Format(layout)))
			if err != nil {
				logrus.Warn("请求时间响应异常", err)
				return
			}
			return
		}
		location, _ := time.LoadLocation("Local")
		current, err := time.ParseInLocation(layout, update, location)
		if err != nil {
			_, err := w.Write([]byte("时间格式错误,请参考格式:" + layout))
			if err != nil {
				logrus.Warn("响应失败", err)
				return
			}
			return
		}
		switch runtime.GOOS {
		case "windows":
			dayFormat := current.Format("2006-01-02")
			dateCmd := exec.Command("cmd", "/c", "date", dayFormat)
			dateCmd.Stdout = w
			err := dateCmd.Run()
			if err != nil {
				logrus.Error("cmd /c date 执行异常", err)
				_, _ = w.Write([]byte("cmd /c date 执行异常" + err.Error()))
				return
			}
			timeFormat := current.Format("15:04:05")
			timeCmd := exec.Command("cmd", "/c", "time", timeFormat)
			timeCmd.Stdout = w
			err = timeCmd.Run()
			if err != nil {
				logrus.Error("cmd /c time 执行异常", err)
				_, _ = w.Write([]byte("cmd /c time 执行异常" + err.Error()))
				return
			}
			now := time.Now()
			duration := int64(now.Sub(current))
			if duration < 0 {
				duration = -duration
			}
			if duration < int64(time.Second) {
				out := fmt.Sprintf("修改系统时间成功%s", now.Format(layout))
				logrus.Info(out)
				_, _ = w.Write([]byte(out))
			} else {
				out := fmt.Sprintf("修改系统时间失败%s差距%s", now.Format(layout), time.Duration(duration).String())
				logrus.Warn(out)
				_, _ = w.Write([]byte(out))
			}
		default:
			_, err := w.Write([]byte("不支持操作系统类型" + runtime.GOOS))
			if err != nil {
				logrus.Warn("响应失败", err)
				return
			}
			return
		}
	})
}
