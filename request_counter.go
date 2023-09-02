package main

import (
	"sync"
	"time"

	"github.com/kpango/fastime"
)

var timeNow = fastime.Now

// локальный интерфейс для RWMutex, чтобы можно было подпихнуть заглушку для тестов.
type mutex interface {
	// Locker ну тут всё понятно
	sync.Locker
	// RLock - метод для блокировки чтения
	RLock()
	// RUnlock - метод для разблокировки чтения
	RUnlock()
}

// requestCounter - счетчик для айпишников с определенным временем жизни записей
type ipCounter struct {
	// мьютекс
	m mutex
	// Сложная мапа мап, строчное значение здесь это айпишник,
	//а вторая мапа это список времён, в которое были запросы.
	// Удалять из мапы проще, чем из слайса, когда удалять надо по значению,
	// хотя возможно в слайсе с учетом порядка тоже было бы не оч сложно.
	counter map[string]map[time.Time]struct{}
	// Время жизни записей
	ttl time.Duration
}

// newIPCounter - конструктор для счетчика айпишников
func newIPCounter(m mutex, ttl time.Duration) *ipCounter {
	// создаем структуру, мьютекс и время жини принимаем из вне, сложную мапу инициализируем сами
	r := &ipCounter{m: m, counter: make(map[string]map[time.Time]struct{}), ttl: ttl}
	// Запускаем сборщик мусора. Он будет проверять время жизни записей и удалять старые.
	go r.runGC()
	return r
}

// countOfIP - показывает сколько раз такой айпишник был использован за последние ttl.
func (r *ipCounter) countOfIP(ip string) uint16 {
	// лочим мьютекс на чтение
	r.m.RLock()
	// читаем количество записей по данному айпи
	count := len(r.counter[ip])
	// разлочиваемся (defer жрёт время)
	r.m.RUnlock()
	// возвращаем значение
	return uint16(count)
}

// addIP - увеличивает счетчик для данного айпишника
func (r *ipCounter) addIP(ip string) {
	// лочимся на запись
	r.m.Lock()
	// получаем список запросов для данного айпишника
	list := r.counter[ip]
	// Если мы еще не создали мапу
	if list == nil {
		// тогда создадим новую и положим в неё значение времени прямо сейчас
		r.counter[ip] = map[time.Time]struct{}{timeNow(): {}}
		// разлочим запись и выходим
		r.m.Unlock()
		return
	}
	// Если список для данного айпи уже был, тогда внесем в него еще одну запись от настоящего времени
	r.counter[ip][timeNow()] = struct{}{}
	// разлочим запись, defer жрёт время.
	r.m.Unlock()
}

func (r *ipCounter) dropDataByIP(ip string) {
	r.m.Lock()
	r.counter[ip] = map[time.Time]struct{}{}
	r.m.Unlock()
}

// runGC - garbage collector удаляет старые записи примерно раз в секунду
func (r *ipCounter) runGC() {
	// запускаем бесконечный цикл
	for {
		// запустим уборку мусора
		r.gc()
		// отдохнём секунду, чтобы не перегружать проц циклами.
		time.Sleep(time.Second)
	}
}

func (r *ipCounter) gc() {
	// лочим чтение
	r.m.RLock()
	// берем из мапы ключ и значение под залоченным чтением
	for ip, times := range r.counter {
		// разлочиваем чтение
		r.m.RUnlock()
		// Проходимся по всем записям в этом списке
		for t := range times {
			// если время между "сейчас" и временем в записи больше чем ttl времени
			dif := timeNow().Sub(t)
			if dif > r.ttl {
				// тогда блочим запись
				r.m.Lock()
				// удаляем это значение
				delete(r.counter[ip], t)
				// разблокируем запись
				r.m.Unlock()
			}
		}
		// здесь надо заблочить чтение, чтобы идти дальше по циклу
		r.m.RLock()
	}
	// если цикл закончен, тогда снимаем лок на чтение
	r.m.RUnlock()
}
