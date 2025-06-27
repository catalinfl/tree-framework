package tree

import (
	"slices"
	"strconv"
	"strings"
)

type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           int
	Skipper          func(*Ctx) bool
}

func CORS(config ...CORSConfig) func(*Ctx) error {
	cfg := UseDefaultCORS()
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(c *Ctx) error {
		if cfg.Skipper != nil && cfg.Skipper(c) {
			return c.Next()
		}

		origin := c.Header().Get("Origin")
		method := c.GetMethod()
		if origin == "" {
			return c.Next()
		}

		c.SetHeader("Vary", "Origin")

		if len(cfg.AllowOrigins) > 0 {
			if slices.Contains(cfg.AllowOrigins, "*") {
				c.SetHeader("Access-Control-Allow-Origin", "*")
			} else if slices.Contains(cfg.AllowOrigins, origin) {
				c.SetHeader("Access-Control-Allow-Origin", origin)
			} else {
				return c.Next()
			}
		}

		if cfg.AllowCredentials {
			c.SetHeader("Access-Control-Allow-Credentials", "true")
		}

		if len(cfg.ExposeHeaders) > 0 {
			c.SetHeader("Access-Control-Expose-Headers", strings.Join(cfg.ExposeHeaders, ", "))
		}

		if method == "OPTIONS" {
			if len(cfg.AllowMethods) > 0 {
				c.SetHeader("Access-Control-Allow-Methods", strings.Join(cfg.AllowMethods, ", "))
			}

			if len(cfg.AllowHeaders) > 0 {
				c.SetHeader("Access-Control-Allow-Headers", strings.Join(cfg.AllowHeaders, ", "))
			}

			if cfg.MaxAge > 0 {
				c.SetHeader("Access-Control-Max-Age", strconv.Itoa(cfg.MaxAge))
			}

			return c.Status(200)
		}

		return c.Next()
	}
}

func SkipCORSPath(paths ...string) func(*Ctx) bool {
	return func(c *Ctx) bool {
		return slices.Contains(paths, c.Path())
	}
}

func SkipSameOrigin(c *Ctx) bool {
	origin := c.Header().Get("Origin")
	host := c.Header().Get("Host")
	return origin == "" || origin == "http://"+host || origin == "https://"+host
}

func UseDefaultCORS() CORSConfig {
	return CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Accept-Language", "Content-Language", "Content-Type", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{},
		AllowCredentials: false,
		MaxAge:           86400,
		Skipper:          SkipSameOrigin,
	}
}
