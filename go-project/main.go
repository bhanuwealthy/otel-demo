package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wealthy/go-kit/web"
)

func main() {
	middleres := []gin.HandlerFunc{}
	r := web.InitServer(middleres, web.CreateRoutes(routerGroup()))
	fmt.Println(">>>> Server is running at http://localhost:8080")
	r.Run(":8080")
}

func ping(c *gin.Context) {
	c.JSON(http.StatusOK, "OK")
}

func pingRoute() []web.Route {
	return []web.Route{
		{
			Handler: ping,
			Method:  http.MethodGet,
			Path:    "/ping/",
		},
	}
}

func routerGroup() web.RouterGroup {
	return web.RouterGroup{
		Prefix: "/api",
		Routes: pingRoute(),
	}
}
