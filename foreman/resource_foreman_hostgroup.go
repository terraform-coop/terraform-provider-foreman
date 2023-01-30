package foreman

import (
	"context"
	"fmt"
	"strconv"

	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
	"github.com/HanseMerkur/terraform-provider-utils/conv"
	"github.com/HanseMerkur/terraform-provider-utils/log"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/api"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceForemanHostgroup() *schema.Resource {
	return &schema.Resource{

		CreateContext: resourceForemanHostgroupCreate,
		ReadContext:   resourceForemanHostgroupRead,
		UpdateContext: resourceForemanHostgroupUpdate,
		DeleteContext: resourceForemanHostgroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{

			autodoc.MetaAttribute: {
				Type:     schema.TypeBool,
				Computed: true,
				Description: fmt.Sprintf(
					"%s Hostgroups are organized in a tree-like structure and inherit "+
						"values from their parent hostgroup(s). When hosts get associated "+
						"with a hostgroup, it will inherit attributes from the hostgroup. "+
						"This allows for easy, shared configuration of various hosts based "+
						"on common attributes.",
					autodoc.MetaSummary,
				),
			},

			"title": {
				Type:     schema.TypeString,
				Computed: true,
				Description: "The title is the fullname of the hostgroup.  A " +
					"hostgroup's title is a path-like string from the head " +
					"of the hostgroup tree down to this hostgroup.  The title will be " +
					"in the form of: \"<parent 1>/<parent 2>/.../<name>\".",
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
				Description: fmt.Sprintf(
					"Hostgroup name. "+
						"%s \"compute\"",
					autodoc.MetaExample,
				),
			},

			"root_password": {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringLenBetween(8, 256),
				Description:  "Default root password",
			},

			"pxe_loader": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				Description: "Operating system family. Value examples: " +
					"\"None\", \"PXELinux BIOS\", \"PXELinux UEFI\", \"Grub UEFI\", " +
					"\"Grub2 UEFI\", \"Grub2 UEFI SecureBoot\", \"Grub2 UEFI HTTP\", " +
					"\"Grub2 UEFI HTTPS\", \"Grub2 UEFI HTTPS SecureBoot\", " +
					"\"iPXE Embedded\", \"iPXE UEFI HTTP\", \"iPXE Chain BIOS\", " +
					"\"iPXE Chain UEFI\"",
			},
			"parameters": {
				Type:     schema.TypeMap,
				ForceNew: false,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "A map of parameters that will be saved as hostgroup parameters " +
					"in the group config.",
			},

			// -- Foreign Key Relationships --

			"architecture_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "ID of the architecture associated with this hostgroup.",
			},

			"compute_profile_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "ID of the compute profile associated with this hostgroup.",
			},

			"config_group_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
				Description: "IDs of the applied config groups.",
			},

			"content_source_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "ID of the content source associated with this hostgroup.",
			},

			"content_view_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "ID of the content view associated with this hostgroup.",
			},

			"domain_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "ID of the domain associated with this hostgroup.",
			},

			"environment_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "ID of the environment associated with this hostgroup.",
			},

			"lifecycle_environment_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "ID of the lifecycle environment associated with this hostgroup.",
			},

			"medium_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "ID of the media associated with this hostgroup.",
			},

			"operatingsystem_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "ID of the operating system associated with this hostgroup.",
			},

			"parent_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "ID of the parent hostgroup.",
			},

			"ptable_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "ID of the partition table associated with this hostgroup.",
			},

			"puppet_ca_proxy_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(0),
				Description: "ID of the smart proxy acting as the puppet certificate " +
					"authority server for this hostgroup.",
			},

			"puppet_class_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
				Description: "IDs of the applied puppet classes.",
			},

			"puppet_proxy_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(0),
				Description: "ID of the smart proxy acting as the puppet proxy " +
					"server for this hostgroup.",
			},

			"realm_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "ID of the realm associated with this hostgroup.",
			},

			"subnet_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "ID of the subnet associated with the hostgroup.",
			},
		},
	}
}

// -----------------------------------------------------------------------------
// Conversion Helpers
// -----------------------------------------------------------------------------

// buildForemanHostgroup constructs a ForemanHostgroup struct from a resource
// data reference. The struct's members are populated from the data populated
// in the resource data. Missing members will be left to the zero value for
// that member's type.
func buildForemanHostgroup(d *schema.ResourceData) *api.ForemanHostgroup {
	log.Tracef("resource_foreman_hostgroup.go#buildForemanHostgroup")

	hostgroup := api.ForemanHostgroup{}

	obj := buildForemanObject(d)
	hostgroup.ForemanObject = *obj

	var attr interface{}
	var ok bool

	if attr, ok = d.GetOk("title"); ok {
		hostgroup.Title = attr.(string)
	}

	if attr, ok = d.GetOk("root_password"); ok {
		hostgroup.RootPassword = attr.(string)
	}

	if attr, ok = d.GetOk("pxe_loader"); ok {
		hostgroup.PXELoader = attr.(string)
	}

	if attr, ok = d.GetOk("architecture_id"); ok {
		hostgroup.ArchitectureId = attr.(int)
	}

	if attr, ok = d.GetOk("compute_profile_id"); ok {
		hostgroup.ComputeProfileId = attr.(int)
	}

	if attr, ok = d.GetOk("config_group_ids"); ok {
		attrSet := attr.(*schema.Set)
		hostgroup.ConfigGroupIds = conv.InterfaceSliceToIntSlice(attrSet.List())
		hostgroup.PuppetAttributes.ConfigGroup_ids = conv.InterfaceSliceToIntSlice(attrSet.List())
	}

	if attr, ok = d.GetOk("content_source_id"); ok {
		hostgroup.ContentSourceId = attr.(int)
	}

	if attr, ok = d.GetOk("content_view_id"); ok {
		hostgroup.ContentViewId = attr.(int)
	}

	if attr, ok = d.GetOk("domain_id"); ok {
		hostgroup.DomainId = attr.(int)
	}

	if attr, ok = d.GetOk("environment_id"); ok {
		hostgroup.EnvironmentId = attr.(int)
	}

	if attr, ok = d.GetOk("lifecycle_environment_id"); ok {
		hostgroup.LifecycleId = attr.(int)
	}

	if attr, ok = d.GetOk("medium_id"); ok {
		hostgroup.MediumId = attr.(int)
	}

	if attr, ok = d.GetOk("operatingsystem_id"); ok {
		hostgroup.OperatingSystemId = attr.(int)
	}

	if attr, ok = d.GetOk("parent_id"); ok {
		hostgroup.ParentId = attr.(int)
	}

	if attr, ok = d.GetOk("ptable_id"); ok {
		hostgroup.PartitionTableId = attr.(int)
	}

	if attr, ok = d.GetOk("puppet_ca_proxy_id"); ok {
		hostgroup.PuppetCAProxyId = attr.(int)
	}

	if attr, ok = d.GetOk("puppet_class_ids"); ok {
		attrSet := attr.(*schema.Set)
		hostgroup.PuppetClassIds = conv.InterfaceSliceToIntSlice(attrSet.List())
		hostgroup.PuppetAttributes.Puppetclass_ids = conv.InterfaceSliceToIntSlice(attrSet.List())
	}

	if attr, ok = d.GetOk("puppet_proxy_id"); ok {
		hostgroup.PuppetProxyId = attr.(int)
	}

	if attr, ok = d.GetOk("realm_id"); ok {
		hostgroup.RealmId = attr.(int)
	}

	if attr, ok = d.GetOk("subnet_id"); ok {
		hostgroup.SubnetId = attr.(int)
	}
	if attr, ok = d.GetOk("parameters"); ok {
		hostgroup.HostGroupParameters = api.ToKV(attr.(map[string]interface{}))
	}

	return &hostgroup
}

// setResourceDataFromForemanHostgroup sets a ResourceData's attributes from
// the attributes of the supplied ForemanHostgroup struct
func setResourceDataFromForemanHostgroup(d *schema.ResourceData, fh *api.ForemanHostgroup) {
	log.Tracef("resource_foreman_hostgroup.go#setResourceDataFromForemanHostgroup")

	d.SetId(strconv.Itoa(fh.Id))
	d.Set("title", fh.Title)
	d.Set("name", fh.Name)
	d.Set("pxe_loader", fh.PXELoader)
	d.Set("parameters", api.FromKV(fh.HostGroupParameters))
	d.Set("architecture_id", fh.ArchitectureId)
	d.Set("compute_profile_id", fh.ComputeProfileId)
	d.Set("config_group_ids", fh.ConfigGroupIds)
	d.Set("content_source_id", fh.ContentSourceId)
	d.Set("content_view_id", fh.ContentViewId)
	d.Set("domain_id", fh.DomainId)
	d.Set("environment_id", fh.EnvironmentId)
	d.Set("lifecycle_environment_id", fh.LifecycleId)
	d.Set("medium_id", fh.MediumId)
	d.Set("operatingsystem_id", fh.OperatingSystemId)
	d.Set("parent_id", fh.ParentId)
	d.Set("ptable_id", fh.PartitionTableId)
	d.Set("puppet_ca_proxy_id", fh.PuppetCAProxyId)
	d.Set("puppet_class_ids", fh.PuppetClassIds)
	d.Set("puppet_proxy_id", fh.PuppetProxyId)
	d.Set("realm_id", fh.RealmId)
	d.Set("subnet_id", fh.SubnetId)
}

// -----------------------------------------------------------------------------
// Resource CRUD Operations
// -----------------------------------------------------------------------------

func resourceForemanHostgroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_hostgroup.go#Create")

	client := meta.(*api.Client)
	h := buildForemanHostgroup(d)

	log.Debugf("ForemanHostgroup: [%+v]", h)

	createdHostgroup, createErr := client.CreateHostgroup(ctx, h)
	if createErr != nil {
		return diag.FromErr(createErr)
	}

	log.Debugf("Created ForemanHostgroup: [%+v]", createdHostgroup)

	setResourceDataFromForemanHostgroup(d, createdHostgroup)

	return nil
}

func resourceForemanHostgroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_hostgroup.go#Read")

	client := meta.(*api.Client)
	h := buildForemanHostgroup(d)

	log.Debugf("ForemanHostgroup: [%+v]", h)

	readHostgroup, readErr := client.ReadHostgroup(ctx, h.Id)
	if readErr != nil {
		return diag.FromErr(api.CheckDeleted(d, readErr))
	}

	log.Debugf("Read ForemanHostgroup: [%+v]", readHostgroup)

	setResourceDataFromForemanHostgroup(d, readHostgroup)

	return nil
}

func resourceForemanHostgroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_hostgroup.go#Update")

	// TODO(ALL): 404 errors here (for v.1.11.4 ) - i think we need to
	//   concatentate the id with the title, replacing forward slash with a dash?
	//   getting weird behavior when updating a hostgroup aside from updating the
	//   hostgroup's name

	client := meta.(*api.Client)
	h := buildForemanHostgroup(d)

	log.Debugf("ForemanHostgroup: [%+v]", h)

	updatedHostgroup, updateErr := client.UpdateHostgroup(ctx, h)
	if updateErr != nil {
		return diag.FromErr(updateErr)
	}

	log.Debugf("Updated ForemanHostgroup: [%+v]", updatedHostgroup)

	setResourceDataFromForemanHostgroup(d, updatedHostgroup)

	return nil
}

func resourceForemanHostgroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_hostgroup.go#Delete")

	client := meta.(*api.Client)
	h := buildForemanHostgroup(d)

	log.Debugf("ForemanHostgroup: [%+v]", h)

	// NOTE(ALL): d.SetId("") is automatically called by terraform assuming delete
	//   returns no errors
	return diag.FromErr(api.CheckDeleted(d, client.DeleteHostgroup(ctx, h.Id)))
}
