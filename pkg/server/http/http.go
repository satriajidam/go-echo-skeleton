package http

import (
	"time"
)

var (
	CORSDefaultAllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"}
	// CORS safelisted-request-header: https://fetch.spec.whatwg.org/#cors-safelisted-request-header
	// CORS forbidden-header-name: https://fetch.spec.whatwg.org/#forbidden-header-name
	CORSDefaultAllowHeaders = []string{
		"Accept",
		"Accept-Charset",
		"Accept-Encoding",
		"Accept-Language",
		"Content-Language",
		"Content-Length",
		"Content-Type",
		"Host",
		"Origin",
	}
	CORSDefaultAllowCredentials = true
	CORSDefaultMaxAge           = 12 * time.Hour
)
