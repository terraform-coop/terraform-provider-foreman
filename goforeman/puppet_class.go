package goforeman

import (
	"context"
	"fmt"
	"net/url"
)

type PuppetClass struct {
	Base
	Name string `json:"name"`
}

func (c *Client) ReadPuppetClass(ctx context.Context, id int) (*PuppetClass, error) {
	var resp PuppetClass
	err := c.Get(ctx, fmt.Sprintf("puppetclasses/%d", id), &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// QueryPuppetClass looks up a puppet class by name. Unlike every
// other Foreman resource's index endpoint, puppetclasses groups its
// "results" by Puppet environment name (confirmed against a real server:
// {"results": {"production": [{"id":1,"name":"foo"}, ...], ...}}), not a
// flat array - the standard QueryResponse shape can't decode it at all.
func (c *Client) FindPuppetClassByName(ctx context.Context, name string) (*PuppetClass, error) {
	var response struct {
		Results map[string][]PuppetClass `json:"results"`
	}
	err := c.Get(ctx, fmt.Sprintf("puppetclasses?search=name=\"%s\"", url.QueryEscape(name)), &response)
	if err != nil {
		return nil, err
	}
	for _, classes := range response.Results {
		for _, class := range classes {
			if class.Name == name {
				return &class, nil
			}
		}
	}
	return nil, nil
}
