package controller

import (
	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
)

func Init() {
	web.Get("/", func(ctx *context.Context) {
		ctx.Output.Body([]byte("hello world"))
	})
	web.Post("/crawlpage/:id", func(ctx *context.Context) {

	})
}
