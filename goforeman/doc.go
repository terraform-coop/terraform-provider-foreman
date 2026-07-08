// Package goforeman is a Go client for the Foreman (and Katello) REST API.
//
// It exists to absorb the API's sharp edges once, so consumers - the
// terraform-provider-foreman this repository ships, or any other tool -
// don't each rediscover them against a live server. Most CRUD methods are
// generated from Foreman's own API documentation (apidoc/v2.json) by
// tools/gen/client in this repository; the quirks below are what the
// hand-written core corrects on top of what that documentation claims.
//
// # Construction
//
//	client := goforeman.NewClient(serverURL,
//	    goforeman.WithBasicAuth("admin", "changeme"),
//	    goforeman.WithTaxonomy(orgID, locID),
//	)
//	host, err := client.FindHostByName(ctx, "web01.example.com")
//
// Per resource there are Create/Read/Update/Delete methods plus
// Find<Resource>ByName (first match, (nil, nil) on miss) and
// List<Resource>s (every record, transparently paginated).
// errors.Is(err, ErrNotFound) identifies 404s.
//
// # Foreman API behavior this package absorbs
//
// URL routing: Foreman's API is spread across multiple URL namespaces -
// /api for core, /katello/api, /foreman_puppet/api, /foreman_tasks/api for
// plugins - and endpoint strings are dispatched to the right prefix here.
// Notably, the core "puppetclasses" endpoint must NOT be routed to the
// Puppet plugin namespace despite its name.
//
// Taxonomy scoping: Foreman only honors organization_id/location_id when
// they are nested inside the resource's own wrapped parameter hash (e.g.
// {"host": {"organization_id": 1, ...}}); the same fields as siblings of
// the hash are silently ignored and creation fails with "Organization
// can't be blank". WithTaxonomy injects them in the honored position on
// every create/update, and into the URL for org-scoped Katello endpoints.
//
// Async operations: several Katello endpoints return 202 with a foreman
// task instead of the resource; the client polls the task to completion
// and surfaces its failure as an error.
//
// Polymorphic parameter values: Foreman parameters are user-typed
// (string/boolean/integer/real/array/hash/yaml/json), so a parameter's
// "value" in a response is not necessarily a JSON string - fields holding
// one are declared json.RawMessage and rendered via RawValueString;
// parameter collections ([{"name": ..., "value": ...}]) go through
// ParseParameters/BuildParameters.
//
// Nested-attributes deletion: for nested collections managed through
// Rails' accepts_nested_attributes_for (a host's interfaces_attributes),
// omitting an entry from an update does NOT delete it server-side - an
// explicit {"id": ..., "_destroy": true} marker is required. See
// AppendDestroyMarkers.
//
// Boolean and foreign-key write semantics: optional booleans in request
// structs are *bool with omitempty - a plain bool can never send an
// explicit false, and an explicit JSON null violates the NOT NULL
// constraint nearly every Foreman boolean column carries. Optional
// foreign-key references (*_id) are *int64 WITHOUT omitempty, because an
// explicit null is precisely how a previously-set reference is cleared.
//
// Interface ordering: Foreman stores and returns a host's network
// interfaces sorted alphabetically by identifier, and hypervisors PXE-boot
// whichever sorts first regardless of the primary/provision flags. See
// FirstOutOfOrderIdentifier.
//
// Fields not returned on read: some attributes are accepted on write but
// never included in responses (e.g. a partition table's snippet, locked,
// audit_comment, host_ids, hostgroup_ids; a repository's
// download_concurrency reads back as 0) - readers must not treat their
// absence or zero value as "cleared".
//
// Server-side auto-association: creating/updating a resource with
// association IDs can make Foreman attach MORE members than submitted (an
// operating system created with one provisioning template gets every
// family-matched stock template and partition table associated too), and
// some associations deadlock deletion in both directions until cleared
// (subnet<->domain; DeleteSubnet clears it first). Declarative consumers
// should treat association ID lists as "ensure these are members", not as
// the exhaustive set - see also MatchNestedIDs for the id-matching needed
// to update nested collections declaratively.
//
// Odd response shapes: settings have string IDs (the setting's key) and
// user-typed values; the puppetclasses index groups results by Puppet
// environment name instead of the flat results array every other endpoint
// uses; a subnet's cidr and vlanid are documented as strings but returned
// as numbers. All confirmed against real server responses (see
// realworld_testdata/) and handled by the hand-written clients here.
package goforeman
