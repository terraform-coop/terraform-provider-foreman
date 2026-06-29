package generated

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"strconv"
)

type ForemanKatelloProductRequest struct {
	Name            string `json:"name,omitempty"`
	Description     string `json:"description,omitempty"`
	Label           string `json:"label,omitempty"`
	GpgKeyID        int    `json:"gpg_key_id,omitempty"`
	SslCaCertID     int    `json:"ssl_ca_cert_id,omitempty"`
	SslClientCertID int    `json:"ssl_client_cert_id,omitempty"`
	SslClientKeyID  int    `json:"ssl_client_key_id,omitempty"`
	SyncPlanID      int    `json:"sync_plan_id,omitempty"`
}

type ForemanKatelloProduct struct {
	ForemanObject
	Description     string `json:"description"`
	Label           string `json:"label"`
	GpgKeyID        int    `json:"gpg_key_id"`
	SslCaCertID     int    `json:"ssl_ca_cert_id"`
	SslClientCertID int    `json:"ssl_client_cert_id"`
	SslClientKeyID  int    `json:"ssl_client_key_id"`
	SyncPlanID      int    `json:"sync_plan_id"`
}

func (c *ForemanClient) CreateForemanKatelloProduct(ctx context.Context, req *ForemanKatelloProductRequest) (*ForemanKatelloProduct, error) {
	var resp ForemanKatelloProduct
	orgID := c.config.OrganizationID
	err := c.Post(ctx, fmt.Sprintf("katello/products?organization_id=%d", orgID), "product", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *ForemanClient) ReadForemanKatelloProduct(ctx context.Context, id int) (*ForemanKatelloProduct, error) {
	var resp ForemanKatelloProduct
	err := c.Get(ctx, fmt.Sprintf("katello/products/%d?organization_id=%d", id, c.config.OrganizationID), &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *ForemanClient) UpdateForemanKatelloProduct(ctx context.Context, id int, req *ForemanKatelloProductRequest) (*ForemanKatelloProduct, error) {
	var resp ForemanKatelloProduct
	err := c.Put(ctx, fmt.Sprintf("katello/products/%d", id), "product", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *ForemanClient) DeleteForemanKatelloProduct(ctx context.Context, id int) error {
	return c.Delete(ctx, fmt.Sprintf("katello/products/%d", id))
}

func (c *ForemanClient) QueryForemanKatelloProduct(ctx context.Context, name string) (*ForemanKatelloProduct, error) {
	var response QueryResponse
	orgID := strconv.Itoa(c.config.OrganizationID)
	err := c.Get(ctx, fmt.Sprintf("katello/products?search=name=\"%s\"&organization_id=%s", url.QueryEscape(name), orgID), &response)
	if err != nil {
		return nil, err
	}
	if len(response.Results) == 0 {
		return nil, nil
	}
	var obj ForemanKatelloProduct
	if err := json.Unmarshal(response.Results[0], &obj); err != nil {
		return nil, err
	}
	return &obj, nil
}
