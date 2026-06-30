package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/terraform-coop/terraform-provider-foreman/generated"
)

// hostgroupTestServer returns an httptest server that mimics the Foreman
// hostgroup endpoints used by the hand-written resource. It is intended as a
// pattern for resource-level unit tests that do not require a real Foreman
// instance.
func hostgroupTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			assert.Equal(t, "/api/hostgroups", r.URL.Path)
			writeHostgroupResponse(t, w, 42)
		case http.MethodGet:
			assert.Equal(t, "/api/hostgroups/42", r.URL.Path)
			writeHostgroupResponse(t, w, 42)
		case http.MethodPut:
			assert.Equal(t, "/api/hostgroups/42", r.URL.Path)
			writeHostgroupResponse(t, w, 42)
		case http.MethodDelete:
			assert.Equal(t, "/api/hostgroups/42", r.URL.Path)
			w.WriteHeader(http.StatusOK)
		}
	}))
}

func writeHostgroupResponse(t *testing.T, w http.ResponseWriter, id int) {
	t.Helper()
	resp := map[string]interface{}{
		"id":                          id,
		"name":                        "test-hostgroup",
		"architecture_id":             1,
		"compute_profile_id":          2,
		"compute_resource_id":         "3",
		"description":                 "desc",
		"domain_id":                   4,
		"medium_id":                   5,
		"operatingsystem_id":          6,
		"parent_id":                   "7",
		"ptable_id":                   8,
		"puppet_ca_proxy_id":          9,
		"puppet_proxy_id":             10,
		"pxe_loader":                  "PXELinux BIOS",
		"realm_id":                    "11",
		"subnet6_id":                  12,
		"subnet_id":                   13,
		"parameters":                  []map[string]string{{"name": "env", "value": "prod"}},
		"group_parameters_attributes": []map[string]string{{"name": "env", "value": "prod"}},
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	require.NoError(t, json.NewEncoder(w).Encode(resp))
}

func newTestHostgroupResource(t *testing.T, srv *httptest.Server) *hostgroupResource {
	t.Helper()
	u, err := url.Parse(srv.URL)
	require.NoError(t, err)
	client := generated.NewClient(*u, generated.ClientCredentials{}, generated.ClientConfig{})
	return &hostgroupResource{client: client}
}

func TestHostgroupResource_Create(t *testing.T) {
	t.Parallel()
	srv := hostgroupTestServer(t)
	defer srv.Close()

	r := newTestHostgroupResource(t, srv)
	plan := hostgroupResourceModel{
		ID:                types.StringNull(),
		ArchitectureID:    types.Int64Value(1),
		ComputeProfileID:  types.Int64Value(2),
		ComputeResourceID: types.Int64Value(3),
		Description:       types.StringValue("desc"),
		DomainID:          types.Int64Value(4),
		MediumID:          types.Int64Value(5),
		OperatingsystemID: types.Int64Value(6),
		ParentID:          types.StringValue("7"),
		PtableID:          types.Int64Value(8),
		PuppetCaProxyID:   types.Int64Value(9),
		PuppetProxyID:     types.Int64Value(10),
		PXELoader:         types.StringValue("PXELinux BIOS"),
		RealmID:           types.StringValue("11"),
		Subnet6ID:         types.Int64Value(12),
		SubnetID:          types.Int64Value(13),
		RootPass:          types.StringValue("secret"),
	}

	var diags diag.Diagnostics
	req := buildHostgroupRequest(plan, &diags)
	require.False(t, diags.HasError(), diags.Errors())

	result, err := r.createHostgroup(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 42, result.ID)
	assert.Equal(t, "test-hostgroup", result.Name)
}

func TestHostgroupResource_Read(t *testing.T) {
	t.Parallel()
	srv := hostgroupTestServer(t)
	defer srv.Close()

	r := newTestHostgroupResource(t, srv)
	result, err := r.readHostgroup(context.Background(), 42)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, 42, result.ID)
	assert.Equal(t, "3", result.ComputeResourceID)
}

func TestHostgroupResource_Update(t *testing.T) {
	t.Parallel()
	srv := hostgroupTestServer(t)
	defer srv.Close()

	r := newTestHostgroupResource(t, srv)
	plan := hostgroupResourceModel{
		ID:                types.StringValue("42"),
		ArchitectureID:    types.Int64Value(1),
		ComputeProfileID:  types.Int64Value(2),
		ComputeResourceID: types.Int64Value(3),
		Description:       types.StringValue("desc"),
		DomainID:          types.Int64Value(4),
		MediumID:          types.Int64Value(5),
		OperatingsystemID: types.Int64Value(6),
		ParentID:          types.StringValue("7"),
		PtableID:          types.Int64Value(8),
		PuppetCaProxyID:   types.Int64Value(9),
		PuppetProxyID:     types.Int64Value(10),
		PXELoader:         types.StringValue("PXELinux BIOS"),
		RealmID:           types.StringValue("11"),
		Subnet6ID:         types.Int64Value(12),
		SubnetID:          types.Int64Value(13),
		RootPass:          types.StringValue("secret"),
	}

	var diags diag.Diagnostics
	req := buildHostgroupRequest(plan, &diags)
	require.False(t, diags.HasError(), diags.Errors())

	result, err := r.updateHostgroup(context.Background(), 42, req)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 42, result.ID)
}

func TestHostgroupResource_Delete(t *testing.T) {
	t.Parallel()
	srv := hostgroupTestServer(t)
	defer srv.Close()

	r := newTestHostgroupResource(t, srv)
	err := r.client.DeleteForemanHostgroup(context.Background(), 42)
	require.NoError(t, err)
}

func TestBuildHostgroupRequest_SurfacesParseErrors(t *testing.T) {
	t.Parallel()
	plan := hostgroupResourceModel{
		ParentID: types.StringValue("not-a-number"),
		RealmID:  types.StringValue("also-not-a-number"),
	}

	var diags diag.Diagnostics
	_ = buildHostgroupRequest(plan, &diags)
	require.True(t, diags.HasError())
	assert.Len(t, diags.Errors(), 2)
}

func TestMarshalHostgroupResultToState_PreservesRootPass(t *testing.T) {
	t.Parallel()
	result := &foremanHostgroupWithParams{
		ForemanHostgroup: generated.ForemanHostgroup{
			ForemanObject: generated.ForemanObject{ID: 1, Name: "hg"},
		},
		Parameters: []byte(`[{"name":"env","value":"prod"}]`),
	}
	state := hostgroupResourceModel{RootPass: types.StringValue("preserve-me")}

	var diags diag.Diagnostics
	marshalHostgroupResultToState(result, &state, &diags)
	require.False(t, diags.HasError(), diags.Errors())

	assert.Equal(t, "preserve-me", state.RootPass.ValueString())
	assert.False(t, state.Parameters.IsNull())
}
