package program

import (
	"fmt"
	"log"
	"net/http"
	"time"

	gin "github.com/gin-gonic/gin"
)

// http服务

func (p *Program) startAPI() {
	router := gin.Default()

	// 跨域问题
	router.Use(p.middlewareCORS())

	// 设置静态文件目录
	router.GET("/ui/*w", p.handlerStatic)
	router.GET("/", func(c *gin.Context) {
		c.Redirect(301, "/ui")
	})

	// 启动所有版本api
	for key, val := range p.vApis {
		vAPI := router.Group("/" + key)
		vAPI.Use()
		val.Register(vAPI)
	}

	addr := fmt.Sprintf("%s:%d", p.cfg.HTTP.Address, p.cfg.HTTP.Port)
	// 监听
	s := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	log.Println("Start HTTP the service:", addr)
	err := s.ListenAndServe()
	if err != nil {
		log.Fatalln(err)
	}

}

// 跨域中间件
func (p *Program) middlewareCORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("origin")
		method := c.Request.Method

		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization, Token, EtcdID")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE, PUT")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")

		//放行所有OPTIONS方法
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		// 处理请求
		c.Next()
	}
}
