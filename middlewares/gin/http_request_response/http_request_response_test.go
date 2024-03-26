package http_request_response

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"testing"
)

func TestHTTPRequestResponse(t *testing.T) {
	md := NewHTTPRequestResponse(10, func(ctx context.Context, al *AccessLog) {
		t.Logf("请求-响应: %+v", *al)
	}).AllowReq(true, 30).
		AllowResp(true, 1024).
		AllowSource(true).
		AllowStartAndEndTime(true)
	ser := gin.Default()

	ser.Use(md.Build())

	ser.POST("/ping", func(c *gin.Context) {
		c.Status(http.StatusOK)
		_, _ = c.Writer.Write([]byte("pong"))
	})

	if err := ser.Run(":8080"); err != nil {
		panic(err)
	}
}

func ExampleHTTPRequestResponse_AllowResp() {
	md := NewHTTPRequestResponse(10, func(ctx context.Context, al *AccessLog) {
		fmt.Printf("请求-响应: %+v", *al)
	}).AllowResp(true, 1024)

	ser := gin.Default()
	ser.Use(md.Build())

	ser.POST("/ping", func(c *gin.Context) {
		c.Status(http.StatusOK)
		_, _ = c.Writer.Write([]byte("pong"))
	})

	if err := ser.Run(":8080"); err != nil {
		panic(err)
	}
}

func ExampleHTTPRequestResponse_AllowReq() {
	md := NewHTTPRequestResponse(10, func(ctx context.Context, al *AccessLog) {
		fmt.Printf("请求-响应: %+v", *al)
	}).AllowReq(true, 30)

	ser := gin.Default()

	ser.Use(md.Build())

	ser.POST("/ping", func(c *gin.Context) {
		c.Status(http.StatusOK)
		_, _ = c.Writer.Write([]byte("pong"))
	})

	if err := ser.Run(":8080"); err != nil {
		panic(err)
	}
}

func ExampleHTTPRequestResponse_AllowSource() {
	md := NewHTTPRequestResponse(10, func(ctx context.Context, al *AccessLog) {
		fmt.Printf("请求-响应: %+v", *al)
	}).AllowSource(true)

	ser := gin.Default()

	ser.Use(md.Build())

	ser.POST("/ping", func(c *gin.Context) {
		c.Status(http.StatusOK)
		_, _ = c.Writer.Write([]byte("pong"))
	})

	if err := ser.Run(":8080"); err != nil {
		panic(err)
	}
}

func ExampleHTTPRequestResponse_AllowStartAndEndTime() {
	md := NewHTTPRequestResponse(10, func(ctx context.Context, al *AccessLog) {
		fmt.Printf("请求-响应: %+v", *al)
	}).AllowStartAndEndTime(true)

	ser := gin.Default()

	ser.Use(md.Build())

	ser.POST("/ping", func(c *gin.Context) {
		c.Status(http.StatusOK)
		_, _ = c.Writer.Write([]byte("pong"))
	})

	if err := ser.Run(":8080"); err != nil {
		panic(err)
	}
}
