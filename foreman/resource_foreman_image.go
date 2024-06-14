package foreman

import (
	"context"
	"fmt"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/utils"
	"strconv"
	"strings"

	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/api"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceForemanImage() *schema.Resource {
	return &schema.Resource{

		CreateContext: resourceForemanImageCreate,
		ReadContext:   resourceForemanImageRead,
		UpdateContext: resourceForemanImageUpdate,
		DeleteContext: resourceForemanImageDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{

			autodoc.MetaAttribute: {
				Type:     schema.TypeBool,
				Computed: true,
				Description: fmt.Sprintf(
					"%s Foreman representation of image.",
					autodoc.MetaSummary,
				),
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the image to be used in Foreman",
			},
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Username used to log into the newly created machine that is based on this image",
			},
			"uuid": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "UUID of the image from the compute resource",
			},
			"compute_resource_id": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "ID of the compute resource in Foreman",
			},
			"operatingsystem_id": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "ID of the operating system in Foreman",
			},
			"architecture_id": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "ID of the architecture in Foreman",
			},
			"user_data": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Does the image support user data (cloud-init etc.)?",
			},
		},
	}
}

// -----------------------------------------------------------------------------
// Conversion Helpers
// -----------------------------------------------------------------------------

// buildForemanImage constructs a ForemanImage reference from a resource data
// reference.  The struct's  members are populated from the data populated in
// the resource data.  Missing members will be left to the zero value for that
// member's type.
func buildForemanImage(d *schema.ResourceData) *api.ForemanImage {
	utils.TraceFunctionCall()

	image := api.ForemanImage{}

	obj := buildForemanObject(d)
	image.ForemanObject = *obj

	var attr interface{}
	var ok bool

	if attr, ok = d.GetOk("name"); ok {
		image.Name = attr.(string)
	}
	if attr, ok = d.GetOk("uuid"); ok {
		image.UUID = attr.(string)
	}
	if attr, ok = d.GetOk("username"); ok {
		image.Username = attr.(string)
	}
	if attr, ok = d.GetOk("operatingsystem_id"); ok {
		image.OperatingSystemID = attr.(int)
	}
	if attr, ok = d.GetOk("architecture_id"); ok {
		image.ArchitectureID = attr.(int)
	}
	if attr, ok = d.GetOk("compute_resource_id"); ok {
		image.ComputeResourceID = attr.(int)
	}
	if attr, ok = d.GetOk("user_data"); ok {
		image.UserData = attr.(bool)
	}

	return &image
}

// setResourceDataFromForemanImage sets a ResourceData's attributes from the
// attributes of the supplied ForemanImage reference
func setResourceDataFromForemanImage(d *schema.ResourceData, fd *api.ForemanImage) {
	utils.TraceFunctionCall()

	d.SetId(strconv.Itoa(fd.Id))
	d.Set("name", fd.Name)
	d.Set("username", fd.Username)
	d.Set("uuid", fd.UUID)
	d.Set("operatingsystem_id", fd.OperatingSystemID)
	d.Set("architecture_id", fd.ArchitectureID)
	d.Set("compute_resource_id", fd.ComputeResourceID)
	d.Set("user_data", fd.UserData)
}

// -----------------------------------------------------------------------------
// Resource CRUD Operations
// -----------------------------------------------------------------------------

func resourceForemanImageCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	img := buildForemanImage(d)

	utils.Debugf("img: [%+v]", img)

	createdImage, createErr := client.CreateImage(ctx, img, img.ComputeResourceID)
	if createErr != nil {
		utils.Debugf("%+v", createErr)

		isUuidError := strings.Contains(createErr.(api.HTTPError).RespBody, "UUID has already been taken")
		if createErr.(api.HTTPError).StatusCode == 422 && isUuidError {
			return diag.Errorf("You cannot use the same UUID for multiple images: '%s' is already taken by another Foreman image", img.UUID)
		}

		return diag.FromErr(createErr)
	}

	setResourceDataFromForemanImage(d, createdImage)

	return nil
}

func resourceForemanImageRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	image := buildForemanImage(d)

	utils.Debugf("ForemanImage: [%+v]", image)

	readImage, readErr := client.ReadImage(ctx, image)
	if readErr != nil {
		return diag.FromErr(api.CheckDeleted(d, readErr))
	}

	utils.Debugf("Read ForemanImage: [%+v]", readImage)

	setResourceDataFromForemanImage(d, readImage)

	return nil
}

func resourceForemanImageUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	img := buildForemanImage(d)

	updatedImage, updateErr := client.UpdateImage(ctx, img)
	if updateErr != nil {
		isUuidError := strings.Contains(updateErr.(api.HTTPError).RespBody, "UUID has already been taken")
		if updateErr.(api.HTTPError).StatusCode == 422 && isUuidError {
			return diag.Errorf("You cannot use the same UUID for multiple images: '%s' is already taken by another Foreman image", img.UUID)
		}

		return diag.FromErr(updateErr)
	}

	setResourceDataFromForemanImage(d, updatedImage)

	return nil
}

func resourceForemanImageDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	image := buildForemanImage(d)

	delErr := client.DeleteImage(ctx, image.ComputeResourceID, image.Id)
	if delErr != nil {
		return diag.FromErr(delErr)
	}

	// NOTE(ALL): d.SetId("") is automatically called by terraform assuming delete
	//   returns no errors

	return nil
}
