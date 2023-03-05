package main

import (
	"bytes"
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
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
		// 如果出现错误，返回一个带有错误信息和 500 状态码的响应
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
	// 释放响应体的资源
	defer resp.Body.Close()

	// 读取响应体的内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		// 如果出现错误，返回一个带有错误信息和 500 状态码的响应
		return nil, http.StatusInternalServerError, err
	}

	// 返回包含响应体内容、响应状态码和 nil 错误
	return body, resp.StatusCode, nil
}

// handler 是 Lambda 函数的处理函数
func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// 从请求头部中获取 Authorization 头部的值
	apiKey := request.Headers["authorization"]
	// 将请求体转换为字节数组
	payload := []byte(request.Body)

	// 调用 sendRequest 函数发送 HTTP 请求并获取响应
	response, statusCode, err := sendRequest(apiKey, payload)
	if err != nil {
		// 如果出现错误，返回一个带有错误信息和 500 状态码的响应
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: http.StatusInternalServerError}, nil
	}

	// 构造响应头部
	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"

	// 返回带有 HTTP 响应体和响应头部的 APIGatewayProxyResponse 对象
	return events.APIGatewayProxyResponse{Body: string(response), StatusCode: statusCode, Headers: headers}, nil
}

func main() {
	// 启动 Lambda 函数
	lambda.Start(handler)
}
