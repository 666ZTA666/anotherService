package main

import (
	"net"
	"net/http"
	"testing"

	"github.com/valyala/fasthttp"
)

type handlerCounterStub struct {
	c bool
}

func newHandlerCounterStub() *handlerCounterStub {
	return &handlerCounterStub{}
}

func (h *handlerCounterStub) handle(*fasthttp.RequestCtx) {
	h.c = true
}

type counterStub struct {
	counter uint16
}

func newCounterStub(c uint16) *counterStub {
	return &counterStub{counter: c}
}

func (c *counterStub) countOfIP(string) uint16 {
	return c.counter
}

func (c *counterStub) addIP(string) {
	c.counter++
}

func Test_newLimiter(t *testing.T) {
	type args struct {
		hand   fasthttp.RequestHandler
		c      counter
		limit  uint16
		prefix uint8
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "bad hand case",
			args: args{
				hand:   nil,
				c:      nil,
				limit:  0,
				prefix: 0,
			},
			wantErr: true,
		},
		{
			name: "bad counter case",
			args: args{
				hand:   newHandlerCounterStub().handle,
				c:      nil,
				limit:  0,
				prefix: 0,
			},
			wantErr: true,
		},
		{
			name: "bad limit case",
			args: args{
				hand:   newHandlerCounterStub().handle,
				c:      newCounterStub(0),
				limit:  0,
				prefix: 0,
			},
			wantErr: true,
		},
		{
			name: "bad prefix case",
			args: args{
				hand:   newHandlerCounterStub().handle,
				c:      newCounterStub(0),
				limit:  1,
				prefix: 33,
			},
			wantErr: true,
		},
		{
			name: "success case",
			args: args{
				hand:   newHandlerCounterStub().handle,
				c:      newCounterStub(0),
				limit:  12,
				prefix: 32,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// проверить что функция правильная нет возможности, но можно хотя бы понять в каких случаях будет ошибка.
			_, err := newLimiter(tt.args.hand, tt.args.c, tt.args.limit, tt.args.prefix)
			if (err != nil) != tt.wantErr {
				t.Errorf("newLimiter() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func ctxWithIP(s string) *fasthttp.RequestCtx {
	ctx := &fasthttp.RequestCtx{}
	ctx.Request.Header.Set(http.CanonicalHeaderKey("X-Client-IP"), s)
	return ctx
}

func Test_limitHandlerDecorator_handle(t *testing.T) {
	type fields struct {
		c     counter
		limit uint16
		mask  net.IPMask
	}
	type args struct {
		ctx *fasthttp.RequestCtx
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantHandle bool
		wantCode   int
	}{
		{
			name: "not a limit case",
			fields: fields{
				c:     newCounterStub(0),
				limit: 2,
				mask:  net.IPMask{},
			},
			args: args{
				ctx: &fasthttp.RequestCtx{},
			},
			wantHandle: true,
			wantCode:   fasthttp.StatusOK,
		},
		{
			name: "bad ip case",
			fields: fields{
				c:     newCounterStub(0),
				limit: 2,
				mask:  net.IPMask{},
			},
			args: args{
				ctx: ctxWithIP("0:0:0:0:0:0"),
			},
			wantHandle: false,
			wantCode:   fasthttp.StatusBadRequest,
		},
		{
			name: "limit case",
			fields: fields{
				c:     newCounterStub(3),
				limit: 2,
				mask:  net.IPMask{},
			},
			args: args{
				ctx: &fasthttp.RequestCtx{},
			},
			wantHandle: false,
			wantCode:   fasthttp.StatusTooManyRequests,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerCounterStub()

			l := &limitHandlerDecorator{
				hand:  h.handle,
				c:     tt.fields.c,
				limit: tt.fields.limit,
				mask:  tt.fields.mask,
			}
			l.handle(tt.args.ctx)
			if tt.wantHandle != h.c {
				t.Errorf("decorated handler\ngot = %v\nwant= %v\n", h.c, tt.wantHandle)
				return
			}
			sc := tt.args.ctx.Response.StatusCode()
			if sc != tt.wantCode {
				t.Errorf("status code\ngot = %v\nwant= %v\n", sc, tt.wantCode)
			}
		})
	}
}
