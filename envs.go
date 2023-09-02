package main

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const (
	prefixEnv = "PREFIX"
	limitEnv  = "LIMIT"
	ttlEnv    = "TTL_IN_SECONDS"
)

// envs - структура для хранения переменных окружения. Эдакий локальный конфиг.
type envs struct {
	// префикс - число от 0 до 32
	prefix uint8
	// лимит Я решил ограничить 16 битами, чтобы памяти много не жрать,
	// порефакторить размеры переменных потом будет не так и сложно
	limit uint16
	// время жизни или кулдаун.
	timeToLive time.Duration
}

// getEnvs - парсит конфиг из переменных окружения. Если какой-то переменной нет, вернёт ошибку
func getEnvs() (envs, error) {
	// создаем результирующую структуру
	res := envs{}
	// ищем префикс
	s := os.Getenv(prefixEnv)
	// если префикс пустой
	if s == "" {
		// вернём ошибку
		return res, errors.New("no PREFIX env")
	}
	// парсим префикс отрезав от него слэш в начале
	data, err := strconv.ParseUint(strings.TrimPrefix(s, "/"), 10, 32)
	if err != nil {
		// если при парсинге вылезла ошибка, обернём её своим описанием.
		return res, errors.Wrap(err, "PREFIX env not correct")
	}
	// сохраняем префикс в конфиг, валидации что он меньше 32 у нас тут нет.
	res.prefix = uint8(data)
	// ищем лимит
	s = os.Getenv(limitEnv)
	// если лимит пустой
	if s == "" {
		// вернем ошибку
		return res, errors.New("no LIMIT env")
	}
	// парсим лимит
	data, err = strconv.ParseUint(s, 10, 8)
	if err != nil {
		// если при парсинге вылезла ошибка, обернём её своим описанием
		return res, errors.Wrap(err, "LIMIT env not correct")
	}
	// сохраняем лимит в конфиг
	res.limit = uint16(data)
	// ищем ттл, он должен быть в секундах, название енвы как бэ намекает
	s = os.Getenv(ttlEnv)
	// если ттл пустой, вернём ошибку
	if s == "" {
		return res, errors.New("no TTL_IN_SECONDS env")
	}
	// парсим ттл
	data, err = strconv.ParseUint(s, 10, 63)
	if err != nil {
		// если при парсинге возникла ошибка обернём её доп контекстом
		return res, errors.Wrap(err, "TTL_IN_SECONDS env not correct")
	}
	// сохраним ттл умножив его на секунды.
	res.timeToLive = time.Second * time.Duration(data)
	// вернем конфиг без ошибок.
	return res, nil
}
