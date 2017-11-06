package routers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context/param"
)

func init() {

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers/websocket:SocketLogController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers/websocket:SocketLogController"],
		beego.ControllerComments{
			Method: "CreateAccessToken",
			Router: `/token`,
			AllowHTTPMethods: []string{"get"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["dev-flows-api-golang/controllers/websocket:SocketLogController"] = append(beego.GlobalControllerRouter["dev-flows-api-golang/controllers/websocket:SocketLogController"],
		beego.ControllerComments{
			Method: "CreateWebSocketConn",
			Router: `/`,
			AllowHTTPMethods: []string{"get"},
			MethodParams: param.Make(),
			Params: nil})

}
