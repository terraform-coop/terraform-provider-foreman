package foreman

import (
	"context"
	"fmt"
	"strconv"

	"github.com/HanseMerkur/terraform-provider-foreman/foreman/api"
	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
	"github.com/HanseMerkur/terraform-provider-utils/conv"
	"github.com/HanseMerkur/terraform-provider-utils/log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceForemanOperatingSystem() *schema.Resource {
	return &schema.Resource{

		CreateContext: resourceForemanOperatingSystemCreate,
		ReadContext:   resourceForemanOperatingSystemRead,
		UpdateContext: resourceForemanOperatingSystemUpdate,
		DeleteContext: resourceForemanOperatingSystemDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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
				Default:  "SHA512",
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
			"provisioning_templates": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				MinItems: 1,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
				Description: "Identifiers of attached provisioning templates",
			},
			"media": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				MinItems: 1,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
				Description: "Identifiers of attached media",
			},
			"architectures": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				MinItems: 1,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
				Description: "Identifiers of attached architectures",
			},
			"partitiontables": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				MinItems: 1,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
				Description: "Identifiers of attached partition tables",
			},
			"parameters": &schema.Schema{
				Type:     schema.TypeMap,
				ForceNew: false,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "A map of parameters that will be saved as operating system parameters " +
					"in the os config.",
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
	if attr, ok = d.GetOk("provisioning_templates"); ok {
		attrSet := attr.(*schema.Set)
		os.ProvisioningTemplateIds = conv.InterfaceSliceToIntSlice(attrSet.List())
	}
	if attr, ok = d.GetOk("media"); ok {
		attrSet := attr.(*schema.Set)
		os.MediumIds = conv.InterfaceSliceToIntSlice(attrSet.List())
	}
	if attr, ok = d.GetOk("architectures"); ok {
		attrSet := attr.(*schema.Set)
		os.ArchitectureIds = conv.InterfaceSliceToIntSlice(attrSet.List())
	}
	if attr, ok = d.GetOk("partitiontables"); ok {
		attrSet := attr.(*schema.Set)
		os.PartitiontableIds = conv.InterfaceSliceToIntSlice(attrSet.List())
	}
	if attr, ok = d.GetOk("parameters"); ok {
		os.OperatingSystemParameters = api.ToKV(attr.(map[string]interface{}))
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
	d.Set("provisioning_templates", fo.ProvisioningTemplateIds)
	d.Set("media", fo.MediumIds)
	d.Set("architectures", fo.ArchitectureIds)
	d.Set("partitiontables", fo.PartitiontableIds)
	d.Set("parameters", api.FromKV(fo.OperatingSystemParameters))
}

// -----------------------------------------------------------------------------
// Resource CRUD Operations
// -----------------------------------------------------------------------------

func resourceForemanOperatingSystemCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_operatingsystem.go#Create")

	client := meta.(*api.Client)
	o := buildForemanOperatingSystem(d)

	createdOs, createErr := client.CreateOperatingSystem(ctx, o)
	if createErr != nil {
		return diag.FromErr(createErr)
	}

	log.Debugf("Created ForemanOperatingSystem: [%+v]", createdOs)

	setResourceDataFromForemanOperatingSystem(d, createdOs)

	return nil
}

func resourceForemanOperatingSystemRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_operatingsystem.go#Read")

	client := meta.(*api.Client)
	o := buildForemanOperatingSystem(d)

	log.Debugf("ForemanOperatingSystem: [%+v]", o)

	readOS, readErr := client.ReadOperatingSystem(ctx, o.Id)
	if readErr != nil {
		return diag.FromErr(readErr)
	}

	log.Debugf("ForemanOperatingSystem: [%+v]", readOS)

	setResourceDataFromForemanOperatingSystem(d, readOS)

	return nil
}

func resourceForemanOperatingSystemUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_operatingsystem.go#Update")

	client := meta.(*api.Client)
	o := buildForemanOperatingSystem(d)

	log.Debugf("ForemanOperatingSystem: [%+v]", o)

	updatedOs, updateErr := client.UpdateOperatingSystem(ctx, o)
	if updateErr != nil {
		return diag.FromErr(updateErr)
	}

	log.Debugf("Updated ForemanOperatingSystem: [%+v]", updatedOs)

	setResourceDataFromForemanOperatingSystem(d, updatedOs)

	return nil
}

func resourceForemanOperatingSystemDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_operatingsystem.go#Delete")

	client := meta.(*api.Client)
	o := buildForemanOperatingSystem(d)

	log.Debugf("ForemanOperatingSystem: [%+v]", o)

	// NOTE(ALL): d.SetId("") is automatically called by terraform assuming delete
	//   returns no errors

	return diag.FromErr(client.DeleteOperatingSystem(ctx, o.Id))
}
