package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestParameterParent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		model    parameterResourceModel
		wantType string
		wantID   int64
		wantErr  bool
	}{
		{
			name:     "host_id set",
			model:    parameterResourceModel{HostID: types.Int64Value(5)},
			wantType: "hosts",
			wantID:   5,
		},
		{
			name:     "organization_id set",
			model:    parameterResourceModel{OrganizationID: types.Int64Value(2)},
			wantType: "organizations",
			wantID:   2,
		},
		{
			name:    "none set",
			model:   parameterResourceModel{},
			wantErr: true,
		},
		{
			name: "two set",
			model: parameterResourceModel{
				HostID:      types.Int64Value(5),
				HostgroupID: types.Int64Value(6),
			},
			wantErr: true,
		},
		{
			name:    "unknown value not treated as set",
			model:   parameterResourceModel{HostID: types.Int64Unknown()},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			parentType, parentID, diags := parameterParent(&tc.model)
			if tc.wantErr {
				assert.True(t, diags.HasError())
				return
			}
			assert.False(t, diags.HasError())
			assert.Equal(t, tc.wantType, parentType)
			assert.Equal(t, tc.wantID, parentID)
		})
	}
}

func TestParseParameterImportID(t *testing.T) {
	t.Parallel()

	t.Run("valid", func(t *testing.T) {
		field, id, param, err := parseParameterImportID("host_id:5:42")
		assert.NoError(t, err)
		assert.Equal(t, "host_id", field)
		assert.Equal(t, int64(5), id)
		assert.Equal(t, "42", param)
	})

	t.Run("wrong number of parts", func(t *testing.T) {
		_, _, _, err := parseParameterImportID("42")
		assert.Error(t, err)
	})

	t.Run("unknown parent field", func(t *testing.T) {
		_, _, _, err := parseParameterImportID("bogus_id:5:42")
		assert.Error(t, err)
	})

	t.Run("non-numeric parent id", func(t *testing.T) {
		_, _, _, err := parseParameterImportID("host_id:notanumber:42")
		assert.Error(t, err)
	})
}
