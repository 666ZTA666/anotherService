package main

import (
	"testing"

	"github.com/valyala/fasthttp"
)

type dropperStub struct{}

func newDropperStub() *dropperStub {
	return &dropperStub{}
}

func (d dropperStub) dropDataByIP(string) {}

func ctxWithPrefix(s string) *fasthttp.RequestCtx {
	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("https://google.gom/?prefix=" + s)
	return ctx
}

func Test_dropLimitHandler_handle(t *testing.T) {
	type fields struct {
		d dropper
	}
	type args struct {
		ctx *fasthttp.RequestCtx
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantCode int
	}{
		{
			name: "ok case",
			fields: fields{
				d: newDropperStub(),
			},
			args: args{
				ctx: ctxWithPrefix("24"),
			},
			wantCode: fasthttp.StatusOK,
		},
		{
			name: "",
			fields: fields{
				d: newDropperStub(),
			},
			args: args{
				ctx: ctxWithIP("0:0:0:0:0:1"),
			},
			wantCode: fasthttp.StatusBadRequest,
		},
		{
			name: "bad query case",
			fields: fields{
				d: newDropperStub(),
			},
			args: args{
				ctx: ctxWithPrefix("33"),
			},
			wantCode: fasthttp.StatusBadRequest,
		},
		{
			name: "empty query case",
			fields: fields{
				d: newDropperStub(),
			},
			args: args{
				ctx: ctxWithPrefix(""),
			},
			wantCode: fasthttp.StatusBadRequest,
		},
		{
			name: "empty request string case",
			fields: fields{
				d: newDropperStub(),
			},
			args: args{
				ctx: &fasthttp.RequestCtx{},
			},
			wantCode: fasthttp.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &dropLimitHandler{
				d: tt.fields.d,
			}
			d.handle(tt.args.ctx)
			sc := tt.args.ctx.Response.StatusCode()
			if sc != tt.wantCode {
				t.Errorf("\ngot = %v\nwant= %v", sc, tt.wantCode)
			}
		})
	}
}
