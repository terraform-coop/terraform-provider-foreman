package foreman

import (
	"strconv"

	"github.com/wayfair/terraform-provider-foreman/foreman/api"

	"github.com/hashicorp/terraform/helper/schema"
)

// buildForemanObject constructs a base ForemanObject reference from a
// ResourceData reference.  The struct's members are populated from the data
// populated in the ResourceData.  Missing members will be left to the zero
// value for that members type.
func buildForemanObject(d *schema.ResourceData) *api.ForemanObject {
	obj := api.ForemanObject{}

	var err error
	if obj.Id, err = strconv.Atoi(d.Id()); err != nil {
		obj.Id = 0
	}

	var attr interface{}
	var ok bool

	if attr, ok = d.GetOk("name"); ok {
		if obj.Name, ok = attr.(string); !ok {
			obj.Name = ""
		}
	}
	if attr, ok = d.GetOk("created_at"); ok {
		if obj.CreatedAt, ok = attr.(string); !ok {
			obj.CreatedAt = ""
		}
	}
	if attr, ok = d.GetOk("updated_at"); ok {
		if obj.UpdatedAt, ok = attr.(string); !ok {
			obj.UpdatedAt = ""
		}
	}

	return &obj
}
