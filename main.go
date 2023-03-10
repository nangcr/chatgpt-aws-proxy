package main

import (
	"github.com/akrylysov/algnhsa"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"strings"
)

const (
	openaiURL = "https://api.openai.com"
)

func main() {
	r := gin.Default()

	// 跨域设置
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	})

	// 任意路径都会被匹配
	r.Any("/*any", func(ctx *gin.Context) {
		url := openaiURL + ctx.Request.URL.Path

		req, err := http.NewRequest(ctx.Request.Method, url, ctx.Request.Body)
		if err != nil {
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}
		req.Header = ctx.Request.Header

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}
		defer resp.Body.Close()

		// 转发响应头
		for k, v := range resp.Header {
			ctx.Header(k, strings.Join(v, ","))
		}

		ctx.Status(resp.StatusCode)

		// 转发响应体
		for {
			buff := make([]byte, 256)
			var n int
			n, err = resp.Body.Read(buff)
			if err != nil {
				break
			}
			_, err = ctx.Writer.Write(buff[:n])
			if err != nil {
				break
			}
			ctx.Writer.Flush()
		}

		if err != nil && err != io.EOF {
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}
	})

	// 启动 Lambda 函数
	algnhsa.ListenAndServe(r, nil)

	// 启动本地服务，使用时请注释掉上面的 algnhsa.ListenAndServe(r, nil)
	//r.Run(":12450")
}
