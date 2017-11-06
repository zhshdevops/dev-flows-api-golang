package main

import (
	_ "dev-flows-api-golang/routers"
	"flag"
	"github.com/astaxie/beego"
	context "github.com/astaxie/beego/context"
)



func init() {
	level := beego.AppConfig.String("logLevel")
	if "" != level {
		flag.Set("v", level)
	}

}

func main() {

	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}

	beego.Get("/",func(ctx *context.Context){
		ctx.Output.Body([]byte("Enncloud DevOps API running"))
	})


	beego.Run()
}
