// Package helper contains schema.Schema helper functions. This package is
// designed similarly to the terraform/helper/schema package.
package helper

import (
	"github.com/hashicorp/terraform/helper/schema"
)

// DataSourceSchemaFromResourceSchema copies the schema map from a resource for
// use in a data source.  In this way, the attributes of the data source and
// resource definitions are always in-sync and reduces errors associated with
// redefining and matching the schema's between the two.  This also enables a
// data source full access to the properties of the associated resource for use
// in other resources.
//
// When copying the schema definition, the struct attributes are modified for
// use with schema.Type and schema.Description left unmodified.
func DataSourceSchemaFromResourceSchema(rs map[string]*schema.Schema) map[string]*schema.Schema {
	schemaMap := make(map[string]*schema.Schema, len(rs))

	// iterate over each attribute in the resource schema map and copy the
	// properties for use in the data sources
	for key, val := range rs {

		s := &schema.Schema{
			// Preserve type and description information
			Type:        val.Type,
			Description: val.Description,
			// Force the attribute to be a computed property
			Computed: true,
			Optional: false,
			Required: false,
		}

		// copy additional schema properties for complex types: schema.TypeList,
		// schema.TypeMap, schema.TypeSet

		// for sets, copy over the hash function.  if this is not a set, the
		// hash function will point to nil and be safe to copy
		s.Set = val.Set

		// for lists/sets, the elements can either be primitive (*schema.Schema)
		// or complex (*schema.Resource). If it is a complex element, recurse
		// and copy the schema definition.
		if elem, ok := val.Elem.(*schema.Resource); ok {
			s.Elem = &schema.Resource{
				Schema: DataSourceSchemaFromResourceSchema(elem.Schema),
			}
		} else {
			// for maps or sets/lists with non-complex typed elements,
			// copy the element definition.  For things that aren't lists, maps,
			// or sets this will be nil and is safe to copy
			s.Elem = val.Elem
		}

		schemaMap[key] = s
	}

	return schemaMap
}
