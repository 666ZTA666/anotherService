package main

import (
	"fmt"
	"net"

	"github.com/asaskevich/govalidator"
	realip "github.com/ferluci/fast-realip"
	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"
)

// counter - позволяет считать количество айпи адресов
type counter interface {
	// Выдает количество по адресу
	countOfIP(ip string) uint16
	// Увеличивает счетчик для айпи
	addIP(ip string)
}

// limitHandlerDecorator декоратор fasthttp.RequestHandler\Обработчика запроса.
type limitHandlerDecorator struct {
	// принимает декорируемый обработчик
	hand fasthttp.RequestHandler
	// счетчик айпишников
	c counter
	// предел количества запросов
	limit uint16
	// маску подсети считаем на основе префикса подсети
	mask net.IPMask
}

// newLimiter - декорирует принимаемых обработчик логикой rate-limiter-a.
func newLimiter(hand fasthttp.RequestHandler, c counter, limit uint16, prefix uint8) (fasthttp.RequestHandler, error) {
	if hand == nil {
		return nil, errors.New("bad handler")
	}
	if c == nil {
		return nil, errors.New("bad counter")
	}
	if limit == 0 {
		return nil, errors.New("empty limit")
	}
	// Создаем декоратор
	l := &limitHandlerDecorator{hand: hand, c: c, limit: limit}
	// Ищем маску для подсети из префикса
	_, network, err := net.ParseCIDR(fmt.Sprintf("0.0.0.0/%d", prefix))
	if err != nil {
		// в случае ошибки(если префикс больше 32), вернем ошибку, добавив ей описание
		return nil, errors.Wrap(err, "cant parse network mask")
	}
	// сохраним маску
	l.mask = network.Mask
	// вернем обработчик и пустую ошибку.
	return l.handle, nil
}

// handle - декорированный метод для обработки запросов
func (l *limitHandlerDecorator) handle(ctx *fasthttp.RequestCtx) {
	// Парсим айпи из контекста запроса.(Я знаю, что можно было просто в заголовок посмотреть, но так надежнее)
	ip := realip.FromRequest(ctx)
	// Проверяем что у нас айпи 4й версии, потому что логику для ipv6 заказчик не указал
	if !govalidator.IsIPv4(ip) {
		// если у нас нашелся айпи не 4й версии, тогда запишем в тело ошибку, статус код "плохой запрос"
		ctx.Error(fasthttp.StatusMessage(fasthttp.StatusBadRequest), fasthttp.StatusBadRequest)
		// и выйдем из обработки
		return
	}
	// Перепишем наш айпи с учетом маски
	ip = net.ParseIP(ip).Mask(l.mask).String()
	// посмотрим сколько у нас было записей по данному айпи
	count := l.c.countOfIP(ip)
	// если количество записей меньше лимита
	if count < l.limit {
		// тогда увеличим счетчик
		l.c.addIP(ip)
		// используем декорируемый обработчик
		l.hand(ctx)
		// и выйдем
		return
	}
	// если лимит превышен, то вернем тело, статус код 429 и всё такое.
	ctx.Error(fasthttp.StatusMessage(fasthttp.StatusTooManyRequests), fasthttp.StatusTooManyRequests)
}
