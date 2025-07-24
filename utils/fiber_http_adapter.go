package utils

import (
	"bytes"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

// fiberResponseWriter implements http.ResponseWriter for fasthttp.Response
type FiberResponseWriter struct {
	headers http.Header
	body    *bytes.Buffer
	status  int
	ctx     *fiber.Ctx
}

func NewFiberResponseWriter(c *fiber.Ctx) *FiberResponseWriter {
	return &FiberResponseWriter{
		headers: make(http.Header),
		body:    new(bytes.Buffer),
		ctx:     c,
	}
}

func (w *FiberResponseWriter) Header() http.Header {
	return w.headers
}

func (w *FiberResponseWriter) Write(b []byte) (int, error) {
	return w.body.Write(b)
}

func (w *FiberResponseWriter) WriteHeader(statusCode int) {
	w.status = statusCode
}

// FlushToFiber flushes the http.ResponseWriter content to Fiber's fasthttp.Response
func (w *FiberResponseWriter) FlushToFiber() {
	// Set status code
	if w.status != 0 {
		w.ctx.Status(w.status)
	}

	// Set headers
	for k, v := range w.headers {
		for _, vv := range v {
			w.ctx.Set(k, vv)
		}
	}

	// Write body
	_, _ = w.ctx.Write(w.body.Bytes()) // Using w.ctx.Write instead of w.ctx.Send
}

// ConvertFiberToHTTPRequest converts Fiber's fasthttp.Request to net/http.Request
func ConvertFiberToHTTPRequest(c *fiber.Ctx) (*http.Request, error) {
	req, err := http.NewRequest(c.Method(), c.OriginalURL(), bytes.NewReader(c.Body()))
	if err != nil {
		return nil, err
	}
	// Copy headers from Fiber to http.Request
	c.Request().Header.VisitAll(func(key, value []byte) {
		req.Header.Add(string(key), string(value))
	})

	// Set context if needed (e.g., for request-scoped values)
	req = req.WithContext(c.UserContext())

	return req, nil
}
