package api

import (
	"math/rand"
	"strconv"
	"testing"
)

// ----------------------------------------------------------------------------
// intIdToJSONString
// ----------------------------------------------------------------------------

// Ensure a negative ID returns nil
func TestIntIdToJSONString_NegativeIdReturnNil(t *testing.T) {
	// NOTE(ALL): rand.Int() returns positive value - flip to a negative
	randInt := rand.Int() * -1
	output := intIdToJSONString(randInt)
	if output != nil {
		t.Fatalf(
			"intIdToJSONString did not return correct value. "+
				"Expected [nil], got [%v] for input [%d]",
			output,
			randInt,
		)
	}
}

// Ensure the ID 0 returns nil
func TestIntIdToJSONString_ZeroIdReturnNil(t *testing.T) {
	output := intIdToJSONString(0)
	if output != nil {
		t.Fatalf(
			"intIdToJSONString did not return correct value. "+
				"Expected [nil], got [%v] for input 0",
			output,
		)
	}
}

// Ensure a positive ID returns non-nil value
func TestIntIdToJSONString_PositiveReturnNotNil(t *testing.T) {
	// NOTE(ALL): rand.Int() returns positive value
	randInt := rand.Int()
	output := intIdToJSONString(randInt)
	if output == nil {
		t.Fatalf(
			"intIdToJSONString did not return correct value. "+
				"Expected non-nil return, got [nil] for input [%d]",
			randInt,
		)
	}
}

// Ensure a positive ID returns a string
func TestIntIdToJSONString_PositiveReturnString(t *testing.T) {
	var ok bool
	// NOTE(ALL): rand.Int() returns positive value
	randInt := rand.Int()
	output := intIdToJSONString(randInt)
	if _, ok = output.(string); !ok {
		t.Fatalf(
			"intIdToJSONString did not return correct value. "+
				"Expected return type to be [string], got [%T] "+
				"for input [%d]",
			output,
			randInt,
		)
	}
}

// Ensure a positive ID returns the string representation of the ID
func TestIntIdToJSONString_PositiveReturnStringValue(t *testing.T) {
	// NOTE(ALL): rand.Int() returns positive value
	randInt := rand.Int()
	output := intIdToJSONString(randInt)
	expectedOutput := strconv.Itoa(randInt)
	if output.(string) != expectedOutput {
		t.Fatalf(
			"intIdToJSONString did not return correct value. "+
				"Expected [%s], got [%v] "+
				"for input [%d]",
			expectedOutput,
			output,
			randInt,
		)
	}
}

// ----------------------------------------------------------------------------
// foremanObjectArrayToIdIntArray
// ----------------------------------------------------------------------------

// Ensures the input and output arrays have the same length
func TestForemanObjectArrayToIdIntArray_SameLength(t *testing.T) {
	// NOTE(ALL): rand.Int() returns positive value
	randInt := rand.Int() % 100
	output := len(foremanObjectArrayToIdIntArray(make([]ForemanObject, randInt)))
	if output != randInt {
		t.Fatalf(
			"foremanObjectArrayToIdIntArray did not return an array with the correct "+
				"length. Expected [%d], got [%d] for value [%d].",
			randInt,
			output,
			randInt,
		)
	}
	output = len(foremanObjectArrayToIdIntArray(make([]ForemanObject, 0)))
	if output != 0 {
		t.Fatalf(
			"foremanObjectArrayToIdIntArray did not return an array with the correct "+
				"length. Expected [0], got [%d] for value [0].",
			output,
		)
	}
}

// Ensures the returned array has the correct ID values in the right index
func TestForemanObjectArrayToIdIntArray_Value(t *testing.T) {

	// NOTE(ALL): rand.Int() returns positive value
	randInt := rand.Int() % 100
	input := make([]ForemanObject, randInt)
	for j := 0; j < randInt; j++ {
		input[j].Id = rand.Int()
	}

	output := foremanObjectArrayToIdIntArray(input)
	for k := 0; k < randInt; k++ {
		if output[k] != input[k].Id {
			t.Fatalf(
				"foremanObjectArrayToIdIntArray did not return correct value. "+
					"Expected [%d], got [%d] for value [%+v] at index [%d]",
				input[k].Id,
				output[k],
				input[k],
				k,
			)
		}
	}

}
