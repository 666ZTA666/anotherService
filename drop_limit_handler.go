package main

import (
	"net"

	"github.com/asaskevich/govalidator"
	realip "github.com/ferluci/fast-realip"
	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"
)

const (
	prefixQuery = "prefix"
)

// dropper - дропает\стирает записи по определенному ай-пи адресу.
type dropper interface {
	// Удаляет записи для данного ай-пи адреса.
	dropDataByIP(ip string)
}

// dropLimitHandler - структура, метод которой подтирает записи о запросах в декорируемый обработчик.
type dropLimitHandler struct {
	d dropper
}

// newDropLimitHandler - создает новый обработчик для сброса настроек лимитов.
func newDropLimitHandler(d dropper) fasthttp.RequestHandler {
	return (&dropLimitHandler{d: d}).handle
}

// handle - имплементация fasthttp.RequestHandler.
func (d *dropLimitHandler) handle(ctx *fasthttp.RequestCtx) {
	// Читаем префикс из параметров урла.
	localPrefix := ctx.QueryArgs().Peek(prefixQuery)
	if len(localPrefix) == 0 {
		// Если он пустой, запрос плохой.
		ctx.Error(fasthttp.StatusMessage(fasthttp.StatusBadRequest), fasthttp.StatusBadRequest)
		// Выходим.
		return
	}
	// Читаем ай-пи адрес из контекста запроса.
	ip := realip.FromRequest(ctx)
	// Если ай-пи не 4-й версии, выдаем ошибку.
	if !govalidator.IsIPv4(ip) {
		ctx.Error(fasthttp.StatusMessage(fasthttp.StatusBadRequest), fasthttp.StatusBadRequest)
		// Выходим.
		return
	}
	// Возможно это костыльно и криво, но Я впервые сталкиваюсь с масками префиксами и подсетями.
	// Работает - не трогай (или кинь пул-реквест).
	_, network, err := net.ParseCIDR(ip + "/" + string(localPrefix))
	if err != nil {
		ctx.Error(errors.Wrap(err, "bad prefix").Error(), fasthttp.StatusBadRequest)
		return
	}
	// Удаляем по данному ай-пи адресу записи.
	d.d.dropDataByIP(net.ParseIP(ip).Mask(network.Mask).String())
}
