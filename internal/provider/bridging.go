package provider

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ---------------------------------------------------------------------------
// Shared bridging helpers — used by host, operating_system, and hostgroup
// resources to convert between Terraform types and Foreman API JSON.
// ---------------------------------------------------------------------------

// Parameters bridging — types.Map ↔ [{name,value}] for API.
// Used by: host (host_parameters_attributes), operating_system (os_parameters_attributes),
// hostgroup (group_parameters_attributes).

// flattenParameters converts types.Map → [{name,value}] for API request body.
func flattenParameters(m types.Map) []map[string]interface{} {
	if m.IsNull() || m.IsUnknown() {
		return nil
	}
	elements := m.Elements()
	result := make([]map[string]interface{}, 0, len(elements))
	for k, v := range elements {
		result = append(result, map[string]interface{}{
			"name":  k,
			"value": v.(types.String).ValueString(),
		})
	}
	return result
}

// expandParameters converts API response [{name,value}] → types.Map.
func expandParameters(raw json.RawMessage) types.Map {
	if len(raw) == 0 || string(raw) == "null" {
		return types.MapNull(types.StringType)
	}
	var params []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}
	if err := json.Unmarshal(raw, &params); err != nil {
		return types.MapNull(types.StringType)
	}
	elements := make(map[string]attr.Value, len(params))
	for _, p := range params {
		elements[p.Name] = types.StringValue(p.Value)
	}
	m, diags := types.MapValue(types.StringType, elements)
	if diags.HasError() {
		return types.MapNull(types.StringType)
	}
	return m
}

// List bridging — types.List of int64 → []int64 for API.
// Used by: operating_system (architecture_ids, medium_ids, provisioning_template_ids, ptable_ids).

// listToInt64Slice converts a types.List with Int64 elements to a []int64.
func listToInt64Slice(list types.List) []int64 {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}
	elems := list.Elements()
	result := make([]int64, 0, len(elems))
	for _, v := range elems {
		if iv, ok := v.(types.Int64); ok {
			result = append(result, iv.ValueInt64())
		}
	}
	return result
}

// int64SliceToList converts a []int64 to a types.List of Int64 elements.
// Diagnostics are appended on error instead of being discarded.
func int64SliceToList(ids []int64, diags *diag.Diagnostics) types.List {
	if ids == nil {
		return types.ListNull(types.Int64Type)
	}
	elems := make([]attr.Value, len(ids))
	for i, v := range ids {
		elems[i] = types.Int64Value(v)
	}
	list, d := types.ListValue(types.Int64Type, elems)
	diags.Append(d...)
	return list
}

// Compute attributes bridging — types.String (JSON) ↔ map for API.
// Used by: host (compute_attributes).

// flattenComputeAttributes converts types.String (JSON) → map for API request body.
func flattenComputeAttributes(s types.String) map[string]interface{} {
	if s.IsNull() || s.IsUnknown() || s.ValueString() == "" {
		return nil
	}
	var attrs map[string]interface{}
	if err := json.Unmarshal([]byte(s.ValueString()), &attrs); err != nil {
		return nil
	}
	return attrs
}

// expandComputeAttributes converts API response → types.String (JSON).
func expandComputeAttributes(raw json.RawMessage) types.String {
	if len(raw) == 0 || string(raw) == "null" {
		return types.StringNull()
	}
	return types.StringValue(string(raw))
}

// parseStringToInt64 parses a string representation of an int64 and appends a
// diagnostic on error. An empty string is treated as 0 (unset) so that optional
// string-as-int64 fields do not produce false positives.
func parseStringToInt64(s, field string, diags *diag.Diagnostics) int64 {
	if s == "" {
		return 0
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		diags.AddAttributeError(
			path.Root(field),
			"Invalid Field",
			fmt.Sprintf("Unable to parse %s: %s", field, err),
		)
		return 0
	}
	return v
}

// Primitive helpers for safe type conversions from API response.

func int64Value(v interface{}) types.Int64 {
	switch val := v.(type) {
	case float64:
		return types.Int64Value(int64(val))
	case int:
		return types.Int64Value(int64(val))
	case int64:
		return types.Int64Value(val)
	default:
		return types.Int64Null()
	}
}

func stringValue(v interface{}) types.String {
	if v == nil {
		return types.StringNull()
	}
	if s, ok := v.(string); ok {
		return types.StringValue(s)
	}
	return types.StringNull()
}

func boolValue(v interface{}) types.Bool {
	if v == nil {
		return types.BoolNull()
	}
	if b, ok := v.(bool); ok {
		return types.BoolValue(b)
	}
	return types.BoolNull()
}
