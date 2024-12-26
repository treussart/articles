package dialer

import (
	"context"
	"fmt"
	"net"
	"time"
)

// GetDialer creates a custom dialer function with specified IP and timeout, returning a function to dial contexts.
func GetDialer(ip string, timeout time.Duration) func(ctx context.Context, network string, addr string) (net.Conn, error) {
	dialer := &net.Dialer{
		Resolver: &net.Resolver{
			Dial: func(ctx context.Context, _, _ string) (net.Conn, error) {
				d := net.Dialer{
					Timeout: timeout,
				}
				conn, err := d.DialContext(ctx, "udp", ip)
				if err != nil {
					return nil, fmt.Errorf("d.DialContext: %w", err)
				}
				return conn, nil
			},
		},
	}
	return dialer.DialContext
}
