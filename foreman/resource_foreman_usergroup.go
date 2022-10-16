package foreman

import (
	"context"
	"fmt"
	"strconv"

	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
	"github.com/HanseMerkur/terraform-provider-utils/log"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/api"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceForemanUsergroup() *schema.Resource {
	return &schema.Resource{

		CreateContext: resourceForemanUsergroupCreate,
		ReadContext:   resourceForemanUsergroupRead,
		UpdateContext: resourceForemanUsergroupUpdate,
		DeleteContext: resourceForemanUsergroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{

			autodoc.MetaAttribute: {
				Type:     schema.TypeBool,
				Computed: true,
				Description: fmt.Sprintf(
					"%s Usergroups can be used to organize permissions and ownership of the hosts.",
					autodoc.MetaSummary,
				),
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
				Description: fmt.Sprintf(
					"Usergroup name. "+
						"%s \"compute\"",
					autodoc.MetaExample,
				),
			},

			"admin": {
				Type:     schema.TypeBool,
				Optional: true,
				Description: fmt.Sprintf(
					"Is an admin user group."+
						"%s true",
					autodoc.MetaExample,
				),
			},
		},
	}
}

// -----------------------------------------------------------------------------
// Conversion Helpers
// -----------------------------------------------------------------------------

// buildForemanUsergroup constructs a ForemanUsergroup struct from a resource
// data reference. The struct's members are populated from the data populated
// in the resource data. Missing members will be left to the zero value for
// that member's type.
func buildForemanUsergroup(d *schema.ResourceData) *api.ForemanUsergroup {
	log.Tracef("resource_foreman_usergroup.go#buildForemanUsergroup")

	usergroup := api.ForemanUsergroup{}

	obj := buildForemanObject(d)
	usergroup.ForemanObject = *obj

	var attr interface{}
	var ok bool

	if attr, ok = d.GetOk("name"); ok {
		usergroup.Name = attr.(string)
	}

	usergroup.Admin = d.Get("admin").(bool)

	return &usergroup
}

// setResourceDataFromForemanUsergroup sets a ResourceData's attributes from
// the attributes of the supplied ForemanUsergroup struct
func setResourceDataFromForemanUsergroup(d *schema.ResourceData, fh *api.ForemanUsergroup) {
	log.Tracef("resource_foreman_usergroup.go#setResourceDataFromForemanUsergroup")

	d.SetId(strconv.Itoa(fh.Id))
	d.Set("name", fh.Name)
	d.Set("admin", fh.Admin)
}

// -----------------------------------------------------------------------------
// Resource CRUD Operations
// -----------------------------------------------------------------------------

func resourceForemanUsergroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_usergroup.go#Create")

	client := meta.(*api.Client)
	h := buildForemanUsergroup(d)

	log.Debugf("ForemanUsergroup: [%+v]", h)

	createdUsergroup, createErr := client.CreateUsergroup(ctx, h)
	if createErr != nil {
		return diag.FromErr(createErr)
	}

	log.Debugf("Created ForemanUsergroup: [%+v]", createdUsergroup)

	setResourceDataFromForemanUsergroup(d, createdUsergroup)

	return nil
}

func resourceForemanUsergroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_usergroup.go#Read")

	client := meta.(*api.Client)
	h := buildForemanUsergroup(d)

	log.Debugf("ForemanUsergroup: [%+v]", h)

	readUsergroup, readErr := client.ReadUsergroup(ctx, h.Id)
	if readErr != nil {
		return diag.FromErr(api.CheckDeleted(d, readErr))
	}

	log.Debugf("Read ForemanUsergroup: [%+v]", readUsergroup)

	setResourceDataFromForemanUsergroup(d, readUsergroup)

	return nil
}

func resourceForemanUsergroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_usergroup.go#Update")

	// TODO(ALL): 404 errors here (for v.1.11.4 ) - i think we need to
	//   concatentate the id with the name, replacing forward slash with a dash?
	//   getting weird behavior when updating a usergroup aside from updating the
	//   usergroup's name

	client := meta.(*api.Client)
	h := buildForemanUsergroup(d)

	log.Debugf("ForemanUsergroup: [%+v]", h)

	updatedUsergroup, updateErr := client.UpdateUsergroup(ctx, h)
	if updateErr != nil {
		return diag.FromErr(updateErr)
	}

	log.Debugf("Updated ForemanUsergroup: [%+v]", updatedUsergroup)

	setResourceDataFromForemanUsergroup(d, updatedUsergroup)

	return nil
}

func resourceForemanUsergroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_usergroup.go#Delete")

	client := meta.(*api.Client)
	h := buildForemanUsergroup(d)

	log.Debugf("ForemanUsergroup: [%+v]", h)

	// NOTE(ALL): d.SetId("") is automatically called by terraform assuming delete
	//   returns no errors
	return diag.FromErr(api.CheckDeleted(d, client.DeleteUsergroup(ctx, h.Id)))
}
