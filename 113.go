package main

import (
	"net/http"

	"github.com/astaxie/beego"
	"github.com/hprose/hprose-golang/rpc"
	ws "github.com/hprose/hprose-golang/rpc/websocket"
	_ "test.com/test/routers"
)

func hello(name string) string {
	return "Hello " + name + "!"
}
func runWebSocket() {
	service := ws.NewWebSocketService()
	service.AddFunction("hello", hello)
	//http.ListenAndServe(":8081", service)
	 http.ListenAndServeTLS(":8081", beego.AppConfig.String("HTTPSCertFile"), beego.AppConfig.String("HTTPSKeyFile"), service)
}
func main() {
	//123
	go runWebSocket()
	service := rpc.NewHTTPService()
	service.AddFunction("hello", hello)
	beego.Handler("/rpc", service)
	beego.Run()
}
