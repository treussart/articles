package main

import (
	"bytes"
	"fmt"
	"slices"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	writer, err := w.ResponseWriter.Write(b)
	if err != nil {
		return writer, fmt.Errorf("bodyLogWriter.ResponseWriter.Write: %w", err)
	}
	return writer, nil
}

// NewLogger creates and returns a gin.HandlerFunc for logging with optional metrics support.
func NewLogger(l zerolog.Logger, packageName string) (gin.HandlerFunc, error) {
	return loggerMiddleware{
		logger:      l,
		packageName: packageName,
	}.handle, nil
}

type loggerMiddleware struct {
	logger      zerolog.Logger
	packageName string
}

func (l loggerMiddleware) handle(c *gin.Context) {
	blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
	c.Writer = blw
	path := c.Request.URL.Path
	var log = l.logger.With().Logger()
	log.UpdateContext(func(lCtx zerolog.Context) zerolog.Context {
		return lCtx.Str("net.ip.source", c.ClientIP()).
			Str("http.path", path).
			Str("http.method", c.Request.Method).
			Str("http.proto", c.Request.Proto)
	})

	// associate main logger with gin context
	ctxWithLogger := log.WithContext(c.Request.Context())
	// Replace gin request context by the new one associated with the logger
	c.Request = c.Request.WithContext(ctxWithLogger)

	// Start timer
	start := time.Now()
	// Process request
	c.Next()
	duration := time.Since(start)
	statusCode := c.Writer.Status()
	logFromGinCtx := zerolog.Ctx(c.Request.Context())
	logFromGinCtx.UpdateContext(func(zlogCtx zerolog.Context) zerolog.Context {
		return zlogCtx.Dur("http.duration_seconds", duration).
			Int("http.status_code", c.Writer.Status())
	})

	pathsExcluded := []string{"/health", "/metrics", "/ready"}
	level := l.levelForStatus(statusCode, path, pathsExcluded)
	if level == zerolog.ErrorLevel {
		logFromGinCtx.Error().Str("error", blw.body.String()).Send()
	} else {
		logFromGinCtx.WithLevel(level).Send()
	}
}

func (l loggerMiddleware) levelForStatus(statusCode int, path string, pathsExcluded []string) zerolog.Level {
	level := zerolog.InfoLevel
	switch {
	case 400 <= statusCode && statusCode < 500:
		level = zerolog.WarnLevel
	case 500 <= statusCode:
		level = zerolog.ErrorLevel
	default:
		if slices.Contains(pathsExcluded, path) {
			level = zerolog.DebugLevel
		}
	}
	return level
}
