package goforeman

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestKatelloLifecycleEnvironment_UnmarshalFlattensPriorSuccessor covers the
// read/write asymmetry: writes take a flat prior_id, reads return nested
// prior/successor objects, which UnmarshalJSON flattens back to plain IDs.
func TestKatelloLifecycleEnvironment_UnmarshalFlattensPriorSuccessor(t *testing.T) {
	t.Parallel()

	var env KatelloLifecycleEnvironment
	require.NoError(t, json.Unmarshal([]byte(`{
		"id": 3, "name": "dev", "library": false,
		"prior": {"id": 1, "name": "Library"},
		"successor": {"id": 5, "name": "prod"}
	}`), &env))
	assert.Equal(t, 1, env.PriorID)
	assert.Equal(t, 5, env.SuccessorID)
	assert.Equal(t, "dev", env.Name)

	// missing prior/successor decode to 0, not an error
	var bare KatelloLifecycleEnvironment
	require.NoError(t, json.Unmarshal([]byte(`{"id": 1, "name": "Library", "library": true}`), &bare))
	assert.Zero(t, bare.PriorID)
	assert.Zero(t, bare.SuccessorID)
}
