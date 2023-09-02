package main

import (
	"os"
	"reflect"
	"testing"
	"time"
)

func setLimit(s string) {
	_ = os.Setenv(limitEnv, s)
}

func setPrefix(s string) {
	_ = os.Setenv(prefixEnv, s)
}

func setTTL(s string) {
	_ = os.Setenv(ttlEnv, s)
}

func Test_getEnvs(t *testing.T) {
	tests := []struct {
		name    string
		prefix  string
		limit   string
		ttl     string
		want    envs
		wantErr bool
	}{
		{
			// енвы не проставили вообще
			name:   "0",
			prefix: "",
			limit:  "",
			ttl:    "",
			want: envs{
				prefix:     0,
				limit:      0,
				timeToLive: 0,
			},
			wantErr: true,
		},
		{
			// есть енва с префиксом, но в ней текст, а не число
			name:   "1",
			prefix: "prefix",
			limit:  "",
			ttl:    "",
			want: envs{
				prefix:     0,
				limit:      0,
				timeToLive: 0,
			},
			wantErr: true,
		},
		{
			// есть енва с префиксом, но других нет
			name:   "2",
			prefix: "1",
			limit:  "",
			ttl:    "",
			want: envs{
				prefix:     1,
				limit:      0,
				timeToLive: 0,
			},
			wantErr: true,
		},
		{
			// есть енва с префиксом, в ней слэш, но это рабочий кейс
			name:   "3",
			prefix: "/1",
			limit:  "",
			ttl:    "",
			want: envs{
				prefix:     1,
				limit:      0,
				timeToLive: 0,
			},
			wantErr: true,
		},
		{
			// есть енва с лимитом, но там не число
			name:   "4",
			prefix: "/1",
			limit:  "aboba",
			ttl:    "",
			want: envs{
				prefix:     1,
				limit:      0,
				timeToLive: 0,
			},
			wantErr: true,
		},
		{
			// есть енва с лимитом, но нет с ттл
			name:   "5",
			prefix: "/1",
			limit:  "1",
			ttl:    "",
			want: envs{
				prefix:     1,
				limit:      1,
				timeToLive: 0,
			},
			wantErr: true,
		},
		{
			// есть енва с ттл, но там не число
			name:   "6",
			prefix: "/1",
			limit:  "1",
			ttl:    "aboba",
			want: envs{
				prefix:     1,
				limit:      1,
				timeToLive: 0,
			},
			wantErr: true,
		},
		{
			// всё енвы на месте, всё окей)
			name:   "7",
			prefix: "/1",
			limit:  "1",
			ttl:    "1",
			want: envs{
				prefix:     1,
				limit:      1,
				timeToLive: time.Second,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setTTL(tt.ttl)
			setPrefix(tt.prefix)
			setLimit(tt.limit)
			got, err := getEnvs()
			if (err != nil) != tt.wantErr {
				t.Errorf("getEnvs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getEnvs() got = %v, want %v", got, tt.want)
			}
		})
	}
}
