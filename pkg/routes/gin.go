package routes

import (
	"github.com/gin-gonic/gin"
	"telegram-energy/pkg/core/global"
	"telegram-energy/pkg/middleware"
)

type AuthTelegram struct {
	Id        uint   `form:"id" query:"id"`
	FirstName string `form:"first_name" query:"first_name"`
	Username  string `form:"username" query:"username"`
	PhotoUrl  string `form:"photo_url" query:"photo_url"`
	AuthDate  int64  `form:"auth_date" query:"auth_date"`
	Hash      string `form:"hash" query:"hash"`
}

func RegisterRoutes() *gin.Engine {
	if global.App.Config.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	router.Use(gin.Logger())
	// 跨域处理
	router.Use(middleware.Cors())
	// 注册 api 分组路由

	return router
}
