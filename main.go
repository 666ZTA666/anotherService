package main

import (
	"log"
	"sync"
	"time"

	"github.com/valyala/fasthttp"
)

var (
	timeToLive = 1 * time.Minute
	limit      = uint16(100)
	prefix     = uint8(24)
)

func main() {
	// Читаем переменные окружения
	envs, err := getEnvs()
	if err != nil {
		log.Println(err)
		// В случае ошибки берем дефолтные значения
		envs.prefix = prefix
		envs.limit = limit
		envs.timeToLive = timeToLive
	}
	// Создаем дефолтный обработчик, который возвращает статику
	h := newHandler()
	// Создаем счетчик ай-пи адресов из мьютекса и времени жизни
	counter := newIPCounter(new(sync.RWMutex), envs.timeToLive)
	// Обвешиваем дефолтный обработчик декорацией из рейт-лимитера
	h, err = newLimiter(h, counter, envs.limit, envs.prefix)
	// Тут может быть ошибка, если настройки неправильные
	if err != nil {
		log.Fatal(err)
	}
	// Создаем ручку, которая будет обнулять счетчик запросов
	dropHandler := newDropLimitHandler(counter)
	// Собираем две ручки в роутер
	h = newCompositeHandler(h, dropHandler)
	// Запускаем сервак.
	err = fasthttp.ListenAndServe(":8080", h)
	if err != nil {
		log.Fatal(err)
	}
}
