package main

import (
	"github.com/valyala/fasthttp"
)

// compositeHandler - композит, который распределяет запросы в зависимости от адреса запроса
type compositeHandler struct {
	main, drop fasthttp.RequestHandler
}

// newCompositeHandler конструктор для данного роутер\хенлера
func newCompositeHandler(main fasthttp.RequestHandler, drop fasthttp.RequestHandler) fasthttp.RequestHandler {
	return (&compositeHandler{main: main, drop: drop}).handle
}

// handle - тут всё просто как 2х2 Я тесты писать не буду
func (c *compositeHandler) handle(ctx *fasthttp.RequestCtx) {
	switch string(ctx.Path()) {
	case "/drop", "/drop/":
		c.drop(ctx)
	default:
		c.main(ctx)
		return
	}
}
