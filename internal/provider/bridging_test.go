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

	t.Run("mixed value types", func(t *testing.T) {
		// Foreman parameters are user-typeable (parameter_type: boolean,
		// integer, array, hash, ...) - "value" is not always a JSON string.
		// See issues #129, #136.
		raw := json.RawMessage(`[
			{"name":"str","value":"hello"},
			{"name":"bool","value":true},
			{"name":"num","value":42},
			{"name":"arr","value":["a","b"]},
			{"name":"obj","value":{"x":1}}
		]`)
		m := expandParameters(raw)
		if m.IsNull() {
			t.Fatal("expected non-null map")
		}
		elems := m.Elements()
		cases := map[string]string{
			"str":  "hello",
			"bool": "true",
			"num":  "42",
			"arr":  `["a","b"]`,
			"obj":  `{"x":1}`,
		}
		for k, want := range cases {
			got := elems[k].(types.String).ValueString()
			if got != want {
				t.Fatalf("%s: got %q, want %q", k, got, want)
			}
		}
	})
}

func TestParameterValueToString(t *testing.T) {
	cases := []struct {
		raw  string
		want string
	}{
		{`"hello"`, "hello"},
		{`true`, "true"},
		{`42`, "42"},
		{`null`, ""},
		{`["a","b"]`, `["a","b"]`},
	}
	for _, tc := range cases {
		got := parameterValueToString(json.RawMessage(tc.raw))
		if got != tc.want {
			t.Errorf("parameterValueToString(%s) = %q, want %q", tc.raw, got, tc.want)
		}
	}
}

// ---------------------------------------------------------------------------
// maybeInt64 / maybeBool / maybeString / maybeJSON
// ---------------------------------------------------------------------------

func TestMaybeInt64(t *testing.T) {
	if v := maybeInt64(types.Int64Value(42)); v != 42 {
		t.Fatalf("expected 42, got %d", v)
	}
	if v := maybeInt64(types.Int64Null()); v != 0 {
		t.Fatalf("expected 0 for null, got %d", v)
	}
	if v := maybeInt64(types.StringValue("nope")); v != 0 {
		t.Fatalf("expected 0 for wrong type, got %d", v)
	}
}

func TestMaybeBool(t *testing.T) {
	if v := maybeBool(types.BoolValue(true)); !v {
		t.Fatal("expected true")
	}
	if v := maybeBool(types.BoolNull()); v {
		t.Fatal("expected false for null")
	}
	if v := maybeBool(types.StringValue("nope")); v {
		t.Fatal("expected false for wrong type")
	}
}

func TestMaybeString(t *testing.T) {
	if v := maybeString(types.StringValue("hi")); v != "hi" {
		t.Fatalf("expected hi, got %s", v)
	}
	if v := maybeString(types.StringNull()); v != "" {
		t.Fatalf("expected empty for null, got %s", v)
	}
	if v := maybeString(types.BoolValue(true)); v != "" {
		t.Fatalf("expected empty for wrong type, got %s", v)
	}
}

func TestMaybeJSON(t *testing.T) {
	m := maybeJSON(types.StringValue(`{"cpus":2}`))
	if m == nil || m["cpus"] != float64(2) {
		t.Fatalf("unexpected result: %v", m)
	}
	if m := maybeJSON(types.StringNull()); m != nil {
		t.Fatalf("expected nil for null, got %v", m)
	}
	if m := maybeJSON(types.StringValue("not-json")); m != nil {
		t.Fatalf("expected nil for invalid json, got %v", m)
	}
}

// ---------------------------------------------------------------------------
// maybeStringList / stringListValue
// ---------------------------------------------------------------------------

func TestMaybeStringList(t *testing.T) {
	l := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("a"), types.StringValue("b")})
	result := maybeStringList(l)
	if len(result) != 2 || result[0] != "a" || result[1] != "b" {
		t.Fatalf("unexpected result: %v", result)
	}
	if result := maybeStringList(types.ListNull(types.StringType)); result != nil {
		t.Fatalf("expected nil for null list, got %v", result)
	}
	if result := maybeStringList(types.StringValue("nope")); result != nil {
		t.Fatalf("expected nil for wrong type, got %v", result)
	}
}

func TestStringListValue(t *testing.T) {
	l := stringListValue([]interface{}{"a", "b"})
	if l.IsNull() || len(l.Elements()) != 2 {
		t.Fatalf("unexpected result: %v", l)
	}
	if l := stringListValue(nil); !l.IsNull() {
		t.Fatal("expected null list for non-array input")
	}
}

// ---------------------------------------------------------------------------
// mapToJSONString
// ---------------------------------------------------------------------------

func TestMapToJSONString(t *testing.T) {
	s := mapToJSONString(map[string]interface{}{"type": "bridge"})
	if s.IsNull() || s.ValueString() != `{"type":"bridge"}` {
		t.Fatalf("unexpected result: %v", s)
	}
	if s := mapToJSONString(nil); !s.IsNull() {
		t.Fatal("expected null for nil input")
	}
	if s := mapToJSONString("not-a-map"); !s.IsNull() {
		t.Fatal("expected null for wrong type")
	}
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

// ---------------------------------------------------------------------------
// boolPointerOrNil
// ---------------------------------------------------------------------------

func TestBoolPointerOrNil(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		p := boolPointerOrNil(types.BoolValue(true))
		if p == nil || !*p {
			t.Fatalf("expected pointer to true, got %v", p)
		}
	})

	t.Run("false", func(t *testing.T) {
		p := boolPointerOrNil(types.BoolValue(false))
		if p == nil || *p {
			t.Fatalf("expected pointer to false, got %v", p)
		}
	})

	t.Run("null", func(t *testing.T) {
		p := boolPointerOrNil(types.BoolNull())
		if p != nil {
			t.Fatalf("expected nil pointer for null, got %v", p)
		}
	})

	t.Run("unknown", func(t *testing.T) {
		p := boolPointerOrNil(types.BoolUnknown())
		if p != nil {
			t.Fatalf("expected nil pointer for unknown, got %v", p)
		}
	})
}
