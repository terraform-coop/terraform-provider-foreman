package provider

import (
	"encoding/json"

	"github.com/terraform-coop/terraform-provider-foreman/goforeman"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ---------------------------------------------------------------------------
// Shared bridging helpers — thin converters between Terraform framework
// types and plain Go values. All Foreman API convention/quirk knowledge
// (parameter collections, polymorphic values, _destroy semantics, ...)
// lives in the goforeman package; these only translate types.
// ---------------------------------------------------------------------------

// flattenParameters converts a types.Map into the [{name,value}] shape
// Foreman expects (see goforeman.BuildParameters) for request bodies.
func flattenParameters(m types.Map) []map[string]interface{} {
	if m.IsNull() || m.IsUnknown() {
		return nil
	}
	elements := m.Elements()
	params := make(map[string]string, len(elements))
	for k, v := range elements {
		params[k] = v.(types.String).ValueString()
	}
	return goforeman.BuildParameters(params)
}

// expandParameters converts a Foreman parameter collection from an API
// response (see goforeman.ParseParameters) into a types.Map.
func expandParameters(raw json.RawMessage) types.Map {
	params, err := goforeman.ParseParameters(raw)
	if err != nil || params == nil {
		return types.MapNull(types.StringType)
	}
	elements := make(map[string]attr.Value, len(params))
	for name, value := range params {
		elements[name] = types.StringValue(value)
	}
	m, diags := types.MapValue(types.StringType, elements)
	if diags.HasError() {
		return types.MapNull(types.StringType)
	}
	return m
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

// intersectIDSet returns the members of tracked that the server still
// reports. Association ID sets track only the IDs the user configured:
// Foreman auto-associates additional members server-side (see the
// generated resources' assoc-field handling), so Read must drop tracked
// IDs the server lost (real drift) without adopting foreign members the
// server added on its own. A null/unknown tracked set stays null.
func intersectIDSet(tracked types.Set, serverIDs []int64) types.Set {
	if tracked.IsNull() || tracked.IsUnknown() {
		return types.SetNull(types.Int64Type)
	}
	onServer := make(map[int64]bool, len(serverIDs))
	for _, id := range serverIDs {
		onServer[id] = true
	}
	kept := []attr.Value{}
	for _, v := range tracked.Elements() {
		if iv, ok := v.(types.Int64); ok && onServer[iv.ValueInt64()] {
			kept = append(kept, iv)
		}
	}
	out, diags := types.SetValue(types.Int64Type, kept)
	if diags.HasError() {
		return types.SetNull(types.Int64Type)
	}
	return out
}

// attrIsSet reports whether a nested object's attribute carries a real,
// user-set value (not null/unknown) and should be sent on the wire.
func attrIsSet(v attr.Value) bool {
	return v != nil && !v.IsNull() && !v.IsUnknown()
}
