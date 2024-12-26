package dialer

import (
	"context"
	"testing"
	"time"
)

func TestGetDialer(t *testing.T) {
	tests := []struct {
		name    string
		ip      string
		timeout time.Duration
		ctx     context.Context
		network string
		addr    string
		wantErr bool
	}{
		{
			name:    "valid IP, no timeout",
			ip:      "8.8.8.8:53",
			timeout: 0,
			ctx:     context.Background(),
			network: "udp",
			addr:    "8.8.8.8:53",
			wantErr: false,
		},
		{
			name:    "invalid IP",
			ip:      "256.256.256.256:53",
			timeout: time.Second,
			ctx:     context.Background(),
			network: "udp",
			addr:    "256.256.256.256:53",
			wantErr: true,
		},
		{
			name:    "valid IP, with timeout",
			ip:      "8.8.8.8:53",
			timeout: time.Nanosecond,
			ctx:     context.Background(),
			network: "udp",
			addr:    "8.8.8.8:53",
			wantErr: false,
		},
		{
			name:    "timeout exceeded",
			ip:      "8.8.8.8:53",
			timeout: time.Millisecond,
			ctx: func() context.Context {
				ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
				cancel()
				return ctx
			}(),
			network: "udp",
			addr:    "8.8.8.8:53",
			wantErr: true,
		},
		{
			name:    "context canceled",
			ip:      "8.8.8.8:53",
			timeout: time.Second,
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			network: "udp",
			addr:    "8.8.8.8:53",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialerFunc := GetDialer(tt.ip, tt.timeout)
			_, err := dialerFunc(tt.ctx, tt.network, tt.addr)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDialer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
