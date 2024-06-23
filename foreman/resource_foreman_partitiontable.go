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

func resourceForemanPartitionTable() *schema.Resource {
	return &schema.Resource{

		CreateContext: resourceForemanPartitionTableCreate,
		ReadContext:   resourceForemanPartitionTableRead,
		UpdateContext: resourceForemanPartitionTableUpdate,
		DeleteContext: resourceForemanPartitionTableDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		// NOTE(ALL): See the note in setResourceDataFromForemanPartitionTable -
		//   some of these attributes are not returned by the Foreman API when
		//   issuing a resource read and therefore aren't always correctly managed
		//   by Terraform

		Schema: map[string]*schema.Schema{

			autodoc.MetaAttribute: {
				Type:     schema.TypeBool,
				Computed: true,
				Description: fmt.Sprintf(
					"%s The disk partition layout of the host.",
					autodoc.MetaSummary,
				),
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
				Description: fmt.Sprintf(
					"The name of the partition table. "+
						"%s \"AutoYaST LVM\"",
					autodoc.MetaExample,
				),
			},

			"layout": {
				Type:     schema.TypeString,
				Required: true,
				Description: fmt.Sprintf(
					"The script that defines the partition table layout. "+
						"%s \"void\"",
					autodoc.MetaExample,
				),
			},

			"snippet": {
				Type:     schema.TypeBool,
				Optional: true,
				Description: "Whether or not this partition table is a snippet to be " +
					"embedded in other partition tables.",
			},

			"audit_comment": {
				Type:     schema.TypeString,
				Optional: true,
				Description: "Any audit comments to associate with the partition " +
					"table. The audit comment field is saved with the template auditing " +
					"to document the template changes.",
			},

			"locked": {
				Type:     schema.TypeBool,
				Optional: true,
				Description: "Whether or not this partition table is locked " +
					"for editing.",
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
				Description: "IDs of the operating system associated with this partition table.",
			},

			"hostgroup_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
				Description: "IDs of the hostgroups associated with this partition table.",
			},

			"host_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
				Description: "IDs of the hosts associated with this partition table.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the partition table",
			},
		},
	}
}

// -----------------------------------------------------------------------------
// Conversion Helpers
// -----------------------------------------------------------------------------

// buildForemanPartitionTable constructs a ForemanPartitionTable struct from a
// resource data reference.  The struct's members are populated from the data
// populated in the resource data.  Missing members will be left to the zero
// value for that member's type.
func buildForemanPartitionTable(d *schema.ResourceData) *api.ForemanPartitionTable {
	log.Tracef("resource_foreman_partitiontable.go#buildForemanPartitionTable")

	table := api.ForemanPartitionTable{}

	obj := buildForemanObject(d)
	table.ForemanObject = *obj

	var attr interface{}
	var ok bool

	table.Layout = d.Get("layout").(string)

	if attr, ok = d.GetOk("snippet"); ok {
		table.Snippet = attr.(bool)
	}

	if attr, ok = d.GetOk("audit_comment"); ok {
		table.AuditComment = attr.(string)
	}

	if attr, ok = d.GetOk("locked"); ok {
		table.Locked = attr.(bool)
	}

	if attr, ok = d.GetOk("os_family"); ok {
		table.OSFamily = attr.(string)
	}

	if attr, ok = d.GetOk("operatingsystem_ids"); ok {
		attrSet := attr.(*schema.Set)
		table.OperatingSystemIds = conv.InterfaceSliceToIntSlice(attrSet.List())
	}

	if attr, ok = d.GetOk("hostgroup_ids"); ok {
		attrSet := attr.(*schema.Set)
		table.HostgroupIds = conv.InterfaceSliceToIntSlice(attrSet.List())
	}

	if attr, ok = d.GetOk("host_ids"); ok {
		attrSet := attr.(*schema.Set)
		table.HostIds = conv.InterfaceSliceToIntSlice(attrSet.List())
	}

	if attr, ok = d.GetOk("description"); ok {
		table.Description = attr.(string)
	}

	return &table
}

// setResourceDataFromForemanPartitionTable sets a ResourceData's attributes
// from the attributes of the supplied ForemanPartitionTable struct
func setResourceDataFromForemanPartitionTable(d *schema.ResourceData, ft *api.ForemanPartitionTable) {
	log.Tracef("resource_foreman_partitiontable.go#setResourceDataFromForemanPartitionTable")

	d.SetId(strconv.Itoa(ft.Id))
	d.Set("name", ft.Name)
	d.Set("layout", ft.Layout)
	d.Set("os_family", ft.OSFamily)
	d.Set("operatingsystem_ids", ft.OperatingSystemIds)

	// NOTE(ALL): The following properties can be sent to the Foreman API
	//   on resource create or update, but are not returned by the Foreman API
	//   on a resource read.  For this reason, we do not save the state of these
	//   attributes in Terraform from the values of the ForemanPartitionTable
	//   struct.  Otherwise, Terraform will want to constantly update the state
	//   of these attributes since the ForemanPartitionTable is populated with
	//   the data from the return of the read call.
	//
	//   1. snippet (bool)
	//   2. locked (bool)
	//   3. audit_comment (string)
	//   4. hostgroup_ids (int array)
	//   5. host_ids (int array)

	var attr interface{}
	var ok bool

	if attr, ok = d.GetOk("snippet"); ok {
		d.Set("snippet", attr.(bool))
	}
	if attr, ok = d.GetOk("locked"); ok {
		d.Set("locked", attr.(bool))
	}
	if attr, ok = d.GetOk("audit_comment"); ok {
		d.Set("audit_comment", attr.(string))
	}
	if attr, ok = d.GetOk("hostgroup_ids"); ok {
		d.Set("hostgroup_ids", attr.(*schema.Set))
	}
	if attr, ok = d.GetOk("host_ids"); ok {
		d.Set("host_ids", attr.(*schema.Set))
	}
	if attr, ok = d.GetOk("description"); ok {
		d.Set("description", attr.(string))
	}
}

// -----------------------------------------------------------------------------
// Resource CRUD Operations
// -----------------------------------------------------------------------------

func resourceForemanPartitionTableCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_partitiontable.go#Create")

	client := meta.(*api.Client)
	t := buildForemanPartitionTable(d)

	log.Debugf("ForemanPartitionTable: [%+v]", t)

	createdTable, createErr := client.CreatePartitionTable(ctx, t)
	if createErr != nil {
		return diag.FromErr(createErr)
	}

	log.Debugf("Created ForemanPartitionTable: [%+v]", createdTable)

	setResourceDataFromForemanPartitionTable(d, createdTable)

	return nil
}

func resourceForemanPartitionTableRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_partitiontable.go#Read")

	client := meta.(*api.Client)
	t := buildForemanPartitionTable(d)

	log.Debugf("ForemanPartitionTable: [%+v]", t)

	readTable, readErr := client.ReadPartitionTable(ctx, t.Id)
	if readErr != nil {
		return diag.FromErr(api.CheckDeleted(d, readErr))
	}

	log.Debugf("Read ForemanPartitionTable: [%+v]", readTable)

	setResourceDataFromForemanPartitionTable(d, readTable)

	return nil
}

func resourceForemanPartitionTableUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_partitiontable.go#Update")

	client := meta.(*api.Client)
	t := buildForemanPartitionTable(d)

	log.Debugf("ForemanPartitionTable: [%+v]", t)

	updatedTable, updateErr := client.UpdatePartitionTable(ctx, t)
	if updateErr != nil {
		return diag.FromErr(updateErr)
	}

	log.Debugf("Updated ForemanPartitionTable: [%+v]", updatedTable)

	setResourceDataFromForemanPartitionTable(d, updatedTable)

	return nil
}

func resourceForemanPartitionTableDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_partitiontable.go#Delete")

	client := meta.(*api.Client)
	t := buildForemanPartitionTable(d)

	log.Debugf("ForemanPartitionTable: [%+v]", t)

	// NOTE(ALL): d.SetId("") is automatically called by terraform assuming delete
	//   returns no errors

	return diag.FromErr(api.CheckDeleted(d, client.DeletePartitionTable(ctx, t.Id)))
}
