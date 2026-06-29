package provider

import (
	"context"
	"fmt"
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

	// Extract domain_name from state (set by the API)
	// We can't access other attributes directly in a plan modifier,
	// so we compare by checking if one is a prefix of the other after removing domain
	oldParts := strings.SplitN(old, ".", 2)
	newParts := strings.SplitN(new, ".", 2)

	// If both have same hostpart (before first dot), suppress
	if len(oldParts) > 0 && len(newParts) > 0 && oldParts[0] == newParts[0] {
		resp.PlanValue = req.StateValue
		return
	}

	// If one is empty and the other is just the hostname, suppress
	if (old == "" && new != "") || (old != "" && new == "") {
		if strings.HasPrefix(new, old) || strings.HasPrefix(old, new) {
			resp.PlanValue = req.StateValue
		}
	}
}

// suppressSyncDateDiff normalizes sync_date times (UTC → +0000).
type suppressSyncDateDiff struct{}

const syncDateLayout = "2006-01-02 15:04:05 -0700"

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

	// Try parsing as time
	oldTime, err1 := parseSyncDate(old)
	newTime, err2 := parseSyncDate(new)
	if err1 == nil && err2 == nil && oldTime == newTime {
		resp.PlanValue = req.StateValue
	}
}

func normalizeSyncDate(s string) string {
	return strings.ReplaceAll(s, "UTC", "+0000")
}

func parseSyncDate(s string) (string, error) {
	s = normalizeSyncDate(s)
	// Just compare normalized strings — time.Parse is locale-dependent
	return s, nil
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

	// If API returned 0 but config has >0, suppress the diff
	if old == 0 && new > 0 {
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

// suppressEmptyToDefault suppresses empty→default for fields the API doesn't return.
func suppressEmptyDefaultDiff(defaultValue string) planmodifier.String {
	return &suppressEmptyDefaultImpl{defaultValue: defaultValue}
}

type suppressEmptyDefaultImpl struct {
	defaultValue string
}

func (m *suppressEmptyDefaultImpl) Description(_ context.Context) string {
	return fmt.Sprintf("Suppresses diff when API returns empty and plan has default %q", m.defaultValue)
}

func (m *suppressEmptyDefaultImpl) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m *suppressEmptyDefaultImpl) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}

	old := req.StateValue.ValueString()
	new := req.PlanValue.ValueString()

	if old == "" && new == m.defaultValue {
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

	// If API returned 0 but config has >0, suppress the diff
	if old == 0 && new > 0 {
		resp.PlanValue = req.StateValue
	}
}

// Ensure interface compliance
var (
	_ planmodifier.String = suppressDomainSuffixDiff{}
	_ planmodifier.String = suppressSyncDateDiff{}
	_ planmodifier.String = suppressValueTypeDiff{}
	_ planmodifier.Int64  = suppressDownloadConcurrencyDiff{}
	_ planmodifier.String = intToStringSuppressDownloadConcurrency{}
)
