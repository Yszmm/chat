package main

import (
	"net/http"

	"example.com/m/controller"
	"example.com/m/ws"
	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	go ws.Manager.Start()
	r := gin.Default()
	//绑定视图
	r.LoadHTMLGlob("views/*")
	//加载静态资源
	r.StaticFS("/static", http.Dir("static"))
	//目录
	sr := r.Group("/", controller.EnableCookieSession())
	{
		sr.GET("/", controller.Index)
		sr.POST("/login", controller.Login)
		sr.GET("/ws", ws.WsHandler)
		sr.GET("/home", controller.Home)
		sr.GET("/room/:room_id", controller.Room)
	}
	r.Run(":9090")
}
