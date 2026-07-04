package generated

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// ForemanComputeProfile is hand-written, not apidoc-driven: a compute
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

type ForemanComputeProfileRequest struct {
	Name string `json:"name,omitempty"`
}

type ForemanComputeProfile struct {
	ForemanObject
	Name              string                     `json:"name"`
	ComputeAttributes []*ForemanComputeAttribute `json:"compute_attributes,omitempty"`
}

// ForemanComputeAttribute is one element of a compute profile's per-compute-resource
// VM sizing attributes. VMAttrs is a free-form, compute-resource-provider-specific
// hash (e.g. cpus/memory_mb for VMware, cpus/memory for Libvirt), carried as opaque
// JSON.
type ForemanComputeAttribute struct {
	ForemanObject
	ComputeResourceID int             `json:"compute_resource_id"`
	VMAttrs           json.RawMessage `json:"vm_attrs,omitempty"`
}

func (c *ForemanClient) CreateForemanComputeProfile(ctx context.Context, req *ForemanComputeProfileRequest) (*ForemanComputeProfile, error) {
	var resp ForemanComputeProfile
	err := c.Post(ctx, "compute_profiles", "compute_profile", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *ForemanClient) ReadForemanComputeProfile(ctx context.Context, id int) (*ForemanComputeProfile, error) {
	var resp ForemanComputeProfile
	err := c.Get(ctx, fmt.Sprintf("compute_profiles/%d", id), &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *ForemanClient) UpdateForemanComputeProfile(ctx context.Context, id int, req *ForemanComputeProfileRequest) (*ForemanComputeProfile, error) {
	var resp ForemanComputeProfile
	err := c.Put(ctx, fmt.Sprintf("compute_profiles/%d", id), "compute_profile", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *ForemanClient) DeleteForemanComputeProfile(ctx context.Context, id int) error {
	return c.Delete(ctx, fmt.Sprintf("compute_profiles/%d", id))
}

func (c *ForemanClient) QueryForemanComputeProfile(ctx context.Context, name string) (*ForemanComputeProfile, error) {
	var response QueryResponse
	err := c.Get(ctx, fmt.Sprintf("compute_profiles?search=name=\"%s\"", url.QueryEscape(name)), &response)
	if err != nil {
		return nil, err
	}
	if len(response.Results) == 0 {
		return nil, nil
	}
	var obj ForemanComputeProfile
	if err = json.Unmarshal(response.Results[0], &obj); err != nil {
		return nil, err
	}
	return &obj, nil
}

type foremanComputeAttributeRequest struct {
	VMAttrs json.RawMessage `json:"vm_attrs,omitempty"`
}

func (c *ForemanClient) CreateForemanComputeAttribute(ctx context.Context, profileID, computeResourceID int, vmAttrs json.RawMessage) (*ForemanComputeAttribute, error) {
	endpoint := fmt.Sprintf("compute_profiles/%d/compute_resources/%d/compute_attributes", profileID, computeResourceID)
	req := &foremanComputeAttributeRequest{VMAttrs: vmAttrs}
	var resp ForemanComputeAttribute
	if err := c.Post(ctx, endpoint, "compute_attribute", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *ForemanClient) UpdateForemanComputeAttribute(ctx context.Context, profileID, computeResourceID, attributeID int, vmAttrs json.RawMessage) (*ForemanComputeAttribute, error) {
	endpoint := fmt.Sprintf("compute_profiles/%d/compute_resources/%d/compute_attributes/%d", profileID, computeResourceID, attributeID)
	req := &foremanComputeAttributeRequest{VMAttrs: vmAttrs}
	var resp ForemanComputeAttribute
	if err := c.Put(ctx, endpoint, "compute_attribute", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *ForemanClient) DeleteForemanComputeAttribute(ctx context.Context, profileID, computeResourceID, attributeID int) error {
	endpoint := fmt.Sprintf("compute_profiles/%d/compute_resources/%d/compute_attributes/%d", profileID, computeResourceID, attributeID)
	return c.Delete(ctx, endpoint)
}
