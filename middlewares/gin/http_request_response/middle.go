package http_request_response

import (
	"context"
	"github.com/gin-gonic/gin"
	"sync/atomic"
)

type LogFunc func(context.Context, *AccessLog)

type HTTPRequestResponse struct {
	// 是否允许打印请求体
	allowReq *atomic.Bool
	// 打印请求体的最大长度，单位字节
	reqLen *atomic.Uint32
	// 是否允许打印响应体
	allowResp *atomic.Bool
	// 打印的响应体的最大长度，单位字节
	respLen *atomic.Uint32
	// 打印的请求路径最大长度，单位字节
	maxUrl *atomic.Uint32
	// 是否允许打印请求来源IP
	allowSource *atomic.Bool
	// 是否允许打印开始时间和结束时间
	allowStartAndEndTime *atomic.Bool
	// 日志回调
	l LogFunc
}

func NewHTTPRequestResponse(urlLen uint32, l LogFunc) *HTTPRequestResponse {
	r := &HTTPRequestResponse{
		allowReq:             new(atomic.Bool),
		reqLen:               new(atomic.Uint32),
		allowResp:            new(atomic.Bool),
		respLen:              new(atomic.Uint32),
		maxUrl:               new(atomic.Uint32),
		allowSource:          new(atomic.Bool),
		allowStartAndEndTime: new(atomic.Bool),
		l:                    l,
	}
	r.maxUrl.Store(urlLen)

	return r
}

func (r *HTTPRequestResponse) AllowReq(allow bool, len uint32) *HTTPRequestResponse {
	_ = r.allowReq.Swap(allow)
	_ = r.reqLen.Swap(len)
	return r
}

func (r *HTTPRequestResponse) AllowResp(allow bool, len uint32) *HTTPRequestResponse {
	_ = r.allowResp.Swap(allow)
	_ = r.respLen.Swap(len)

	return r
}

func (r *HTTPRequestResponse) AllowSource(allow bool) *HTTPRequestResponse {
	_ = r.allowSource.Swap(allow)
	return r
}

func (r *HTTPRequestResponse) AllowStartAndEndTime(allow bool) *HTTPRequestResponse {
	_ = r.allowStartAndEndTime.Swap(allow)
	return r
}

func (r *HTTPRequestResponse) Build() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

type AccessLog struct {
	// 请求方法
	method string
	// 请求路径
	url string
	// 请求体
	reqBody string
	// 响应体
	respBody string
	// 状态码
	status string
	// 开始时间
	startTime string
	// 结束时间
	endTime string
	// 请求耗时
	duration string
}
