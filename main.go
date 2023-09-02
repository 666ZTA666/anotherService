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
	// парсим енвы
	envs, err := getEnvs()
	if err != nil {
		log.Println(err)
		// в случае ошибки берем дефолтные значения
		envs.prefix = prefix
		envs.limit = limit
		envs.timeToLive = timeToLive
	}
	// создаем дефолтный хендлер, который возвращает статику
	h := newHandler()
	// Создаем счетчик айпишников из мьютекса и времени жизни
	counter := newIPCounter(new(sync.RWMutex), envs.timeToLive)
	// обвешиваем дефолтный хендлер декорацией из рейт лимитера
	h, err = newLimiter(h, counter, envs.limit, envs.prefix)
	// тут может быть ошибка, если настройки неправильные
	if err != nil {
		log.Fatal(err)
	}
	// созздаем ручку, которая будет обнулять счетчик запросов
	dropHandler := newDropLimitHandler(counter)
	// Собираем две ручки в роутер
	h = newCompositeHandler(h, dropHandler)
	// запускаем сервак.
	err = fasthttp.ListenAndServe(":8080", h)
	if err != nil {
		log.Fatal(err)
	}
}
