package provider

import (
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/attr"
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
//
// Foreman parameters are user-typeable (parameter_type: string, boolean,
// integer, real, array, hash, yaml, or json), so "value" is not always a
// JSON string - decoding it into a plain Go string unconditionally makes
// json.Unmarshal fail for the whole parameter list the moment any single
// parameter has a non-string type, which previously surfaced as e.g. host
// or hostgroup parameters silently disappearing (or erroring) on import/read
// (see issues #129, #136). Decode "value" as raw JSON per-element instead,
// and render it as text for storage in this provider's Map<String>
// representation.
func expandParameters(raw json.RawMessage) types.Map {
	if len(raw) == 0 || string(raw) == "null" {
		return types.MapNull(types.StringType)
	}
	var params []struct {
		Name  string          `json:"name"`
		Value json.RawMessage `json:"value"`
	}
	if err := json.Unmarshal(raw, &params); err != nil {
		return types.MapNull(types.StringType)
	}
	elements := make(map[string]attr.Value, len(params))
	for _, p := range params {
		elements[p.Name] = types.StringValue(parameterValueToString(p.Value))
	}
	m, diags := types.MapValue(types.StringType, elements)
	if diags.HasError() {
		return types.MapNull(types.StringType)
	}
	return m
}

// parameterValueToString renders a Foreman parameter's raw JSON "value" as
// plain text: a JSON string decodes to its unquoted contents (the common
// case), while a boolean/number/array/object's JSON text is used as-is
// (e.g. "true", "123", ["a","b"]) so the value round-trips losslessly
// through this provider's string representation instead of crashing.
func parameterValueToString(raw json.RawMessage) string {
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	return string(raw)
}

// maybeStringList extracts a []string from a types.List attr.Value, for
// string-list fields nested inside an object (e.g. an interface's
// attached_devices).
func maybeStringList(v attr.Value) []string {
	l, ok := v.(types.List)
	if !ok || l.IsNull() || l.IsUnknown() {
		return nil
	}
	elements := l.Elements()
	out := make([]string, 0, len(elements))
	for _, e := range elements {
		if s, ok := e.(types.String); ok {
			out = append(out, s.ValueString())
		}
	}
	return out
}

// stringListValue builds a types.List of strings from a raw decoded JSON
// array, for string-list fields nested inside an object.
func stringListValue(v interface{}) types.List {
	arr, ok := v.([]interface{})
	if !ok {
		return types.ListNull(types.StringType)
	}
	elems := make([]attr.Value, 0, len(arr))
	for _, item := range arr {
		if s, ok := item.(string); ok {
			elems = append(elems, types.StringValue(s))
		}
	}
	list, diags := types.ListValue(types.StringType, elems)
	if diags.HasError() {
		return types.ListNull(types.StringType)
	}
	return list
}

// maybeInt64 extracts int64 from attr.Value, returns 0 if null.
func maybeInt64(v attr.Value) int64 {
	if iv, ok := v.(types.Int64); ok && !iv.IsNull() {
		return iv.ValueInt64()
	}
	return 0
}

// maybeBool extracts bool from attr.Value, returns false if null.
func maybeBool(v attr.Value) bool {
	if bv, ok := v.(types.Bool); ok && !bv.IsNull() {
		return bv.ValueBool()
	}
	return false
}

// boolPointerOrNil converts a types.Bool to *bool: nil when null/unknown,
// otherwise a pointer to its value. Used for optional bool request fields
// tagged without "omitempty" (see isOptionalBoolPointerField in the
// generator) so an explicit "false" can actually reach the API instead of
// being silently dropped by Go's zero-value/omitempty interaction.
func boolPointerOrNil(v types.Bool) *bool {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	b := v.ValueBool()
	return &b
}

// maybeString extracts string from attr.Value, returns "" if null.
func maybeString(v attr.Value) string {
	if sv, ok := v.(types.String); ok && !sv.IsNull() {
		return sv.ValueString()
	}
	return ""
}

// maybeJSON extracts a JSON string from attr.Value, unmarshals to map.
func maybeJSON(v attr.Value) map[string]interface{} {
	sv, ok := v.(types.String)
	if !ok || sv.IsNull() {
		return nil
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(sv.ValueString()), &m); err != nil {
		return nil
	}
	return m
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

// mapToJSONString re-serializes a decoded JSON object (e.g. a nested
// interface's compute_attributes) back into a JSON string attr.Value, for
// free-form hash sub-fields exposed as opaque JSON strings.
func mapToJSONString(v interface{}) types.String {
	m, ok := v.(map[string]interface{})
	if !ok || m == nil {
		return types.StringNull()
	}
	b, err := json.Marshal(m)
	if err != nil {
		return types.StringNull()
	}
	return types.StringValue(string(b))
}
