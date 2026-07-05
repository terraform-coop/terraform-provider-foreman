package goforeman

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRawValueString(t *testing.T) {
	t.Parallel()

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
		assert.Equal(t, tc.want, RawValueString(json.RawMessage(tc.raw)), "raw=%s", tc.raw)
	}
}

func TestParseParameters(t *testing.T) {
	t.Parallel()

	t.Run("mixed value types", func(t *testing.T) {
		params, err := ParseParameters(json.RawMessage(`[
			{"name":"a","value":"str"},
			{"name":"b","value":true},
			{"name":"c","value":[1,2]}
		]`))
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"a": "str", "b": "true", "c": "[1,2]"}, params)
	})

	t.Run("null and empty", func(t *testing.T) {
		for _, raw := range []string{"", "null"} {
			params, err := ParseParameters(json.RawMessage(raw))
			require.NoError(t, err)
			assert.Nil(t, params)
		}
	})

	t.Run("malformed", func(t *testing.T) {
		_, err := ParseParameters(json.RawMessage(`{"not":"an array"}`))
		assert.Error(t, err)
	})
}

func TestBuildParameters(t *testing.T) {
	t.Parallel()

	assert.Nil(t, BuildParameters(nil))
	out := BuildParameters(map[string]string{"k": "v"})
	require.Len(t, out, 1)
	assert.Equal(t, map[string]interface{}{"name": "k", "value": "v"}, out[0])
}

func TestAppendDestroyMarkers(t *testing.T) {
	t.Parallel()

	t.Run("removed entry gets a destroy marker", func(t *testing.T) {
		plan := []map[string]interface{}{{"id": int64(1), "name": "eth0"}}
		prior := []map[string]interface{}{
			{"id": int64(1), "name": "eth0"},
			{"id": int64(2), "name": "eth1"},
		}
		out := AppendDestroyMarkers(plan, prior)
		require.Len(t, out, 2)
		assert.Equal(t, map[string]interface{}{"id": int64(2), "_destroy": true}, out[1])
	})

	t.Run("nothing removed, nothing appended", func(t *testing.T) {
		plan := []map[string]interface{}{{"id": int64(1)}}
		out := AppendDestroyMarkers(plan, plan)
		assert.Len(t, out, 1)
	})

	t.Run("entries without server ids are ignored", func(t *testing.T) {
		plan := []map[string]interface{}{{"name": "new-iface"}}
		prior := []map[string]interface{}{{"id": int64(0)}, {"name": "never-created"}}
		out := AppendDestroyMarkers(plan, prior)
		assert.Len(t, out, 1)
	})

	t.Run("float64 ids from generic json decoding", func(t *testing.T) {
		out := AppendDestroyMarkers(nil, []map[string]interface{}{{"id": float64(7)}})
		require.Len(t, out, 1)
		assert.Equal(t, int64(7), out[0]["id"])
	})
}

func TestFirstOutOfOrderIdentifier(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		ids        []string
		wantPrev   string
		wantNext   string
		wantResult bool
	}{
		{"empty", nil, "", "", false},
		{"single", []string{"eth0"}, "", "", false},
		{"already sorted", []string{"eth0", "eth1", "eth2"}, "", "", false},
		{"already sorted, non-numeric suffixes", []string{"ens160", "ens192"}, "", "", false},
		{"out of order", []string{"eth1", "eth0"}, "eth1", "eth0", true},
		{"out of order later in list", []string{"eth0", "eth2", "eth1"}, "eth2", "eth1", true},
		{"duplicate identifiers", []string{"eth0", "eth0"}, "", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			prev, next, found := FirstOutOfOrderIdentifier(tc.ids)
			assert.Equal(t, tc.wantResult, found)
			assert.Equal(t, tc.wantPrev, prev)
			assert.Equal(t, tc.wantNext, next)
		})
	}
}
