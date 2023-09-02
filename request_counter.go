package main

import (
	"sync"
	"time"

	"github.com/kpango/fastime"
)

var timeNow = fastime.Now

// mutex - Локальный интерфейс для RWMutex, чтобы можно было подпихнуть заглушку для тестов.
type mutex interface {
	// Locker - ну тут всё понятно.
	sync.Locker
	// RLock - метод для блокировки чтения.
	RLock()
	// RUnlock - метод для разблокировки чтения.
	RUnlock()
}

// requestCounter - счетчик для ай-пи адресов с определенным временем жизни записей
type ipCounter struct {
	// Мьютекс
	m mutex
	// Сложный список списков, строчное значение здесь это ай-пи адрес,
	// а второй список - список времён, в которое были запросы.
	// Удалять из списка проще, чем из слайса, когда удалять надо по значению,
	// хотя возможно в слайсе с учетом порядка тоже было бы не оч сложно.
	counter map[string]map[time.Time]struct{}
	// Время жизни записей
	ttl time.Duration
}

// newIPCounter - конструктор для счетчика ай-пи адресов
func newIPCounter(m mutex, ttl time.Duration) *ipCounter {
	// Создаем структуру, мьютекс и время жизни принимаем из вне, сложный список инициализируем сами.
	r := &ipCounter{m: m, counter: make(map[string]map[time.Time]struct{}), ttl: ttl}
	// Запускаем сборщик мусора. Он будет проверять время жизни записей и удалять старые.
	go r.runGC()
	return r
}

// countOfIP - показывает сколько раз такой ай-пи адрес был использован за последние ttl.
func (r *ipCounter) countOfIP(ip string) uint16 {
	// Блокируем мьютекс на чтение.
	r.m.RLock()
	// Читаем количество записей по-данному ай-пи.
	count := len(r.counter[ip])
	// Разблокируем мьютекс (defer жрёт время).
	r.m.RUnlock()
	// Возвращаем значение.
	return uint16(count)
}

// addIP - увеличивает счетчик для данного ай-пи адреса,
func (r *ipCounter) addIP(ip string) {
	// Блокируем запись.
	r.m.Lock()
	// Получаем список запросов для данного ай-пи адреса.
	list := r.counter[ip]
	// Если мы еще не создали список.
	if list == nil {
		// Тогда создадим новый и положим в него значение времени прямо сейчас.
		r.counter[ip] = map[time.Time]struct{}{timeNow(): {}}
		// Разблокируем запись и выходим.
		r.m.Unlock()
		return
	}
	// Если список для данного ай-пи уже был, тогда внесем в него еще одну запись от настоящего времени.
	r.counter[ip][timeNow()] = struct{}{}
	// Разблокируем запись, defer жрёт время.
	r.m.Unlock()
}

// dropDataByIP - удаляет все записи для ай-пи адреса.
func (r *ipCounter) dropDataByIP(ip string) {
	// Блокируем запись.
	r.m.Lock()
	// Для данного ай-пи адреса опустошаем список обращений.
	r.counter[ip] = map[time.Time]struct{}{}
	// Разблокируем запись.
	r.m.Unlock()
}

// runGC - garbage collector. Удаляет старые записи, примерно раз в секунду.
func (r *ipCounter) runGC() {
	// Запускаем бесконечный цикл
	for {
		// Запустим уборку мусора.
		r.gc()
		// Отдохнём секунду, чтобы не перегружать проц циклами.
		time.Sleep(time.Second)
	}
}

// gc - очистка мусора из всех списков.
func (r *ipCounter) gc() {
	// Заблокируем чтение.
	r.m.RLock()
	// Берем из списка ключ и значение под заблокированным чтением.
	for ip, times := range r.counter {
		// Разблокируем чтение.
		r.m.RUnlock()
		// Проходимся по всем записям в этом списке.
		for t := range times {
			// если время между "сейчас" и временем в записи больше чем ttl времени.
			dif := timeNow().Sub(t)
			if dif > r.ttl {
				// Тогда блокируем запись.
				r.m.Lock()
				// Удаляем это значение.
				delete(r.counter[ip], t)
				// Разблокируем запись.
				r.m.Unlock()
			}
		}
		// Здесь надо заблокировать чтение, чтобы идти дальше по циклу.
		r.m.RLock()
	}
	// Если цикл закончен, тогда снимаем блок на чтение.
	r.m.RUnlock()
}
