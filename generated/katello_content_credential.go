package generated

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

type ForemanKatelloContentCredentialRequest struct {
	Name    string `json:"name,omitempty"`
	Content string `json:"content,omitempty"`
}

type ForemanKatelloContentCredential struct {
	ForemanObject
	Content string `json:"content"`
}

func (c *ForemanClient) CreateForemanKatelloContentCredential(ctx context.Context, req *ForemanKatelloContentCredentialRequest) (*ForemanKatelloContentCredential, error) {
	var resp ForemanKatelloContentCredential
	err := c.Post(ctx, "katello/content_credentials", "content_credential", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *ForemanClient) ReadForemanKatelloContentCredential(ctx context.Context, id int) (*ForemanKatelloContentCredential, error) {
	var resp ForemanKatelloContentCredential
	err := c.Get(ctx, fmt.Sprintf("katello/content_credentials/%d", id), &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *ForemanClient) UpdateForemanKatelloContentCredential(ctx context.Context, id int, req *ForemanKatelloContentCredentialRequest) (*ForemanKatelloContentCredential, error) {
	var resp ForemanKatelloContentCredential
	err := c.Put(ctx, fmt.Sprintf("katello/content_credentials/%d", id), "content_credential", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *ForemanClient) DeleteForemanKatelloContentCredential(ctx context.Context, id int) error {
	return c.Delete(ctx, fmt.Sprintf("katello/content_credentials/%d", id))
}

func (c *ForemanClient) QueryForemanKatelloContentCredential(ctx context.Context, name string) (*ForemanKatelloContentCredential, error) {
	var response QueryResponse
	err := c.Get(ctx, fmt.Sprintf("katello/content_credentials?search=name=\"%s\"", url.QueryEscape(name)), &response)
	if err != nil {
		return nil, err
	}
	if len(response.Results) == 0 {
		return nil, nil
	}
	var obj ForemanKatelloContentCredential
	if err := json.Unmarshal(response.Results[0], &obj); err != nil {
		return nil, err
	}
	return &obj, nil
}
