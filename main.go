package main

import (
	"bytes"
	"encoding/json"
	"github.com/akrylysov/algnhsa"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
)

const (
	openaiURL = "https://api.openai.com/v1/chat/completions"
)

// sendRequest 发送 HTTP 请求并获取响应
func sendRequest(apiKey string, payload []byte) ([]byte, int, error) {
	// 创建一个新的 HTTP 请求
	req, err := http.NewRequest(http.MethodPost, openaiURL, bytes.NewBuffer(payload))
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	// 设置请求头部
	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", apiKey)
	}

	// 创建一个 HTTP 客户端对象
	client := &http.Client{}

	// 发送 HTTP 请求
	resp, err := client.Do(req)
	if err != nil {
		// 如果出现错误，返回一个带有错误信息和 500 状态码的响应
		return nil, http.StatusInternalServerError, err
	}
	defer resp.Body.Close()

	// 读取响应体的内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		// 如果出现错误，返回一个带有错误信息和 500 状态码的响应
		return nil, http.StatusInternalServerError, err
	}

	// 返回响应体的内容，状态码和错误信息
	return body, resp.StatusCode, nil
}

func sendRequestSSE(ctx *gin.Context, apiKey string, payload []byte) {
	// 创建一个新的 HTTP 请求
	req, err := http.NewRequest(http.MethodPost, openaiURL, bytes.NewBuffer(payload))
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	// 设置请求头部
	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", apiKey)
	}
	req.Header.Set("Accept", "text/event-stream")

	// 执行HTTP请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer resp.Body.Close()

	// 设置响应头部
	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")

	// 转发响应体
	_, err = io.Copy(ctx.Writer, resp.Body)
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}
}

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
		// 读取请求头部的 Authorization 字段
		apiKey := ctx.GetHeader("authorization")

		// 读取请求体的内容
		payload, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}
		body := map[string]interface{}{}
		err = json.Unmarshal(payload, &body)
		if err != nil {
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}

		// 判断是否是 SSE 请求
		if sse, ok := body["stream"]; ok && sse == true {
			sendRequestSSE(ctx, apiKey, payload)
			return
		} else {
			response, statusCode, err := sendRequest(apiKey, payload)
			if err != nil {
				ctx.String(http.StatusInternalServerError, err.Error())
				return
			}
			ctx.Data(statusCode, "application/json", response)
		}

	})

	// 启动 Lambda 函数
	algnhsa.ListenAndServe(r, nil)
}
