package echo

import (
	"net/http"

	"github.com/Conansgithub/due-private/v2/codes"
	"github.com/labstack/echo/v4"
)

type Resp struct {
	Code    int    `json:"code"`           // 响应码
	Message string `json:"message"`        // 响应消息
	Data    any    `json:"data,omitempty"` // 响应数据
}

type Context interface {
	echo.Context
	// CTX 获取echo.Context
	CTX() echo.Context
	// Proxy 获取代理API
	Proxy() *Proxy
	// Failure 失败响应
	Failure(rst any) error
	// Success 成功响应
	Success(data ...any) error
	// StdRequest 获取标准请求（net/http）
	StdRequest() *http.Request
}

type context struct {
	echo.Context
	proxy *Proxy
}

// CTX 获取echo.Context
func (c *context) CTX() echo.Context {
	return c.Context
}

// Proxy 代理API
func (c *context) Proxy() *Proxy {
	return c.proxy
}

func (c *context) Failure(rst any) error {
	switch v := rst.(type) {
	case error:
		code := codes.Convert(v)

		return c.JSON(code.Code(), &Resp{Code: code.Code(), Message: code.Message()})
	case *codes.Code:
		return c.JSON(v.Code(), &Resp{Code: v.Code(), Message: v.Message()})
	default:
		return c.JSON(codes.Unknown.Code(), &Resp{Code: codes.Unknown.Code(), Message: codes.Unknown.Message()})
	}
}

// Success 成功响应
func (c *context) Success(data ...any) error {
	if len(data) > 0 {
		return c.JSON(codes.OK.Code(), &Resp{Code: codes.OK.Code(), Message: codes.OK.Message(), Data: data[0]})
	} else {
		return c.JSON(codes.OK.Code(), &Resp{Code: codes.OK.Code(), Message: codes.OK.Message()})
	}
}

// StdRequest 获取标准请求（net/http）
func (c *context) StdRequest() *http.Request {
	return c.Request()
}
