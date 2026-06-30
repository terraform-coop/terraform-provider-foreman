package provider

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ---------------------------------------------------------------------------
// flattenParameters
// ---------------------------------------------------------------------------

func TestFlattenParameters(t *testing.T) {
	t.Run("non-empty map", func(t *testing.T) {
		m := types.MapValueMust(types.StringType, map[string]attr.Value{
			"foo": types.StringValue("bar"),
			"baz": types.StringValue("qux"),
		})
		result := flattenParameters(m)
		if len(result) != 2 {
			t.Fatalf("expected 2 entries, got %d", len(result))
		}
		found := map[string]string{}
		for _, r := range result {
			found[r["name"].(string)] = r["value"].(string)
		}
		if found["foo"] != "bar" || found["baz"] != "qux" {
			t.Fatalf("unexpected values: %v", found)
		}
	})

	t.Run("empty map", func(t *testing.T) {
		m := types.MapValueMust(types.StringType, map[string]attr.Value{})
		result := flattenParameters(m)
		if len(result) != 0 {
			t.Fatalf("expected 0 entries, got %d", len(result))
		}
	})

	t.Run("null map", func(t *testing.T) {
		m := types.MapNull(types.StringType)
		result := flattenParameters(m)
		if result != nil {
			t.Fatalf("expected nil, got %v", result)
		}
	})
}

// ---------------------------------------------------------------------------
// expandParameters
// ---------------------------------------------------------------------------

func TestExpandParameters(t *testing.T) {
	t.Run("valid json array", func(t *testing.T) {
		raw := json.RawMessage(`[{"name":"k1","value":"v1"},{"name":"k2","value":"v2"}]`)
		m := expandParameters(raw)
		if m.IsNull() {
			t.Fatal("expected non-null map")
		}
		elems := m.Elements()
		if elems["k1"].(types.String).ValueString() != "v1" {
			t.Fatalf("k1 mismatch")
		}
		if elems["k2"].(types.String).ValueString() != "v2" {
			t.Fatalf("k2 mismatch")
		}
	})

	t.Run("empty array", func(t *testing.T) {
		raw := json.RawMessage(`[]`)
		m := expandParameters(raw)
		if m.IsNull() {
			t.Fatal("expected non-null map for empty array")
		}
		if len(m.Elements()) != 0 {
			t.Fatalf("expected 0 elements, got %d", len(m.Elements()))
		}
	})

	t.Run("null", func(t *testing.T) {
		raw := json.RawMessage(`null`)
		m := expandParameters(raw)
		if !m.IsNull() {
			t.Fatal("expected null map")
		}
	})

	t.Run("empty bytes", func(t *testing.T) {
		m := expandParameters(json.RawMessage{})
		if !m.IsNull() {
			t.Fatal("expected null map for empty bytes")
		}
	})
}

// ---------------------------------------------------------------------------
// flattenComputeAttributes
// ---------------------------------------------------------------------------

func TestFlattenComputeAttributes(t *testing.T) {
	t.Run("valid json string", func(t *testing.T) {
		s := types.StringValue(`{"cpus":2,"memory":1024}`)
		result := flattenComputeAttributes(s)
		if result == nil {
			t.Fatal("expected non-nil map")
		}
		if result["cpus"] != float64(2) {
			t.Fatalf("cpus mismatch: %v", result["cpus"])
		}
		if result["memory"] != float64(1024) {
			t.Fatalf("memory mismatch: %v", result["memory"])
		}
	})

	t.Run("empty string", func(t *testing.T) {
		s := types.StringValue("")
		result := flattenComputeAttributes(s)
		if result != nil {
			t.Fatalf("expected nil, got %v", result)
		}
	})

	t.Run("null string", func(t *testing.T) {
		s := types.StringNull()
		result := flattenComputeAttributes(s)
		if result != nil {
			t.Fatalf("expected nil, got %v", result)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		s := types.StringValue("not-json")
		result := flattenComputeAttributes(s)
		if result != nil {
			t.Fatalf("expected nil for invalid json, got %v", result)
		}
	})
}

// ---------------------------------------------------------------------------
// expandComputeAttributes
// ---------------------------------------------------------------------------

func TestExpandComputeAttributes(t *testing.T) {
	t.Run("valid json", func(t *testing.T) {
		raw := json.RawMessage(`{"cpus":4}`)
		s := expandComputeAttributes(raw)
		if s.IsNull() {
			t.Fatal("expected non-null string")
		}
		if s.ValueString() != `{"cpus":4}` {
			t.Fatalf("unexpected value: %s", s.ValueString())
		}
	})

	t.Run("empty bytes", func(t *testing.T) {
		s := expandComputeAttributes(json.RawMessage{})
		if !s.IsNull() {
			t.Fatal("expected null string for empty bytes")
		}
	})

	t.Run("null", func(t *testing.T) {
		s := expandComputeAttributes(json.RawMessage(`null`))
		if !s.IsNull() {
			t.Fatal("expected null string for null input")
		}
	})
}

// ---------------------------------------------------------------------------
// flattenInterfacesAttributes
// ---------------------------------------------------------------------------

func TestFlattenInterfacesAttributes(t *testing.T) {
	t.Run("list with objects", func(t *testing.T) {
		obj := types.ObjectValueMust(interfaceAttrTypes, map[string]attr.Value{
			"id":                 types.Int64Value(1),
			"primary":            types.BoolValue(true),
			"ip":                 types.StringValue("10.0.0.1"),
			"mac":                types.StringValue("aa:bb:cc:dd:ee:ff"),
			"name":               types.StringValue("eth0"),
			"subnet_id":          types.Int64Value(5),
			"identifier":         types.StringValue("eth0"),
			"managed":            types.BoolValue(true),
			"provision":          types.BoolValue(true),
			"virtual":            types.BoolValue(false),
			"type":               types.StringValue("interface"),
			"bmc_provider":       types.StringValue("ipmitool"),
			"username":           types.StringValue("admin"),
			"password":           types.StringValue("secret"),
			"domain_id":          types.Int64Value(3),
			"attached_to":        types.StringValue(""),
			"attached_devices":   types.StringValue(""),
			"compute_attributes": types.StringValue(`{"type":"bridge"}`),
		})
		l := types.ListValueMust(types.ObjectType{AttrTypes: interfaceAttrTypes}, []attr.Value{obj})
		result := flattenInterfacesAttributes(l)
		if len(result) != 1 {
			t.Fatalf("expected 1 entry, got %d", len(result))
		}
		m := result[0]
		if m["ip"] != "10.0.0.1" {
			t.Fatalf("ip mismatch: %v", m["ip"])
		}
		if m["mac"] != "aa:bb:cc:dd:ee:ff" {
			t.Fatalf("mac mismatch: %v", m["mac"])
		}
		if m["provider"] != "ipmitool" {
			t.Fatalf("provider mismatch: %v", m["provider"])
		}
		ca, ok := m["compute_attributes"].(map[string]interface{})
		if !ok || ca["type"] != "bridge" {
			t.Fatalf("compute_attributes mismatch: %v", m["compute_attributes"])
		}
	})

	t.Run("empty list", func(t *testing.T) {
		l := types.ListValueMust(types.ObjectType{AttrTypes: interfaceAttrTypes}, []attr.Value{})
		result := flattenInterfacesAttributes(l)
		if result != nil {
			t.Fatalf("expected nil for empty list, got %v", result)
		}
	})

	t.Run("null list", func(t *testing.T) {
		l := types.ListNull(types.ObjectType{AttrTypes: interfaceAttrTypes})
		result := flattenInterfacesAttributes(l)
		if result != nil {
			t.Fatalf("expected nil for null list, got %v", result)
		}
	})
}

// ---------------------------------------------------------------------------
// expandInterfacesAttributes
// ---------------------------------------------------------------------------

func TestExpandInterfacesAttributes(t *testing.T) {
	t.Run("valid json array", func(t *testing.T) {
		raw := json.RawMessage(`[{"id":1,"primary":true,"ip":"10.0.0.1","mac":"aa:bb:cc:dd:ee:ff","name":"eth0","subnet_id":5,"identifier":"eth0","managed":true,"provision":true,"virtual":false,"type":"interface","bmc_provider":"ipmitool","username":"admin","password":"secret","domain_id":3,"attached_to":"","attached_devices":"","compute_attributes":"{\"type\":\"bridge\"}"}]`)
		l := expandInterfacesAttributes(raw)
		if l.IsNull() {
			t.Fatal("expected non-null list")
		}
		if len(l.Elements()) != 1 {
			t.Fatalf("expected 1 element, got %d", len(l.Elements()))
		}
		obj := l.Elements()[0].(types.Object)
		attrs := obj.Attributes()
		if attrs["ip"].(types.String).ValueString() != "10.0.0.1" {
			t.Fatalf("ip mismatch")
		}
		if attrs["bmc_provider"].(types.String).ValueString() != "ipmitool" {
			t.Fatalf("bmc_provider mismatch")
		}
	})

	t.Run("empty array", func(t *testing.T) {
		raw := json.RawMessage(`[]`)
		l := expandInterfacesAttributes(raw)
		if l.IsNull() {
			t.Fatal("expected non-null list for empty array")
		}
		if len(l.Elements()) != 0 {
			t.Fatalf("expected 0 elements, got %d", len(l.Elements()))
		}
	})

	t.Run("null", func(t *testing.T) {
		l := expandInterfacesAttributes(json.RawMessage(`null`))
		if !l.IsNull() {
			t.Fatal("expected null list")
		}
	})

	t.Run("empty bytes", func(t *testing.T) {
		l := expandInterfacesAttributes(json.RawMessage{})
		if !l.IsNull() {
			t.Fatal("expected null list for empty bytes")
		}
	})
}

// ---------------------------------------------------------------------------
// int64Value
// ---------------------------------------------------------------------------

func TestInt64Value(t *testing.T) {
	t.Run("float64", func(t *testing.T) {
		v := int64Value(float64(42))
		if v.ValueInt64() != 42 {
			t.Fatalf("expected 42, got %d", v.ValueInt64())
		}
	})

	t.Run("int", func(t *testing.T) {
		v := int64Value(int(7))
		if v.ValueInt64() != 7 {
			t.Fatalf("expected 7, got %d", v.ValueInt64())
		}
	})

	t.Run("int64", func(t *testing.T) {
		v := int64Value(int64(99))
		if v.ValueInt64() != 99 {
			t.Fatalf("expected 99, got %d", v.ValueInt64())
		}
	})

	t.Run("nil", func(t *testing.T) {
		v := int64Value(nil)
		if !v.IsNull() {
			t.Fatal("expected null int64")
		}
	})

	t.Run("unsupported type", func(t *testing.T) {
		v := int64Value("not a number")
		if !v.IsNull() {
			t.Fatal("expected null int64 for unsupported type")
		}
	})
}

// ---------------------------------------------------------------------------
// stringValue
// ---------------------------------------------------------------------------

func TestStringValue(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		v := stringValue("hello")
		if v.ValueString() != "hello" {
			t.Fatalf("expected hello, got %s", v.ValueString())
		}
	})

	t.Run("nil", func(t *testing.T) {
		v := stringValue(nil)
		if !v.IsNull() {
			t.Fatal("expected null string")
		}
	})

	t.Run("non-string", func(t *testing.T) {
		v := stringValue(123)
		if !v.IsNull() {
			t.Fatal("expected null string for non-string input")
		}
	})
}

// ---------------------------------------------------------------------------
// boolValue
// ---------------------------------------------------------------------------

func TestBoolValue(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		v := boolValue(true)
		if !v.ValueBool() {
			t.Fatal("expected true")
		}
	})

	t.Run("false", func(t *testing.T) {
		v := boolValue(false)
		if v.ValueBool() {
			t.Fatal("expected false")
		}
	})

	t.Run("nil", func(t *testing.T) {
		v := boolValue(nil)
		if !v.IsNull() {
			t.Fatal("expected null bool")
		}
	})

	t.Run("non-bool", func(t *testing.T) {
		v := boolValue("yes")
		if !v.IsNull() {
			t.Fatal("expected null bool for non-bool input")
		}
	})
}
