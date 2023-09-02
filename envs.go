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
	// Префикс - число от 0 до 32.
	prefix uint8
	// Лимит Я решил ограничить 16 битами, чтобы памяти много не жрать.
	// Поправить размеры переменных потом будет не так и сложно.
	limit uint16
	// Время жизни или кулдаун.
	timeToLive time.Duration
}

// getEnvs - парсит конфиг из переменных окружения. Если какой-то переменной нет, вернёт ошибку.
func getEnvs() (envs, error) {
	// Создаем результирующую структуру
	res := envs{}
	// Ищем префиксную переменную окружения.
	s := os.Getenv(prefixEnv)
	// Если префикс пустой\такой переменной окружения не нашлось.
	if s == "" {
		// вернём ошибку
		return res, errors.New("no PREFIX env")
	}
	// Пытаемся преобразовать префикс в число, отрезав от него слэш в начале.
	data, err := strconv.ParseUint(strings.TrimPrefix(s, "/"), 10, 32)
	if err != nil {
		// если при парсинге вылезла ошибка, обернём её своим описанием.
		return res, errors.Wrap(err, "PREFIX env not correct")
	}
	// Сохраняем префикс в конфиг, валидации что он меньше 32 у нас тут нет.
	res.prefix = uint8(data)
	// Ищем лимит среди переменных окружения.
	s = os.Getenv(limitEnv)
	// Если лимит пустой.
	if s == "" {
		// Вернем ошибку.
		return res, errors.New("no LIMIT env")
	}
	// Пытаемся преобразовать лимит в числовое значение.
	data, err = strconv.ParseUint(s, 10, 8)
	if err != nil {
		// Если вылезла ошибка, обернём её своим описанием.
		return res, errors.Wrap(err, "LIMIT env not correct")
	}
	// Сохраняем лимит в конфиг.
	res.limit = uint16(data)
	// Ищем ттл в переменных окружения, он должен быть в секундах, название переменной как бэ намекает.
	s = os.Getenv(ttlEnv)
	// Если ттл пустой, вернём ошибку.
	if s == "" {
		return res, errors.New("no TTL_IN_SECONDS env")
	}
	// Пытаемся преобразовать ттл из строки в число.
	data, err = strconv.ParseUint(s, 10, 63)
	if err != nil {
		// Если возникла ошибка, обернём её доп. контекстом.
		return res, errors.Wrap(err, "TTL_IN_SECONDS env not correct")
	}
	// Сохраним ттл умножив его на секунды.
	res.timeToLive = time.Second * time.Duration(data)
	// Вернем конфиг без ошибок.
	return res, nil
}
