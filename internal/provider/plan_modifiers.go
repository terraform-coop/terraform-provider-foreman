package provider

import (
	"context"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// suppressDomainSuffixDiff suppresses diffs when only the domain suffix changed.
// For host.name: "host.example.com" vs "host" should not diff if domain is "example.com".
type suppressDomainSuffixDiff struct{}

func (m suppressDomainSuffixDiff) Description(_ context.Context) string {
	return "Suppresses diff when only the domain suffix changed"
}

func (m suppressDomainSuffixDiff) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m suppressDomainSuffixDiff) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() || req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
		return
	}

	old := req.StateValue.ValueString()
	new := req.PlanValue.ValueString()

	if old == new {
		return
	}

	// We can't access other attributes directly in a plan modifier, so we treat
	// the diff as a domain-suffix change only when one value is the other value
	// plus a domain suffix (e.g. "host" <-> "host.example.com"). This avoids
	// incorrectly suppressing "host.example.com" <-> "host.malicious.com".
	if old == "" || new == "" {
		if strings.HasPrefix(new, old+".") || strings.HasPrefix(old, new+".") {
			resp.PlanValue = req.StateValue
		}
		return
	}
	if strings.HasPrefix(old, new+".") || strings.HasPrefix(new, old+".") {
		resp.PlanValue = req.StateValue
	}
}

// suppressSyncDateDiff normalizes sync_date times (UTC → +0000).
type suppressSyncDateDiff struct{}

func (m suppressSyncDateDiff) Description(_ context.Context) string {
	return "Normalizes sync_date times (UTC → +0000)"
}

func (m suppressSyncDateDiff) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m suppressSyncDateDiff) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() || req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
		return
	}

	old := normalizeSyncDate(req.StateValue.ValueString())
	new := normalizeSyncDate(req.PlanValue.ValueString())

	if old == new {
		return
	}

	// Normalize "UTC" to "+0000" and compare strings directly. Using time.Parse
	// is intentionally avoided because Foreman date formats are locale-dependent.
	if old == new {
		resp.PlanValue = req.StateValue
	}
}

func normalizeSyncDate(s string) string {
	return strings.ReplaceAll(s, "UTC", "+0000")
}

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

// suppressValueTypeDiff suppresses when API doesn't return value_type.
// On existing resources, empty→"plain" should be suppressed.
type suppressValueTypeDiff struct{}

func (m suppressValueTypeDiff) Description(_ context.Context) string {
	return "Suppresses diff when API doesn't return value_type"
}

func (m suppressValueTypeDiff) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m suppressValueTypeDiff) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}

	old := req.StateValue.ValueString()
	new := req.PlanValue.ValueString()

	// API never returns value_type, so on existing resources old="" new="plain" is expected
	if old == "" && new == "plain" {
		resp.PlanValue = req.StateValue
	}
}

// intToStringPlanModifier converts a plan modifier from string to int64 context.
// Used for fields that are int in API but string in model (like download_concurrency).
type intToStringSuppressDownloadConcurrency struct{}

func (m intToStringSuppressDownloadConcurrency) Description(_ context.Context) string {
	return "Suppresses download_concurrency diff when API returns 0"
}

func (m intToStringSuppressDownloadConcurrency) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m intToStringSuppressDownloadConcurrency) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() || req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
		return
	}

	oldStr := req.StateValue.ValueString()
	newStr := req.PlanValue.ValueString()

	old, err1 := strconv.ParseInt(oldStr, 10, 64)
	new, err2 := strconv.ParseInt(newStr, 10, 64)

	if err1 != nil || err2 != nil {
		return
	}

	if shouldSuppressDownloadConcurrency(old, new) {
		resp.PlanValue = req.StateValue
	}
}

// shouldSuppressDownloadConcurrency returns true when the API returned 0 but
// the config has a positive value, which is common for fields the API omits.
func shouldSuppressDownloadConcurrency(old, new int64) bool {
	return old == 0 && new > 0
}

// Ensure interface compliance
var (
	_ planmodifier.String = suppressDomainSuffixDiff{}
	_ planmodifier.String = suppressSyncDateDiff{}
	_ planmodifier.String = suppressValueTypeDiff{}
	_ planmodifier.Int64  = suppressDownloadConcurrencyDiff{}
	_ planmodifier.String = intToStringSuppressDownloadConcurrency{}
)
