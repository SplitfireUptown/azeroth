package main

import (
	"flash-turtle/router"
	"github.com/gin-gonic/gin"
	"net/http"
)

func sayHello(c *gin.Context) {

	//返回JSON类型的数据
	c.JSON(http.StatusOK, gin.H{
		"message": "flash turtle",
	})
}

func main() {

	//创建一个默认的路由引擎
	r := gin.Default()
	r = router.InitRouter(r)
	r.Run("localhost:7777")
}
