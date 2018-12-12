package foreman

import (
	"fmt"
	"strconv"

	"github.com/wayfair/terraform-provider-foreman/foreman/api"
	"github.com/wayfair/terraform-provider-utils/autodoc"
	"github.com/wayfair/terraform-provider-utils/log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceForemanOperatingSystem() *schema.Resource {
	return &schema.Resource{

		Create: resourceForemanOperatingSystemCreate,
		Read:   resourceForemanOperatingSystemRead,
		Update: resourceForemanOperatingSystemUpdate,
		Delete: resourceForemanOperatingSystemDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{

			autodoc.MetaAttribute: &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
				Description: fmt.Sprintf(
					"%s Foreman representation of an operating system.",
					autodoc.MetaSummary,
				),
			},

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				Description: fmt.Sprintf(
					"Operating system name. "+
						"%s \"CentOS\"",
					autodoc.MetaExample,
				),
			},

			"major": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				Description: fmt.Sprintf(
					"Major release version. "+
						"%s \"7\"",
					autodoc.MetaExample,
				),
			},

			"minor": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Description: fmt.Sprintf(
					"Minor release version. "+
						"%s \"4\"",
					autodoc.MetaExample,
				),
			},

			"title": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
				Description: "The operating system's title is a concatentation of " +
					"the OS name, major, and minor versions to get a full operating " +
					"system release.",
			},

			"description": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Additional operating system information.",
			},

			"family": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"AIX",
					"Altlinux",
					"Archlinux",
					"Coreos",
					"Debian",
					"Freebsd",
					"Gentoo",
					"Junos",
					"NXOS",
					"Redhat",
					"Solaris",
					"Suse",
					"Windows",
					// NOTE(ALL): false - do not ignore case when comparing values
				}, false),
				Description: "Operating system family. Values include: " +
					"`\"AIX\"`, `\"Altlinux\"`, `\"Archlinux\"`, `\"Coreos\"`, " +
					"`\"Debian\"`, `\"Freebsd\"`, `\"Gentoo\"`, `\"Junos\"`, " +
					"`\"NXOS\"`, `\"Redhat\"`, `\"Solaris\"`, `\"Suse\"`, `\"Windows\"`.",
			},

			"release_name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Description: "Code name or release name for the specific operating " +
					"system version.",
			},

			"password_hash": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"MD5",
					"SHA256",
					"SHA512",
					"Base64",
					// NOTE(ALL): false - do not ignore case when comparing values
				}, false),
				Description: "Root password hash function to use. Valid values " +
					"include: `\"MD5\"`, `\"SHA256\"`, `\"SHA512\"`, `\"Base64\"`.",
			},
		},
	}
}

// -----------------------------------------------------------------------------
// Conversion Helpers
// -----------------------------------------------------------------------------

// buildForemanOperatingSystem constructs a ForemanOperatingSystem reference
// from a resource data reference.  The struct's  members are populated from
// the data populated in the resource data.  Missing members will be left to
// the zero value for that member's type.
func buildForemanOperatingSystem(d *schema.ResourceData) *api.ForemanOperatingSystem {
	log.Tracef("resource_foreman_operatingsystem.go#buildForemanOperatingSystem")

	os := api.ForemanOperatingSystem{}

	obj := buildForemanObject(d)
	os.ForemanObject = *obj

	var attr interface{}
	var ok bool

	if attr, ok = d.GetOk("major"); ok {
		os.Major = attr.(string)
	}
	if attr, ok = d.GetOk("minor"); ok {
		os.Minor = attr.(string)
	}
	if attr, ok = d.GetOk("title"); ok {
		os.Title = attr.(string)
	}
	if attr, ok = d.GetOk("description"); ok {
		os.Description = attr.(string)
	}
	if attr, ok = d.GetOk("family"); ok {
		os.Family = attr.(string)
	}
	if attr, ok = d.GetOk("release_name"); ok {
		os.ReleaseName = attr.(string)
	}
	if attr, ok = d.GetOk("password_hash"); ok {
		os.PasswordHash = attr.(string)
	}

	return &os
}

// setResourceDataFromOperatingSystem sets a ResourceData's attributes from the
// attributes of the supplied ForemanOperatingSystem reference
func setResourceDataFromForemanOperatingSystem(d *schema.ResourceData, fo *api.ForemanOperatingSystem) {
	log.Tracef("resource_foreman_operatingsystem.go#setResourceDataFromForemanOperatingSystem")

	d.SetId(strconv.Itoa(fo.Id))
	d.Set("name", fo.Name)
	d.Set("major", fo.Major)
	d.Set("minor", fo.Minor)
	d.Set("title", fo.Title)
	d.Set("description", fo.Description)
	d.Set("family", fo.Family)
	d.Set("release_name", fo.ReleaseName)
	d.Set("password_hash", fo.PasswordHash)
}

// -----------------------------------------------------------------------------
// Resource CRUD Operations
// -----------------------------------------------------------------------------

func resourceForemanOperatingSystemCreate(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_operatingsystem.go#Create")
	return nil
}

func resourceForemanOperatingSystemRead(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_operatingsystem.go#Read")

	client := meta.(*api.Client)
	o := buildForemanOperatingSystem(d)

	log.Debugf("ForemanOperatingSystem: [%+v]", o)

	readOS, readErr := client.ReadOperatingSystem(o.Id)
	if readErr != nil {
		return readErr
	}

	log.Debugf("ForemanOperatingSystem: [%+v]", readOS)

	setResourceDataFromForemanOperatingSystem(d, readOS)

	return nil
}

func resourceForemanOperatingSystemUpdate(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_operatingsystem.go#Update")
	return nil
}

func resourceForemanOperatingSystemDelete(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_operatingsystem.go#Delete")

	// NOTE(ALL): d.SetId("") is automatically called by terraform assuming delete
	//   returns no errors

	return nil
}
