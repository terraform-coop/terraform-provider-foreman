package goforeman

// This file holds Foreman/Rails API conventions that consumers building on
// this client (the Terraform provider included) would otherwise each have
// to rediscover the hard way.

// AppendDestroyMarkers implements Rails' accepts_nested_attributes_for
// deletion convention for nested collection attributes (e.g. a host's
// interfaces_attributes): on update, Foreman silently IGNORES an entry
// that is simply missing from the submitted array - actually removing one
// requires an explicit {"id": <id>, "_destroy": true} entry (undocumented,
// confirmed against a real server). This returns planItems plus a destroy
// marker for every id present in priorItems but absent from planItems.
// Entries without a usable numeric id (e.g. not yet created server-side)
// are ignored on both sides.
func AppendDestroyMarkers(planItems, priorItems []map[string]interface{}) []map[string]interface{} {
	out := planItems

	planIDs := make(map[int64]bool, len(planItems))
	for _, m := range planItems {
		if id := numericID(m["id"]); id != 0 {
			planIDs[id] = true
		}
	}

	for _, m := range priorItems {
		id := numericID(m["id"])
		if id == 0 || planIDs[id] {
			continue
		}
		out = append(out, map[string]interface{}{"id": id, "_destroy": true})
	}

	return out
}

// numericID coerces the loosely-typed "id" value of a nested-attributes map
// (int64 from this provider's own bridging, float64 from generic
// json.Unmarshal, int from hand-built literals) to int64; anything else is 0.
func numericID(v interface{}) int64 {
	switch id := v.(type) {
	case int64:
		return id
	case float64:
		return int64(id)
	case int:
		return int64(id)
	default:
		return 0
	}
}

// FirstOutOfOrderIdentifier returns the first adjacent pair of host
// interface identifiers that violate ascending alphabetical order, or
// ("", "", false) if ids is already sorted.
//
// Why callers care: Foreman stores and returns a host's network interfaces
// sorted alphabetically by "identifier", and hypervisors (confirmed for
// VMware and libvirt) always use whichever interface is FIRST in that
// order for PXE boot, regardless of the primary/provision flags. A client
// submitting interfaces in any other order gets them silently reordered on
// read-back - and possibly the wrong interface PXE-booted. See
// https://github.com/terraform-coop/terraform-provider-foreman/issues/126.
func FirstOutOfOrderIdentifier(ids []string) (prev, next string, found bool) {
	for i := 1; i < len(ids); i++ {
		if ids[i] < ids[i-1] {
			return ids[i-1], ids[i], true
		}
	}
	return "", "", false
}

// MatchNestedIDs copies server-assigned "id"s from priorItems onto planItems
// that lack one, matched by the collection's natural key (a host
// interface's "identifier"). Rails' accepts_nested_attributes_for treats an
// entry without an id as a NEW record: re-submitting an existing entry
// id-less makes Foreman try to create a duplicate and fail validation
// ("Identifier has already been taken", confirmed against a real server) -
// so declarative callers that only know the desired state must re-attach
// ids before submitting an update. Pair with AppendDestroyMarkers.
func MatchNestedIDs(planItems, priorItems []map[string]interface{}, key string) []map[string]interface{} {
	priorByKey := make(map[interface{}]int64, len(priorItems))
	for _, m := range priorItems {
		if id := numericID(m["id"]); id != 0 && m[key] != nil {
			priorByKey[m[key]] = id
		}
	}
	for _, m := range planItems {
		if numericID(m["id"]) == 0 && m[key] != nil {
			if id, ok := priorByKey[m[key]]; ok {
				m["id"] = id
			}
		}
	}
	return planItems
}
