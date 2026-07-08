package goforeman

import "net/url"

func parseURL(raw string) url.URL {
	u, _ := url.Parse(raw)
	return *u
}
