package main

import (
	"fmt"
	"net"

	"github.com/asaskevich/govalidator"
	realip "github.com/ferluci/fast-realip"
	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"
)

// counter - позволяет считать количество ай-пи адресов.
type counter interface {
	// Выдает количество по адресу.
	countOfIP(ip string) uint16
	// Увеличивает счетчик для ай-пи.
	addIP(ip string)
}

// limitHandlerDecorator - декоратор fasthttp.RequestHandler\обработчика запроса.
type limitHandlerDecorator struct {
	// Принимает декорируемый обработчик.
	hand fasthttp.RequestHandler
	// Счетчик ай-пи адресов.
	c counter
	// Предел количества запросов.
	limit uint16
	// Маску подсети считаем на основе префикса подсети.
	mask net.IPMask
}

// newLimiter - декорирует принимаемых обработчик логикой rate-limiter-a.
func newLimiter(hand fasthttp.RequestHandler, c counter, limit uint16, prefix uint8) (fasthttp.RequestHandler, error) {
	// Проверяем функцию обработчик на пустоту.
	if hand == nil {
		return nil, errors.New("bad handler")
	}
	// Проверяем счетчик на пустоту.
	if c == nil {
		return nil, errors.New("bad counter")
	}
	// Проверяем значение лимита.
	if limit == 0 {
		return nil, errors.New("empty limit")
	}
	// Создаем декоратор.
	l := &limitHandlerDecorator{hand: hand, c: c, limit: limit}
	// Ищем маску для подсети из префикса.
	_, network, err := net.ParseCIDR(fmt.Sprintf("0.0.0.0/%d", prefix))
	if err != nil {
		// В случае ошибки(если префикс больше 32), вернем ошибку, добавив ей описание.
		return nil, errors.Wrap(err, "cant parse network mask")
	}
	// сохраним маску.
	l.mask = network.Mask
	// вернем обработчик и пустую ошибку.
	return l.handle, nil
}

// handle - декорированный метод для обработки запросов.
func (l *limitHandlerDecorator) handle(ctx *fasthttp.RequestCtx) {
	// Читаем ай-пи адрес из контекста запроса (Я знаю, что можно было просто в заголовок посмотреть, но так надежнее).
	ip := realip.FromRequest(ctx)
	// Проверяем что у нас ай-пи 4‑й версии, потому что логику для ipv6 заказчик не указал.
	if !govalidator.IsIPv4(ip) {
		// Если у нас нашелся ай-пи не 4-й версии, тогда запишем в тело ошибку, статус код "плохой запрос".
		ctx.Error(fasthttp.StatusMessage(fasthttp.StatusBadRequest), fasthttp.StatusBadRequest)
		// И выйдем из обработки.
		return
	}
	// Перепишем наш ай-пи с учетом маски.
	ip = net.ParseIP(ip).Mask(l.mask).String()
	// Посмотрим сколько у нас было записей по данному ай-пи.
	count := l.c.countOfIP(ip)
	// Если количество записей меньше лимита.
	if count < l.limit {
		// Тогда увеличим счетчик.
		l.c.addIP(ip)
		// Используем декорируемый обработчик.
		l.hand(ctx)
		// И выйдем
		return
	}
	// Если лимит превышен, то вернем тело, статус код 429 и всё такое.
	ctx.Error(fasthttp.StatusMessage(fasthttp.StatusTooManyRequests), fasthttp.StatusTooManyRequests)
}
