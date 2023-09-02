package main

import (
	"reflect"
	"testing"
	"time"
)

var nTime = time.Unix(0, 0)

type mutexStub struct {
	rLock bool
	lock  bool
}

func newMutexStub() *mutexStub {
	return &mutexStub{}
}

func (m *mutexStub) Lock() {
	if m.lock || m.rLock {
		panic("lock")
	}
	m.lock = true
}

func (m *mutexStub) Unlock() {
	if m.rLock || !m.lock {
		panic("unlock")
	}
	m.lock = false
}

func (m *mutexStub) RLock() {
	if m.rLock || m.lock {
		panic("rlock")
	}
	m.rLock = true
}

func (m *mutexStub) RUnlock() {
	if !m.rLock || m.lock {
		panic("runlock")
	}
	m.rLock = false
}

func Test_ipCounter_countOfIP(t *testing.T) {
	type fields struct {
		m       mutex
		counter map[string]map[time.Time]struct{}
		ttl     time.Duration
	}
	type args struct {
		ip string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   uint16
	}{
		{
			name: "1",
			fields: fields{
				m: newMutexStub(),
				counter: map[string]map[time.Time]struct{}{
					"1": {nTime: {}},
				},
			},
			args: args{
				ip: "1",
			},
			want: 1,
		},
		{
			name: "0",
			fields: fields{
				m: newMutexStub(),
				counter: map[string]map[time.Time]struct{}{
					"2": {nTime: {}},
				},
			},
			args: args{
				ip: "1",
			},
			want: 0,
		},
		{
			name: "0",
			fields: fields{
				m:       newMutexStub(),
				counter: map[string]map[time.Time]struct{}{},
			},
			args: args{
				ip: "1",
			},
			want: 0,
		},
		{
			name: "2",
			fields: fields{
				m: newMutexStub(),
				counter: map[string]map[time.Time]struct{}{
					"1": {
						nTime:                  {},
						nTime.Add(time.Second): {},
					},
				},
			},
			args: args{
				ip: "1",
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ipCounter{
				m:       tt.fields.m,
				counter: tt.fields.counter,
				ttl:     tt.fields.ttl,
			}
			if got := r.countOfIP(tt.args.ip); got != tt.want {
				t.Errorf("countOfIP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ipCounter_addIP(t *testing.T) {
	type fields struct {
		m       mutex
		counter map[string]map[time.Time]struct{}
		ttl     time.Duration
	}
	type args struct {
		ip string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		timeNow func() time.Time
		want    ipCounter
	}{
		{
			name: "setter case",
			fields: fields{
				m:       newMutexStub(),
				counter: map[string]map[time.Time]struct{}{},
			},
			args: args{
				ip: "1",
			},
			timeNow: func() time.Time {
				return nTime
			},
			want: ipCounter{
				m: newMutexStub(),
				counter: map[string]map[time.Time]struct{}{
					"1": {nTime: {}},
				},
			},
		},
		{
			name: "adder case",
			fields: fields{
				m: newMutexStub(),
				counter: map[string]map[time.Time]struct{}{
					"1": {
						nTime: {},
					},
				},
			},
			args: args{
				ip: "1",
			},
			timeNow: func() time.Time {
				return nTime.Add(2)
			},
			want: ipCounter{
				m: newMutexStub(),
				counter: map[string]map[time.Time]struct{}{
					"1": {
						nTime:        {},
						nTime.Add(2): {},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ipCounter{
				m:       tt.fields.m,
				counter: tt.fields.counter,
				ttl:     tt.fields.ttl,
			}
			timeNow = tt.timeNow
			r.addIP(tt.args.ip)
			if !reflect.DeepEqual(*r, tt.want) {
				t.Errorf("\ngot = %#v\nwant= %#v", *r, tt.want)
			}
		})
	}
}

func Test_ipCounter_gc(t *testing.T) {
	type fields struct {
		m       mutex
		counter map[string]map[time.Time]struct{}
		ttl     time.Duration
	}
	tests := []struct {
		name    string
		fields  fields
		timeNow func() time.Time
		want    ipCounter
	}{
		{
			name: "first case",
			fields: fields{
				m: newMutexStub(),
				counter: map[string]map[time.Time]struct{}{
					"1": {
						nTime.Add(-2): {},
					},
				},
				ttl: time.Nanosecond,
			},
			timeNow: func() time.Time {
				return nTime
			},
			want: ipCounter{
				m: newMutexStub(),
				counter: map[string]map[time.Time]struct{}{
					"1": {},
				},
				ttl: time.Nanosecond,
			},
		},
		{
			name: "second case",
			fields: fields{
				m: newMutexStub(),
				counter: map[string]map[time.Time]struct{}{
					"1": {
						nTime.Add(-2): {},
						nTime.Add(-3): {},
					},
				},
				ttl: time.Nanosecond,
			},
			timeNow: func() time.Time {
				return nTime
			},
			want: ipCounter{
				m: newMutexStub(),
				counter: map[string]map[time.Time]struct{}{
					"1": {},
				},
				ttl: time.Nanosecond,
			},
		},
		{
			name: "third case",
			fields: fields{
				m: newMutexStub(),
				counter: map[string]map[time.Time]struct{}{
					"1": {
						nTime:         {},
						nTime.Add(-2): {},
						nTime.Add(-3): {},
					},
				},
				ttl: time.Nanosecond,
			},
			timeNow: func() time.Time {
				return nTime
			},
			want: ipCounter{
				m: newMutexStub(),
				counter: map[string]map[time.Time]struct{}{
					"1": {
						nTime: {},
					},
				},
				ttl: time.Nanosecond,
			},
		},
		{
			name: "fourth case",
			fields: fields{
				m: newMutexStub(),
				counter: map[string]map[time.Time]struct{}{
					"1": {
						nTime.Add(1):  {},
						nTime:         {},
						nTime.Add(-2): {},
						nTime.Add(-3): {},
					},
				},
				ttl: time.Nanosecond,
			},
			timeNow: func() time.Time {
				return nTime
			},
			want: ipCounter{
				m: newMutexStub(),
				counter: map[string]map[time.Time]struct{}{
					"1": {
						nTime.Add(1): {},
						nTime:        {},
					},
				},
				ttl: time.Nanosecond,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ipCounter{
				m:       tt.fields.m,
				counter: tt.fields.counter,
				ttl:     tt.fields.ttl,
			}
			timeNow = tt.timeNow
			r.gc()
			if !reflect.DeepEqual(*r, tt.want) {
				t.Errorf("\ngot = %#v\nwant= %#v", *r, tt.want)
			}
		})
	}
}

func Test_ipCounter_dropDataByIP(t *testing.T) {
	type fields struct {
		m       mutex
		counter map[string]map[time.Time]struct{}
		ttl     time.Duration
	}
	type args struct {
		ip string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   ipCounter
	}{
		{
			name: "deleting nil case",
			fields: fields{
				m: newMutexStub(),
				counter: map[string]map[time.Time]struct{}{
					"1": nil,
				},
			},
			args: args{
				ip: "1",
			},
			want: ipCounter{
				m: newMutexStub(),
				counter: map[string]map[time.Time]struct{}{
					"1": {},
				},
			},
		},
		{
			name: "deleting empty map case",
			fields: fields{
				m: newMutexStub(),
				counter: map[string]map[time.Time]struct{}{
					"1": {},
				},
			},
			args: args{
				ip: "1",
			},
			want: ipCounter{
				m: newMutexStub(),
				counter: map[string]map[time.Time]struct{}{
					"1": {},
				},
			},
		},
		{
			name: "deleting non empty map case",
			fields: fields{
				m: newMutexStub(),
				counter: map[string]map[time.Time]struct{}{
					"1": {
						nTime:        {},
						nTime.Add(1): {},
					},
				},
			},
			args: args{
				ip: "1",
			},
			want: ipCounter{
				m: newMutexStub(),
				counter: map[string]map[time.Time]struct{}{
					"1": {},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ipCounter{
				m:       tt.fields.m,
				counter: tt.fields.counter,
				ttl:     tt.fields.ttl,
			}
			r.dropDataByIP(tt.args.ip)
			if !reflect.DeepEqual(*r, tt.want) {
				t.Errorf("\ngot = %v\nwant= %v", *r, tt.want)
			}
		})
	}
}
