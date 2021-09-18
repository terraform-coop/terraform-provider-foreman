package foreman

import (
	"fmt"
	"strconv"

	"github.com/HanseMerkur/terraform-provider-foreman/foreman/api"
	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
	"github.com/HanseMerkur/terraform-provider-utils/conv"
	"github.com/HanseMerkur/terraform-provider-utils/log"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceForemanUser() *schema.Resource {
	return &schema.Resource{

		Create: resourceForemanUserCreate,
		Read:   resourceForemanUserRead,
		Update: resourceForemanUserUpdate,
		Delete: resourceForemanUserDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{

			autodoc.MetaAttribute: &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
				Description: fmt.Sprintf(
					"%s User can be used to allow access to foreman.",
					autodoc.MetaSummary,
				),
			},

			"login": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Username used for logging-in",
			},

			"admin": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "If the user is allow admin privileges",
			},

			"firstname": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "First name of the user",
			},

			"lastname": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Last name of user",
			},

			"mail": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Email of user",
			},

			"description": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of user",
			},

			"password": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Password of user, required if auth_source_id is 1 (internal)",
			},

			"default_location_id": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Default location for the user, if empty takes global default",
			},

			"default_organization_id": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Default organization for the user, if empty takes global default",
			},

			"auth_source_id": &schema.Schema{
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      1,
				ValidateFunc: validation.IntBetween(1, 2),
				Description:  "Set the authentication source, i.e internal (1,default) or external (2)",
			},

			"locale": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Sets the timezone/location of a user",
			},

			"location_ids": &schema.Schema{
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
				Optional:    true,
				Description: "List of all locations a user has access to",
			},

			"organization_ids": &schema.Schema{
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
				Optional:    true,
				Description: "List of all organizations a user has access to",
			},
		},
	}
}

// -----------------------------------------------------------------------------
// Conversion Helpers
// -----------------------------------------------------------------------------

// buildForemanUser constructs a ForemanUser struct from a resource
// data reference. The struct's members are populated from the data populated
// in the resource data. Missing members will be left to the zero value for
// that member's type.
func buildForemanUser(d *schema.ResourceData) *api.ForemanUser {
	log.Tracef("resource_foreman_user.go#buildForemanUser")

	u := api.ForemanUser{}

	obj := buildForemanObject(d)
	u.ForemanObject = *obj

	var attr interface{}
	var ok bool

	if attr, ok = d.GetOk("login"); ok {
		u.Login = attr.(string)
	}
	if attr, ok = d.GetOk("admin"); ok {
		u.Admin = attr.(bool)
	}
	if attr, ok = d.GetOk("firstname"); ok {
		u.Firstname = attr.(string)
	}
	if attr, ok = d.GetOk("lastname"); ok {
		u.Lastname = attr.(string)
	}
	if attr, ok = d.GetOk("mail"); ok {
		u.Mail = attr.(string)
	}
	if attr, ok = d.GetOk("description"); ok {
		u.Description = attr.(string)
	}
	if attr, ok = d.GetOk("password"); ok {
		u.Password = attr.(string)
	}
	if attr, ok = d.GetOk("default_location_id"); ok {
		u.DefaultLocationId = attr.(int)
	}
	if attr, ok = d.GetOk("default_organization_id"); ok {
		u.DefaultOrganizationId = attr.(int)
	}
	if attr, ok = d.GetOk("auth_source_id"); ok {
		u.AuthSourceId = attr.(int)
	}
	if attr, ok = d.GetOk("locale"); ok {
		u.Locale = attr.(string)
	}
	if attr, ok = d.GetOk("location_ids"); ok {
		attrSet := attr.(*schema.Set)
		u.LocationIds = conv.InterfaceSliceToIntSlice(attrSet.List())
	}
	if attr, ok = d.GetOk("organization_ids"); ok {
		attrSet := attr.(*schema.Set)
		u.OrganizationIds = conv.InterfaceSliceToIntSlice(attrSet.List())
	}
	return &u
}

// setResourceDataFromForemanUser sets a ResourceData's attributes from
// the attributes of the supplied ForemanUser struct
func setResourceDataFromForemanUser(d *schema.ResourceData, fu *api.ForemanUser) {
	log.Tracef("resource_foreman_user.go#setResourceDataFromForemanUsergroup")

	d.SetId(strconv.Itoa(fu.Id))
	d.Set("login", fu.Login)
	d.Set("admin", fu.Admin)
	d.Set("firstname", fu.Firstname)
	d.Set("lastname", fu.Lastname)
	d.Set("mail", fu.Mail)
	d.Set("description", fu.Description)
	d.Set("password", fu.Password)
	d.Set("default_location_id", fu.DefaultLocationId)
	d.Set("default_organization_id", fu.DefaultOrganizationId)
	d.Set("auth_source_id", fu.AuthSourceId)
	d.Set("locale", fu.Locale)
	d.Set("location_ids", fu.LocationIds)
	d.Set("organization_ids", fu.OrganizationIds)
}

// -----------------------------------------------------------------------------
// Resource CRUD Operations
// -----------------------------------------------------------------------------

func resourceForemanUserCreate(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_user.go#Create")

	client := meta.(*api.Client)
	u := buildForemanUser(d)

	log.Debugf("ForemanUser: [%+v]", u)

	createdUser, createErr := client.CreateUser(u)
	if createErr != nil {
		return createErr
	}

	log.Debugf("Created ForemanUser: [%+v]", createdUser)

	setResourceDataFromForemanUser(d, createdUser)

	return nil
}

func resourceForemanUserRead(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_user.go#Read")

	client := meta.(*api.Client)
	u := buildForemanUser(d)

	log.Debugf("ForemanUser: [%+v]", u)

	readUser, readErr := client.ReadUser(u.Id)
	if readErr != nil {
		return readErr
	}

	log.Debugf("Read ForemanUser: [%+v]", readUser)

	setResourceDataFromForemanUser(d, readUser)

	return nil
}

func resourceForemanUserUpdate(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_user.go#Update")

	client := meta.(*api.Client)
	u := buildForemanUser(d)

	log.Debugf("ForemanUser: [%+v]", u)

	updatedUser, updateErr := client.UpdateUser(u)
	if updateErr != nil {
		return updateErr
	}

	log.Debugf("Updated ForemanUser: [%+v]", updatedUser)

	setResourceDataFromForemanUser(d, updatedUser)

	return nil
}

func resourceForemanUserDelete(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_user.go#Delete")

	client := meta.(*api.Client)
	u := buildForemanUser(d)

	log.Debugf("ForemanUser: [%+v]", u)

	// NOTE(ALL): d.SetId("") is automatically called by terraform assuming delete
	//   returns no errors
	return client.DeleteUser(u.Id)
}
