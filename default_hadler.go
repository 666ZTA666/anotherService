package main

import (
	"github.com/valyala/fasthttp"
)

// body - тело ответа
var body = []byte(`
{
"biba":"boba"
}
`)

// hand - пустая структура для привязывания метода-обработчика
type hand struct{}

// newHandler - конструктор для метода обработчика
func newHandler() fasthttp.RequestHandler {
	return hand{}.handle
}

// handle - метод обработчик, всегда будет писать в тело body
func (hand) handle(ctx *fasthttp.RequestCtx) {
	ctx.Response.SetBody(body)
}
