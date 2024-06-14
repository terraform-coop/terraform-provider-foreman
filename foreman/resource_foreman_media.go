package foreman

import (
	"context"
	"fmt"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/utils"
	"strconv"

	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
	"github.com/HanseMerkur/terraform-provider-utils/conv"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/api"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceForemanMedia() *schema.Resource {
	return &schema.Resource{

		CreateContext: resourceForemanMediaCreate,
		ReadContext:   resourceForemanMediaRead,
		UpdateContext: resourceForemanMediaUpdate,
		DeleteContext: resourceForemanMediaDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{

			autodoc.MetaAttribute: {
				Type:     schema.TypeBool,
				Computed: true,
				Description: fmt.Sprintf(
					"%s Remote installation media.",
					autodoc.MetaSummary,
				),
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
				Description: fmt.Sprintf(
					"Name of the media. "+
						"%s \"CentOS mirror\"",
					autodoc.MetaExample,
				),
			},

			"path": {
				Type:     schema.TypeString,
				Required: true,
				Description: fmt.Sprintf(
					"The path to the medium, can be a URL or a valid NFS server (exclusive "+
						"of the architecture).  For example:\n"+
						"\nhttp://mirror.centos.org/centos/$version/os/$arch\n\n"+
						"Where $arch will be substituted for the host's actual OS architecture "+
						"and $version, $major, $minor will be substituted for the version of the "+
						"operating system. \n"+
						"\nSolaris and Debian media may also use $release. "+
						"%s \"http://mirror.averse.net/centos/$major.$minor/os/$arch\"",
					autodoc.MetaExample,
				),
			},

			"os_family": {
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

			// -- Foreign Key Relationships --

			"operatingsystem_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
				Description: "IDs of the operating systems associated with this media.",
			},
		},
	}
}

// -----------------------------------------------------------------------------
// Conversion Helpers
// -----------------------------------------------------------------------------

// buildForemanMedia constucts a ForemanMedia struct from a resource data
// reference.  The struct's members are populated from the data populated in
// the resource data.  Missing members will be left to the zero value for that
// member's type.
func buildForemanMedia(d *schema.ResourceData) *api.ForemanMedia {
	utils.TraceFunctionCall()

	media := api.ForemanMedia{}

	obj := buildForemanObject(d)
	media.ForemanObject = *obj

	var attr interface{}
	var ok bool

	media.Path = d.Get("path").(string)

	if attr, ok = d.GetOk("os_family"); ok {
		media.OSFamily = attr.(string)
	}

	if attr, ok = d.GetOk("operatingsystem_ids"); ok {
		attrSet := attr.(*schema.Set)
		media.OperatingSystemIds = conv.InterfaceSliceToIntSlice(attrSet.List())
	}

	return &media
}

// setResourceDataFromForemanMedia sets a ResourceData's attributes from the
// attributes of the supplied ForemanMedia struct.
func setResourceDataFromForemanMedia(d *schema.ResourceData, fm *api.ForemanMedia) {
	utils.TraceFunctionCall()

	d.SetId(strconv.Itoa(fm.Id))
	d.Set("name", fm.Name)
	d.Set("path", fm.Path)
	d.Set("os_family", fm.OSFamily)
	d.Set("operatingsystem_ids", fm.OperatingSystemIds)
}

// -----------------------------------------------------------------------------
// Resource CRUD Operations
// -----------------------------------------------------------------------------

func resourceForemanMediaCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	m := buildForemanMedia(d)

	utils.Debugf("ForemanMedia: [%+v]", m)

	createdMedia, createErr := client.CreateMedia(ctx, m)
	if createErr != nil {
		return diag.FromErr(createErr)
	}

	utils.Debugf("Created ForemanMedia: [%+v]", createdMedia)

	setResourceDataFromForemanMedia(d, createdMedia)

	return nil
}

func resourceForemanMediaRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	m := buildForemanMedia(d)

	utils.Debugf("ForemanMedia: [%+v]", m)

	readMedia, readErr := client.ReadMedia(ctx, m.Id)
	if readErr != nil {
		return diag.FromErr(api.CheckDeleted(d, readErr))
	}

	utils.Debugf("Read ForemanMedia: [%+v]", readMedia)

	setResourceDataFromForemanMedia(d, readMedia)

	return nil
}

func resourceForemanMediaUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	m := buildForemanMedia(d)

	utils.Debugf("ForemanMedia: [%+v]", m)

	updatedMedia, updateErr := client.UpdateMedia(ctx, m)
	if updateErr != nil {
		return diag.FromErr(updateErr)
	}

	utils.Debugf("Updated ForemanMedia: [%+v]", updatedMedia)

	setResourceDataFromForemanMedia(d, updatedMedia)

	return nil
}

func resourceForemanMediaDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	m := buildForemanMedia(d)

	utils.Debugf("ForemanMedia: [%+v]", m)

	// NOTE(ALL): d.SetId("") is automatically called by terraform assuming delete
	//   returns no errors
	return diag.FromErr(api.CheckDeleted(d, client.DeleteMedia(ctx, m.Id)))
}
