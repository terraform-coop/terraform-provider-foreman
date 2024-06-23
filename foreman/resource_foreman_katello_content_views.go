package foreman

import (
	"context"
	"fmt"
	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/api"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/utils"
	"strconv"
)

func resourceForemanKatelloContentView() *schema.Resource {
	return &schema.Resource{

		CreateContext: resourceForemanKatelloContentViewCreate,
		ReadContext:   resourceForemanKatelloContentViewRead,
		UpdateContext: resourceForemanKatelloContentViewUpdate,
		DeleteContext: resourceForemanKatelloContentViewDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			autodoc.MetaAttribute: {
				Type:     schema.TypeBool,
				Computed: true,
				Description: fmt.Sprintf(
					"%s (Composite) Content Views create an abstract view on a collection of repositories and "+
						"allow versioning of these views. Additional fine tuning can be done with package filters.",
					autodoc.MetaSummary,
				),
			},

			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Name of the (composite) content view. %s \"My new CV\"", autodoc.MetaExample),
			},

			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description for the (composite) content view",
			},

			"label": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true, // Created from name if not passed in
				ForceNew: true,
				Description: fmt.Sprintf(
					"Label for the (composite) content view. Cannot be changed after creation. "+
						"By default set to the name, with underscores as spaces replacement. %s",
					autodoc.MetaExample,
				),
			},

			"organization_id": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},

			"composite": {
				Type:          schema.TypeBool,
				Optional:      true,
				Default:       false,
				ConflictsWith: []string{"repository_ids"},
				Description:   fmt.Sprintf("Is this Content View a Composite CV? %s false", autodoc.MetaExample),
			},

			"solve_dependencies": {
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
				Description: "Relevant for Content Views: 'This will solve RPM and module stream dependencies on " +
					"every publish of this content " +
					"view. Dependency solving significantly increases publish time (publishes can take over three " +
					"times as long) and filters will be ignored when adding packages to solve dependencies. Also, " +
					"certain scenarios involving errata may still cause dependency errors.'",
			},

			"auto_publish": {
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
				Description: "Relevant for Composite Content Views: 'Automatically publish a new version of the " +
					"composite content view whenever one of its content views is published. Autopublish will only " +
					"happen for component views that use the 'Always use latest version' option.'",
			},

			"repository_ids": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
				Optional:      true,
				Computed:      true, // See DiffSuppressFunc below for more info
				ConflictsWith: []string{"composite"},
				Description:   fmt.Sprintf("List of repository IDs. %s [1, 4, 5]", autodoc.MetaExample),
				DiffSuppressFunc: func(k, oldValue, newValue string, d *schema.ResourceData) bool {
					// The following checks determine if Terraform tries to remove a list of repository IDs from
					// a composite content view. This happens, because the Katello API fills this field when a
					// CCV is created, the repo IDs from the contained CVs are inserted here.

					// Since using "composite = true" conflicts with "repository_ids", this becomes a read-only
					// field that can not be updated in the CCV itself. Yet, Terraform will think the value
					// was removed from the resource. This is partly solved by using "Computed: true" above, but
					// does not always work. Therefore, both "computed" and this diff suppression are enabled for now.

					// First check: is the diff trying to remove all repository_ids from this resource?
					if k == "repository_ids.#" && oldValue != "0" && newValue == "0" {
						// Second check: is it a composite content view?
						if composite, ok := d.GetOk("composite"); ok {
							if composite.(bool) == true {
								// If it is trying to remove all repository IDs and it is a
								// composite, suppress the diff.
								return true
							}
						}
					}
					return false
				},
			},

			"component_ids": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
				Optional:    true,
				Description: fmt.Sprintf("Relevant for CCVs: list of CV version IDs. %s [1, 4]", autodoc.MetaExample),
			},

			"latest_version_id": {
				Type:     schema.TypeInt,
				Computed: true,
				Required: false,
				Optional: false,
				Description: "Holds the ID of the latest published version of a Content View " +
					"to be used as reference in CCVs",
			},

			"filter": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Content view filters and their rules.",
				Elem:        resourceForemanKatelloContentViewFilter(),
			},

			"filtered": {
				Type:     schema.TypeBool,
				Required: false,
				Computed: true,
			},
		},
	}
}

func resourceForemanKatelloContentViewFilter() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"deb",
					"rpm",
					"package_group",
					"erratum",
					"erratum_id",
					"erratum_date",
					"docker",
					"modulemd",
				}, false),
				Description: "Type of this filter, e.g. DEB or RPM",
			},

			"inclusion": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				Description: "specifies if content should be included or excluded, " +
					"default: inclusion=false",
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"rule": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     resourceForemanKatelloContentViewFilterRule(),
			},

			//original_packages bool
			//original_module_streams bool
			//repository_ids []interface
		},
	}
}

func resourceForemanKatelloContentViewFilterRule() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"architecture": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
				Description: fmt.Sprintf("Filter pattern of this filter %s apt*",
					autodoc.MetaExample),
			},
		},
	}
}

func buildForemanKatelloContentView(d *schema.ResourceData) *api.ContentView {
	utils.TraceFunctionCall()

	cv := api.ContentView{}
	cv.ForemanObject = *buildForemanObject(d)

	cv.Description = d.Get("description").(string)
	cv.Label = d.Get("label").(string)
	cv.OrganizationId = d.Get("organization_id").(int)
	cv.Composite = d.Get("composite").(bool)
	cv.AutoPublish = d.Get("auto_publish").(bool)
	cv.SolveDependencies = d.Get("solve_dependencies").(bool)

	if filtered, ok := d.GetOk("filtered"); ok {
		cv.Filtered = filtered.(bool)
	}

	// repository_ids and component_ids are defined as "TypeList" which can
	// be any type according to Terraform docs. So we need to cast to interface and then to int.

	if repoIds, ok := d.GetOk("repository_ids"); ok {
		casted := repoIds.([]interface{})
		var ids []int
		for _, item := range casted {
			ids = append(ids, item.(int))
		}
		cv.RepositoryIds = ids
	}

	if componentIds, ok := d.GetOk("component_ids"); ok {
		casted := componentIds.([]interface{})
		var ids []int
		for _, item := range casted {
			ids = append(ids, item.(int))
		}
		cv.ComponentIds = ids
	}

	// Handle list of ContentViewFilters
	if filters, ok := d.GetOk("filter"); ok {
		var cvfs []api.ContentViewFilter
		filters := filters.([]interface{})

		for _, cvfsResData := range filters {
			var cvf api.ContentViewFilter
			utils.Debugf("cvfsResData: %+v", cvfsResData)

			cvfsResData := cvfsResData.(map[string]interface{})

			cvf.Id = cvfsResData["id"].(int)
			cvf.Name = cvfsResData["name"].(string)
			cvf.Type = cvfsResData["type"].(string)
			cvf.Description = cvfsResData["description"].(string)
			cvf.Inclusion = cvfsResData["inclusion"].(bool)

			if rules, ok := cvfsResData["rule"]; ok {
				var cvfrs []api.ContentViewFilterRule
				rules := rules.([]interface{})

				for _, rulesResData := range rules {
					var cvfr api.ContentViewFilterRule
					rulesResData := rulesResData.(map[string]interface{})

					cvfr.Id = rulesResData["id"].(int)
					cvfr.Name = rulesResData["name"].(string)
					cvfr.Architecture = rulesResData["architecture"].(string)

					cvfrs = append(cvfrs, cvfr)
				}
				cvf.Rules = cvfrs
			}
			cvfs = append(cvfs, cvf)
		}
		cv.Filters = cvfs
	}

	return &cv
}

func setResourceDataFromForemanKatelloContentView(d *schema.ResourceData, cv *api.ContentView) {
	utils.TraceFunctionCall()

	d.SetId(strconv.Itoa(cv.Id))
	d.Set("name", cv.Name)
	d.Set("description", cv.Description)
	d.Set("label", cv.Label)
	d.Set("organization_id", cv.OrganizationId)
	d.Set("composite", cv.Composite)
	d.Set("auto_publish", cv.AutoPublish)
	d.Set("solve_dependencies", cv.SolveDependencies)
	d.Set("filtered", cv.Filtered)
	d.Set("repository_ids", cv.RepositoryIds)
	d.Set("component_ids", cv.ComponentIds)

	// Handle ContentViewFilters and their ContentViewFilterRules

	//hashSetFuncFilters := schema.HashResource(resourceForemanKatelloContentViewFilter())
	//hashSetFuncFilterRules := schema.HashResource(resourceForemanKatelloContentViewFilterRule())

	filterSet := make([]interface{}, len(cv.Filters))
	for idx, item := range cv.Filters {
		newFilter := map[string]interface{}{
			"id":          item.Id,
			"name":        item.Name,
			"type":        item.Type,
			"inclusion":   item.Inclusion,
			"description": item.Description,
			"rule":        nil,
		}

		ruleSet := make([]interface{}, len(item.Rules))
		for idx2, item2 := range item.Rules {
			newRule := map[string]interface{}{
				"id":           item2.Id,
				"name":         item2.Name,
				"architecture": item2.Architecture,
			}
			ruleSet[idx2] = newRule
		}
		//srs := schema.NewSet(hashSetFuncFilterRules, ruleSet)

		newFilter["rule"] = ruleSet

		filterSet[idx] = newFilter
	}

	//sfs := schema.NewSet(hashSetFuncFilters, filterSet)

	err := d.Set("filter", filterSet)
	if err != nil {
		panic(err)
	}

	// Latest published version
	latest_published_version := 0

	if cv.LatestVersionId != 0 {
		// Try using the dedicated field for LatestVersionId first
		latest_published_version = cv.LatestVersionId
	} else {
		// If that fails, try finding the highest ID of a version from the CV's versions
		for _, version := range cv.Versions {
			if version.Id > latest_published_version {
				latest_published_version = version.Id
			}
		}
	}

	// If a latest version ID was defined, put it into the Terraform state field
	if latest_published_version > 0 {
		err = d.Set("latest_version_id", latest_published_version)
		if err != nil {
			panic(err)
		}
	}
}

func resourceForemanKatelloContentViewCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	cv := buildForemanKatelloContentView(d)
	utils.Debugf("cv: %+v", cv)

	createdCv, err := client.CreateKatelloContentView(ctx, cv)
	if err != nil {
		return diag.FromErr(err)
	}
	utils.Debugf("createdCv: %+v", createdCv)

	setResourceDataFromForemanKatelloContentView(d, createdCv)
	return nil
}

func resourceForemanKatelloContentViewRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	cv := buildForemanKatelloContentView(d)

	readCv, err := client.ReadKatelloContentView(ctx, cv)
	if err != nil {
		return diag.FromErr(api.CheckDeleted(d, err))
	}
	utils.Debugf("readCv: %+v", readCv)

	setResourceDataFromForemanKatelloContentView(d, readCv)
	return nil
}

func resourceForemanKatelloContentViewUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	cv := buildForemanKatelloContentView(d)
	utils.Debugf("cv: [%+v]", cv)

	updatedCv, err := client.UpdateKatelloContentView(ctx, cv)
	if err != nil {
		return diag.FromErr(err)
	}
	utils.Debugf("updatedCv: %+v", updatedCv)

	setResourceDataFromForemanKatelloContentView(d, updatedCv)
	return nil
}

func resourceForemanKatelloContentViewDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	cv := buildForemanKatelloContentView(d)

	utils.Debugf("cv to be deleted: %+v", cv)

	return diag.FromErr(api.CheckDeleted(d, client.DeleteKatelloContentView(ctx, cv.Id)))
}
