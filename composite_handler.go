package main

import (
	"github.com/valyala/fasthttp"
)

// compositeHandler - композит, который распределяет запросы в зависимости от адреса запроса
type compositeHandler struct {
	main, drop fasthttp.RequestHandler
}

// newCompositeHandler - конструктор для данного роутер\обработчика.
func newCompositeHandler(main fasthttp.RequestHandler, drop fasthttp.RequestHandler) fasthttp.RequestHandler {
	return (&compositeHandler{main: main, drop: drop}).handle
}

// handle - тут всё просто как 2х2 Я тесты писать не буду
func (c *compositeHandler) handle(ctx *fasthttp.RequestCtx) {
	// Смотрим на адрес запроса.
	switch string(ctx.Path()) {
	// Если адрес "drop".
	case "/drop", "/drop/":
		// Тогда переводим на дроп обработчик.
		c.drop(ctx)
		// Выходим.
		return
		// Во всех остальных случаях
	default:
		// Используем основной обработчик.
		c.main(ctx)
		// Выходим.
		return
	}
}
