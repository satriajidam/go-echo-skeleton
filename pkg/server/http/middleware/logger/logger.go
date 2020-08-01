package logger

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/satriajidam/go-echo-skeleton/pkg/log"
	"github.com/satriajidam/go-echo-skeleton/pkg/server/http/middleware/requestid"
)

// Config defines the config for logger middleware
type Config struct {
	Stdout *zerolog.Logger
	Stderr *zerolog.Logger
	// UTC a boolean stating whether to use UTC time zone or local.
	UTC      bool
	Routes   []Route
	SkipPath []string
}

type Route struct {
	Method       string
	RelativePath string
	LogPayload   bool
}

type logFields struct {
	requestID string
	status    int
	method    string
	path      string
	clientIP  string
	host      string
	latency   time.Duration
	userAgent string
	payload   string
}

func createDumplogger(logger *zerolog.Logger, fields logFields) zerolog.Logger {
	return logger.With().
		Str("requestID", fields.requestID).
		Int("status", fields.status).
		Str("method", fields.method).
		Str("path", fields.path).
		Str("clientIP", fields.clientIP).
		Dur("latency", fields.latency).
		Str("userAgent", fields.userAgent).
		Str("payload", fields.payload).
		Logger()
}

func pathKey(method, path string) string {
	return fmt.Sprintf("%s %s", method, path)
}

// New initializes the logging middleware.
func New(port string, config ...Config) echo.MiddlewareFunc {
	var newConfig Config
	if len(config) > 0 {
		newConfig = config[0]
	}

	var skipped map[string]struct{}
	if length := len(newConfig.SkipPath); length > 0 {
		skipped = make(map[string]struct{}, length)
		for _, p := range newConfig.SkipPath {
			skipped[p] = struct{}{}
		}
	}

	var logged map[string]bool
	if length := len(newConfig.Routes); length > 0 {
		logged = make(map[string]bool, length)
		for _, p := range newConfig.Routes {
			logged[pathKey(p.Method, p.RelativePath)] = p.LogPayload
		}
	}

	var stdout *zerolog.Logger
	if newConfig.Stdout == nil {
		stdout = log.Stdout()
	} else {
		stdout = newConfig.Stdout
	}

	var stderr *zerolog.Logger
	if newConfig.Stderr == nil {
		stderr = log.Stderr()
	} else {
		stderr = newConfig.Stderr
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			start := time.Now()
			requestID := ctx.Request().Header.Get(requestid.HeaderXRequestID)
			method := ctx.Request().Method
			routePath := ctx.Path()
			path := ctx.Request().URL.Path
			raw := ctx.Request().URL.RawQuery
			if raw != "" {
				path = path + "?" + raw
			}

			payload := "-"
			if yes, ok := logged[pathKey(method, routePath)]; ok && yes {
				var buf bytes.Buffer
				tee := io.TeeReader(ctx.Request().Body, &buf)
				body, _ := ioutil.ReadAll(tee)
				ctx.Request().Body = ioutil.NopCloser(&buf)
				payload = string(body)
			}

			var err error
			if err = next(ctx); err != nil {
				ctx.Error(err)
			}

			track := true
			if _, ok := skipped[routePath]; ok {
				track = false
			}

			if track {
				end := time.Now()
				latency := end.Sub(start)
				if newConfig.UTC {
					end = end.UTC()
				}

				if err == nil &&
					ctx.Response().Status >= http.StatusBadRequest &&
					ctx.Response().Status <= http.StatusNetworkAuthenticationRequired {
					err = fmt.Errorf(http.StatusText(ctx.Response().Status))
				}

				msg := fmt.Sprintf("HTTP request to port %s", port)

				fields := logFields{
					requestID: requestID,
					status:    ctx.Response().Status,
					method:    ctx.Request().Method,
					path:      path,
					clientIP:  ctx.RealIP(),
					host:      ctx.Request().Host,
					latency:   latency,
					userAgent: ctx.Request().UserAgent(),
					payload:   payload,
				}

				dumpStdout := createDumplogger(stdout, fields)
				dumpStderr := createDumplogger(stderr, fields)

				switch {
				case ctx.Response().Status >= http.StatusBadRequest &&
					ctx.Response().Status < http.StatusInternalServerError:
					{
						dumpStdout.Warn().Timestamp().Err(err).Msg(msg)
					}
				case ctx.Response().Status >= http.StatusInternalServerError:
					{
						dumpStderr.Error().Timestamp().Err(err).Msg(msg)
					}
				default:
					dumpStdout.Info().Timestamp().Msg(msg)
				}
			}

			return err
		}
	}
}
