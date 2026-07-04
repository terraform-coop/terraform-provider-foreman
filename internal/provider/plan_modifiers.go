package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// suppressDownloadConcurrencyDiff suppresses when API returns 0 but config has >0.
type suppressDownloadConcurrencyDiff struct{}

func (m suppressDownloadConcurrencyDiff) Description(_ context.Context) string {
	return "Suppresses diff when API returns 0 but config has >0"
}

func (m suppressDownloadConcurrencyDiff) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m suppressDownloadConcurrencyDiff) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() || req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
		return
	}

	old := req.StateValue.ValueInt64()
	new := req.PlanValue.ValueInt64()

	if shouldSuppressDownloadConcurrency(old, new) {
		resp.PlanValue = req.StateValue
	}
}

// shouldSuppressDownloadConcurrency returns true when the API returned 0 but
// the config has a positive value, which is common for fields the API omits.
func shouldSuppressDownloadConcurrency(old, new int64) bool {
	return old == 0 && new > 0
}

var _ planmodifier.Int64 = suppressDownloadConcurrencyDiff{}
