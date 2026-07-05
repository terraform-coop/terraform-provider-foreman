package generated

import (
	"context"
	"fmt"
	"net/url"
)

type ForemanPuppetClass struct {
	ForemanObject
	Name string `json:"name"`
}

func (c *ForemanClient) ReadForemanPuppetClass(ctx context.Context, id int) (*ForemanPuppetClass, error) {
	var resp ForemanPuppetClass
	err := c.Get(ctx, fmt.Sprintf("puppetclasses/%d", id), &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// QueryForemanPuppetClass looks up a puppet class by name. Unlike every
// other Foreman resource's index endpoint, puppetclasses groups its
// "results" by Puppet environment name (confirmed against a real server:
// {"results": {"production": [{"id":1,"name":"foo"}, ...], ...}}), not a
// flat array - the standard QueryResponse shape can't decode it at all.
func (c *ForemanClient) QueryForemanPuppetClass(ctx context.Context, name string) (*ForemanPuppetClass, error) {
	var response struct {
		Results map[string][]ForemanPuppetClass `json:"results"`
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
