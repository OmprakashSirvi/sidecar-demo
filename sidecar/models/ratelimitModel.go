package models

import "golang.org/x/time/rate"

type RateLimiter struct {
	Limit *rate.Limiter
	Type  string
}
