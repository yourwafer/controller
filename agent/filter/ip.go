package filter

import (
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"strings"
	"xa.com/manager/agent/life"
)

func init() {
	life.AddAgentInitial(func() {
		RegisterFilter(ipFilter)
	})
}

func ipFilter(writer http.ResponseWriter, request *http.Request, chain *Chain) {
	remoteAddr := request.RemoteAddr
	if strings.HasPrefix(remoteAddr, "127.0.0.1:") ||
		strings.HasPrefix(remoteAddr, "10.") ||
		strings.HasPrefix(remoteAddr, "[::1]:") ||
		strings.HasPrefix(remoteAddr, "192.168.") {
		chain.Next(writer, request)
		return
	}
	if strings.HasPrefix(remoteAddr, "172.") {
		splits := strings.Split(remoteAddr, ".")
		if len(splits) < 2 {
			fail(writer, request)
			return
		}
		second, err := strconv.Atoi(splits[1])
		if err == nil && 16 <= second && second <= 31 {
			chain.Next(writer, request)
			return
		}
	}
	fail(writer, request)
}

func fail(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusForbidden)
	logrus.Info(request.RemoteAddr, "非法访问agent")
}
