package goforeman

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// ComputeProfile is hand-written, not apidoc-driven: a compute
// profile's own create/update body only ever has "name" (confirmed against
// apidoc/v2.json), but each profile also owns a set of per-compute-resource
// "compute attributes" (VM sizing: cpus, memory, disks, ...), which Foreman
// models as a SEPARATE nested resource
// (/api/compute_profiles/:id/compute_resources/:compute_resource_id/compute_attributes)
// scoped by two parents at once - a shape the generic apidoc-driven
// pipeline's single-parent ParentEndpoint mechanism can't express. The old
// SDKv2 provider modeled this as a list attribute on the compute_profile
// resource, orchestrating one API call per element; this hand-written file
// restores that behavior.

type ComputeProfileRequest struct {
	Name string `json:"name,omitempty"`
}

type ComputeProfile struct {
	Base
	Name              string              `json:"name"`
	ComputeAttributes []*ComputeAttribute `json:"compute_attributes,omitempty"`
}

// ComputeAttribute is one element of a compute profile's per-compute-resource
// VM sizing attributes. VMAttrs is a free-form, compute-resource-provider-specific
// hash (e.g. cpus/memory_mb for VMware, cpus/memory for Libvirt), carried as opaque
// JSON.
type ComputeAttribute struct {
	Base
	ComputeResourceID int             `json:"compute_resource_id"`
	VMAttrs           json.RawMessage `json:"vm_attrs,omitempty"`
}

func (c *Client) CreateComputeProfile(ctx context.Context, req *ComputeProfileRequest) (*ComputeProfile, error) {
	var resp ComputeProfile
	err := c.Post(ctx, "compute_profiles", "compute_profile", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) ReadComputeProfile(ctx context.Context, id int) (*ComputeProfile, error) {
	var resp ComputeProfile
	err := c.Get(ctx, fmt.Sprintf("compute_profiles/%d", id), &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) UpdateComputeProfile(ctx context.Context, id int, req *ComputeProfileRequest) (*ComputeProfile, error) {
	var resp ComputeProfile
	err := c.Put(ctx, fmt.Sprintf("compute_profiles/%d", id), "compute_profile", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) DeleteComputeProfile(ctx context.Context, id int) error {
	return c.Delete(ctx, fmt.Sprintf("compute_profiles/%d", id))
}

func (c *Client) FindComputeProfileByName(ctx context.Context, name string) (*ComputeProfile, error) {
	var response QueryResponse
	err := c.Get(ctx, fmt.Sprintf("compute_profiles?search=name=\"%s\"", url.QueryEscape(name)), &response)
	if err != nil {
		return nil, err
	}
	if len(response.Results) == 0 {
		return nil, nil
	}
	var obj ComputeProfile
	if err = json.Unmarshal(response.Results[0], &obj); err != nil {
		return nil, err
	}
	return &obj, nil
}

type foremanComputeAttributeRequest struct {
	VMAttrs json.RawMessage `json:"vm_attrs,omitempty"`
}

func (c *Client) CreateComputeAttribute(ctx context.Context, profileID, computeResourceID int, vmAttrs json.RawMessage) (*ComputeAttribute, error) {
	endpoint := fmt.Sprintf("compute_profiles/%d/compute_resources/%d/compute_attributes", profileID, computeResourceID)
	req := &foremanComputeAttributeRequest{VMAttrs: vmAttrs}
	var resp ComputeAttribute
	if err := c.Post(ctx, endpoint, "compute_attribute", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) UpdateComputeAttribute(ctx context.Context, profileID, computeResourceID, attributeID int, vmAttrs json.RawMessage) (*ComputeAttribute, error) {
	endpoint := fmt.Sprintf("compute_profiles/%d/compute_resources/%d/compute_attributes/%d", profileID, computeResourceID, attributeID)
	req := &foremanComputeAttributeRequest{VMAttrs: vmAttrs}
	var resp ComputeAttribute
	if err := c.Put(ctx, endpoint, "compute_attribute", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) DeleteComputeAttribute(ctx context.Context, profileID, computeResourceID, attributeID int) error {
	endpoint := fmt.Sprintf("compute_profiles/%d/compute_resources/%d/compute_attributes/%d", profileID, computeResourceID, attributeID)
	return c.Delete(ctx, endpoint)
}
