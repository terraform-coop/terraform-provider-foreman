package goforeman

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

type SettingRequest struct {
	Value string `json:"value,omitempty"`
}

// Setting does not embed Base: unlike most Foreman
// entities, a setting's own "id" is its key (e.g.
// "append_domain_name_for_hosts"), not a server-assigned integer -
// confirmed against a real Foreman 1.11 API response. Value is decoded as
// json.RawMessage, not string, for the same reason as common_parameters/
// parameters' "value": settings are user-typed (that same fixture shows a
// JSON boolean, not a string), so a plain string field fails
// json.Unmarshal for the whole struct whenever the setting isn't a string.
type Setting struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	CreatedAt string          `json:"created_at"`
	UpdatedAt string          `json:"updated_at"`
	Value     json.RawMessage `json:"value,omitempty"`
}

func (c *Client) ReadSetting(ctx context.Context, id string) (*Setting, error) {
	var resp Setting
	err := c.Get(ctx, fmt.Sprintf("settings/%s", url.PathEscape(id)), &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) UpdateSetting(ctx context.Context, id string, req *SettingRequest) (*Setting, error) {
	var resp Setting
	err := c.Put(ctx, fmt.Sprintf("settings/%s", url.PathEscape(id)), "setting", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) FindSettingByName(ctx context.Context, name string) (*Setting, error) {
	var response QueryResponse
	err := c.Get(ctx, fmt.Sprintf("settings?search=name=\"%s\"", url.QueryEscape(name)), &response)
	if err != nil {
		return nil, err
	}
	if len(response.Results) == 0 {
		return nil, nil
	}
	var obj Setting
	if err = json.Unmarshal(response.Results[0], &obj); err != nil {
		return nil, err
	}
	return &obj, nil
}
