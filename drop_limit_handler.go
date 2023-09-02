package main

import (
	"net"

	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"
)

const (
	prefixQuery = "prefix"
	justIP      = "0.0.0.0"
)

// dropper - дропает\стирает записи по определенному айпишнику
type dropper interface {
	dropDataByIP(ip string)
}

// dropLimitHandler хендлер, метод которого подтирает записи о запросах в декорируемый хендлер
type dropLimitHandler struct {
	d dropper
}

// newDropLimitHandler создает новый хендлер для сброса настроек лимитов
func newDropLimitHandler(d dropper) fasthttp.RequestHandler {
	return (&dropLimitHandler{d: d}).handle
}

// handle имплементация fasthttp.RequestHandler
func (d *dropLimitHandler) handle(ctx *fasthttp.RequestCtx) {
	// парсим префикс из кверей
	localPrefix := ctx.QueryArgs().Peek(prefixQuery)
	if len(localPrefix) == 0 {
		// если он пустой, запрос плохой
		ctx.Error(fasthttp.StatusMessage(fasthttp.StatusBadRequest), fasthttp.StatusBadRequest)
		return
	}
	// возможно это костыльно и криво, но Я вообще впервые сталкиваюсь с масками префиксами и подсетями
	// Работает - не трогай (или кинь пул реквест)
	_, network, err := net.ParseCIDR(justIP + "/" + string(localPrefix))
	if err != nil {
		ctx.Error(errors.Wrap(err, "bad prefix").Error(), fasthttp.StatusBadRequest)
		return
	}
	// дропаем по данному айпишнику записи в дропере
	d.d.dropDataByIP(net.ParseIP(justIP).Mask(network.Mask).String())
}
