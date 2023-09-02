package main

import (
	"reflect"
	"testing"

	"github.com/valyala/fasthttp"
)

func Test_hand_handle(t *testing.T) {
	type args struct {
		ctx *fasthttp.RequestCtx
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "only one working case",
			args: args{
				ctx: &fasthttp.RequestCtx{},
			},
			want: body,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ha := hand{}
			ha.handle(tt.args.ctx)
			b := tt.args.ctx.Response.Body()
			if !reflect.DeepEqual(b, tt.want) {
				t.Errorf("\ngot = %v\nwant= %v", b, tt.want)
			}
		})
	}
}
