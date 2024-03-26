package http_request_response

import (
	"bytes"
	"context"
	"github.com/gin-gonic/gin"
	"io"
	"io/ioutil"
	"sync/atomic"
	"time"
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
		t := time.Now()
		l := r.maxUrl.Load()
		if uint32(len(c.Request.URL.Path)) >= l {
			c.Request.URL.Path = c.Request.URL.Path[:l]
		}

		al := &AccessLog{
			method: c.Request.Method,
			url:    c.Request.URL.Path,
		}

		if r.allowReq.Load() && c.Request.Body != nil {
			body, _ := ioutil.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewReader(body))
			rql := r.reqLen.Load()
			if uint32(len(body)) >= rql {
				body = body[:rql]
			}
			al.reqBody = string(body)
		}

		if r.allowSource.Load() {
			al.source = c.Request.RemoteAddr
		}

		if r.allowStartAndEndTime.Load() {
			al.startTime = t.String()
			al.endTime = time.Now().String()
		}

		c.Writer = &ResponseWriter{
			al:             al,
			ResponseWriter: c.Writer,
		}

		defer func() {
			al.duration = time.Since(t).String()
			r.l(c, al)
		}()

		c.Next()
	}
}

type ResponseWriter struct {
	al *AccessLog
	gin.ResponseWriter
}

func (w *ResponseWriter) Write(body []byte) (int, error) {
	w.al.respBody = string(body)
	return w.ResponseWriter.Write(body)
}

func (w *ResponseWriter) WriteString(body string) (int, error) {
	w.al.respBody = body
	return w.ResponseWriter.WriteString(body)
}

func (w *ResponseWriter) WriteHeader(statusCode int) {
	w.al.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
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
	// 请求来源IP
	source string
	// 状态码
	status int
	// 开始时间
	startTime string
	// 结束时间
	endTime string
	// 请求耗时
	duration string
}
