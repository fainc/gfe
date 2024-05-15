package response

import (
	"context"
	"fmt"
	"net"
	"os"

	"github.com/gogf/gf/v2/frame/g"

	"github.com/gogf/gf/v2/crypto/gmd5"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
)

var agent = g.Cfg().MustGet(context.Background(), "server.serverAgent").String()
var serverId = g.Cfg().MustGet(context.Background(), "server.serverId").String()
var mac = getMacAddressStr()
var serverName, _ = os.Hostname()

func SetDefaultResponseHeader(r *ghttp.Request) {
	if serverId == "" {
		serverId, _ = gmd5.Encrypt(fmt.Sprintf("%v%v", serverName, mac)) // Make a default serverId with serverName and mac.
	}
	if agent == "" {
		r.Server.SetServerAgent("Unknown") // Overwrite the default server agent "GoFrame HTTP Server".
	}
	r.Response.Header().Set("Server-Id", serverId)
	r.Response.Header().Set("Server-Timing", gconv.String(gtime.Now().Sub(r.EnterTime).Milliseconds())) // request timing(ms)
}

func getMacAddressStr() (mac string) {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		return "00:00:00:00:00:00"
	}
	for _, netInterface := range netInterfaces {
		macAddr := netInterface.HardwareAddr.String()
		if len(macAddr) == 0 {
			continue
		}
		mac += macAddr
	}
	if len(mac) == 0 {
		return "00:00:00:00:00:00"
	}
	return
}
