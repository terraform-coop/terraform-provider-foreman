package generated

import (
	"context"
	"net/url"
)

func parseURL(raw string) url.URL {
	u, _ := url.Parse(raw)
	return *u
}

func ctx() context.Context {
	return context.Background()
}
