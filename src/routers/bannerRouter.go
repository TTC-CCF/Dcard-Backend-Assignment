package routers

import (
	"main/controllers"

	"github.com/gin-gonic/gin"
)

func Init() *gin.Engine {
	router := gin.New()

	api := router.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			v1.POST("/ad", controllers.CreateBanner)
			v1.GET("/ad", controllers.SearchBanner)
		}
	}

	return router
}
