package main

// This package parses the Terraform provider by importing it and fetching the schema.
// Works with both resources and data sources.
// Then prints out the schema's fields and meta data (type, required, computed etc.).

// Parsing fields of types TypeList and TypeSet is more complex and WIP. Please refer to
// the provider documentation to get insight into the elements of these fields.

// One thing to note is that importing the provider together with the "schema" package from
// Terraform SDK does not work and is unsupported. This blocks parsing of deeper elements
// in the schema, e.g. TypeList and TypeSet on a "native" basis.
// See https://github.com/hashicorp/terraform-plugin-sdk/issues/268 for further info.

import (
	"fmt"
	"github.com/terraform-coop/terraform-provider-foreman/foreman"
	"reflect"
	"slices"
)

func checkFieldExists(f reflect.Value) bool {
	if f == (reflect.Value{}) {
		return false
	}
	return true
}

func main() {
	var schemaKeysIgnore = []string{"__meta__"}
	provider := foreman.Provider()

	for name, resource := range provider.ResourcesMap {
		fmt.Printf("Resource '%s'\n", name)
		for schemaKey, schema := range resource.Schema {
			if slices.Contains(schemaKeysIgnore, schemaKey) {
				continue
			}

			if (!schema.Required && !schema.Optional) || schema.Computed {
				continue
			}

			fmt.Printf("- '%s': Type %v, required %t, optional %t, computed %t, forcenew %t, default '%v'\n",
				schemaKey, schema.Type, schema.Required, schema.Optional, schema.Computed, schema.ForceNew, schema.Default,
			)

			// Handle TypeList and TypeSet
			if schema.Elem != nil {
				val := reflect.ValueOf(schema.Elem)
				val = reflect.Indirect(val)

				fieldsInElem := []string{}

				vt := val.Type()
				for i := 0; i < vt.NumField(); i++ {
					field := vt.Field(i)
					// Skip embedded fields
					if field.Anonymous {
						continue
					}
					fieldsInElem = append(fieldsInElem, field.Name)
				}

				// This is WIP
				// fmt.Printf("   Elem fields:\n", )
				// for _, item := range fieldsInElem {
				//~ fmt.Printf("   - %s\n", item)
				// }

                                // Handle "type"
				f_type := val.FieldByName("Type")
				if !checkFieldExists(f_type) {
					continue
				}

                                // Handle "required"
				f_req := val.FieldByName("Required")
				if !checkFieldExists(f_req) {
					continue
				}
				if f_req.Kind() == reflect.Bool && !f_req.Bool() {
					continue
				}
			}
		}
		fmt.Println()
	}

	for name, datasource := range provider.DataSourcesMap {
		fmt.Printf("Data source: '%s'\n", name)
		for schemaKey, schema := range datasource.Schema {
			if slices.Contains(schemaKeysIgnore, schemaKey) {
				continue
			}
			fmt.Printf("- '%s': Type %v, required %t, optional %t, computed %t, forcenew %t, default '%v'\n",
				schemaKey, schema.Type, schema.Required, schema.Optional, schema.Computed, schema.ForceNew, schema.Default,
			)
		}
		fmt.Println()
	}
}
