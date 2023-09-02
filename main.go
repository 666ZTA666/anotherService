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
	envs, err := getEnvs()
	if err != nil {
		log.Println(err)
		envs.prefix = prefix
		envs.limit = limit
		envs.timeToLive = timeToLive
	}
	var h fasthttp.RequestHandler
	h = newHandler()
	var c counter
	c = newIPCounter(new(sync.RWMutex), envs.timeToLive)
	h, err = newLimiter(h, c, envs.limit, envs.prefix)
	if err != nil {
		log.Fatal(err)
	}
	err = fasthttp.ListenAndServe(":8080", h)
	if err != nil {
		log.Fatal(err)
	}
}
