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
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: fmt.Sprintf("Is this Content View a Composite CV? %s false", autodoc.MetaExample),
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
				Optional:    true,
				Description: fmt.Sprintf("List of repository IDs. %s [1, 4, 5]", autodoc.MetaExample),
			},

			"component_ids": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
				Optional:    true,
				Description: fmt.Sprintf("Relevant for CCVs: list of CV IDs. %s [1, 4]", autodoc.MetaExample),
			},

			"filter": {
				Type:        schema.TypeSet,
				Required:    false,
				Optional:    true,
				Computed:    true,
				Description: "Content view filters and their rules. Currently read-only, to be used as data source",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
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
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
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
							},
						},

						//original_packages bool
						//original_module_streams bool
						//repository_ids []interface
					},
				},
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
		filters := filters.(*schema.Set)

		for _, cvfsResData := range filters.List() {
			var cvf api.ContentViewFilter
			utils.Debugf("cvfsResData: %+v", cvfsResData)

			cvfsResData := cvfsResData.(map[string]interface{})

			cvf.Name = cvfsResData["name"].(string)
			cvf.Type = cvfsResData["type"].(string)
			cvf.Description = cvfsResData["description"].(string)
			cvf.Inclusion = cvfsResData["inclusion"].(bool)

			if rules, ok := cvfsResData["rule"]; ok {
				var cvfrs []api.ContentViewFilterRule
				rules := rules.(*schema.Set)

				for _, rulesResData := range rules.List() {
					var cvfr api.ContentViewFilterRule
					rulesResData := rulesResData.(map[string]interface{})

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
	var filterSet []map[string]interface{}
	for _, item := range cv.Filters {
		newFilter := map[string]interface{}{
			"name":        item.Name,
			"type":        item.Type,
			"inclusion":   item.Inclusion,
			"description": item.Description,
			"rule":        nil,
		}

		var ruleSet []map[string]interface{}
		for _, item2 := range item.Rules {
			newRule := map[string]interface{}{
				"name":         item2.Name,
				"architecture": item2.Architecture,
			}
			ruleSet = append(ruleSet, newRule)
		}

		newFilter["rule"] = ruleSet

		filterSet = append(filterSet, newFilter)
	}
	d.Set("filter", filterSet)
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
