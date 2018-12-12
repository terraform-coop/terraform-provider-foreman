package api

import (
	"strconv"
)

// ----------------------------------------------------------------------------
// Foreman Object Model
// ----------------------------------------------------------------------------

// Base Foreman API object in the Foreman object model.  Every API entity
// has the following attributes:
type ForemanObject struct {
	// Unique identifier for this object
	Id int `json:"id"`
	// Human readable name of the API object
	Name string `json:"name"`
	// Timestamp of when the API object was created in the following format:
	// "%Y-%m-%d %H-%M-%S UTC"
	CreatedAt string `json:"created_at"`
	// Timestamp of when the API object was last updated in the following format:
	// "%Y-%m-%d %H-%M-%S UTC"
	UpdatedAt string `json:"updated_at"`
}

// ----------------------------------------------------------------------------
// Foreman API Helper Functions
// ----------------------------------------------------------------------------

// intIdToJSONString converts a Foreman int type ID to a String.
//
// 0 is an invalid ID to Foreman - if we want to explicitly set a value to
// have 0 for an ID attribute, we need to convert it to the JavaScript "null"
// literal. We do this by returning nil for the interface, which when used
// in MarshalJSON() will product the JavaScript null literal.  Otherwise, it
// will return the string representation of the ID, since Foreman expects the
// ID to be enclosed in double quotes.
func intIdToJSONString(id int) interface{} {
	if id <= 0 {
		return nil
	}
	return strconv.Itoa(id)
}

// foremanObjectArrayToIdIntArray converts an array of ForemanObject structs
// into an integer array containing the ForemanObject's IDs.
//
// For objects that have nested or foreign-key relationships, the create and
// update operations only need to supply the object's IDs as an array.  When
// the API returns the results of the operation, Foreman includes these as an
// array of ForemanObjects.
//
// It is simpler to deal with these nested/foreign relationships by only
// managing the IDs rather than embedding the other attributes and properties
// of the object itself.  This could lead to cascading changes when an
// embedded resource is updated.
func foremanObjectArrayToIdIntArray(foa []ForemanObject) []int {
	numObjects := len(foa)
	if numObjects == 0 {
		return []int{}
	}
	intArr := make([]int, numObjects)
	for idx, val := range foa {
		intArr[idx] = val.Id
	}
	return intArr
}

// ----------------------------------------------------------------------------
// Foreman API Query Responses
// ----------------------------------------------------------------------------

// Base API query response struct.  For all "search" API calls (following the
// format /api/<resource name>), the response will be in the following format.
//
// The Results attribute will be an array of Foreman API objects from the model
// package that matched the search criteria.
type QueryResponse struct {
	// Total number of objects of that resource type in Foreman
	Total int `json:"total"`
	// Number of results matching the search criteria
	Subtotal int `json:"subtotal"`
	// Current result page (if using pagination in searches)
	Page int `json:"page"`
	// How many results to display per page (if using pagination in searches)
	PerPage int `json:"per_page"`
	// The search filter string in the form property=value&property=value&...
	Search string `json:"search,omitempty"`
	// Sorting options provided for the search
	Sort QueryResponseSort `json:"sort,omitempty"`
	// Foreman API objects that matched the search criteria for the query.
	Results []interface{} `json:"results"`
}

// Sort options as part of the generic query resposne
type QueryResponseSort struct {
	// In which manner to order results (ASC, DESC)
	Order string `json:"order,omitempty"`
	// Which field to order by
	By string `json:"by,omitempty"`
}
