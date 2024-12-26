package main

import (
	"net/http"
	"runtime"

	"github.com/rs/zerolog"
)

func recovery(next http.Handler, logger zerolog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				stack := make([]byte, 8096)
				stack = stack[:runtime.Stack(stack, false)]
				logger.Error().Interface("error", err).Bytes("stack", stack).Msg("Recovered from panic")
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
