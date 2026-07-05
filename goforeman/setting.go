package goforeman

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

type ForemanSettingRequest struct {
	Value string `json:"value,omitempty"`
}

// ForemanSetting does not embed ForemanObject: unlike most Foreman
// entities, a setting's own "id" is its key (e.g.
// "append_domain_name_for_hosts"), not a server-assigned integer -
// confirmed against a real Foreman 1.11 API response. Value is decoded as
// json.RawMessage, not string, for the same reason as common_parameters/
// parameters' "value": settings are user-typed (that same fixture shows a
// JSON boolean, not a string), so a plain string field fails
// json.Unmarshal for the whole struct whenever the setting isn't a string.
type ForemanSetting struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	CreatedAt string          `json:"created_at"`
	UpdatedAt string          `json:"updated_at"`
	Value     json.RawMessage `json:"value,omitempty"`
}

func (c *ForemanClient) ReadForemanSetting(ctx context.Context, id string) (*ForemanSetting, error) {
	var resp ForemanSetting
	err := c.Get(ctx, fmt.Sprintf("settings/%s", url.PathEscape(id)), &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *ForemanClient) UpdateForemanSetting(ctx context.Context, id string, req *ForemanSettingRequest) (*ForemanSetting, error) {
	var resp ForemanSetting
	err := c.Put(ctx, fmt.Sprintf("settings/%s", url.PathEscape(id)), "setting", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *ForemanClient) QueryForemanSetting(ctx context.Context, name string) (*ForemanSetting, error) {
	var response QueryResponse
	err := c.Get(ctx, fmt.Sprintf("settings?search=name=\"%s\"", url.QueryEscape(name)), &response)
	if err != nil {
		return nil, err
	}
	if len(response.Results) == 0 {
		return nil, nil
	}
	var obj ForemanSetting
	if err = json.Unmarshal(response.Results[0], &obj); err != nil {
		return nil, err
	}
	return &obj, nil
}
