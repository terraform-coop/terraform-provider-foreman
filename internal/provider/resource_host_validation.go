package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Foreman stores and returns a host's network interfaces sorted
// alphabetically by "identifier", and hypervisors (confirmed for VMware and
// libvirt) always use whichever interface is FIRST in that order for PXE
// boot, regardless of the primary/provision flags. If interfaces_attributes
// isn't written in that same order:
//   - the interface actually used for PXE boot may not be the one intended
//   - every subsequent `terraform plan` shows a perpetual reordering diff
//
// A previous attempt to fix this by switching interfaces_attributes to an
// unordered Set (matching how Terraform normally represents "order doesn't
// matter" collections) was tried in the old SDKv2 provider and reverted -
// it broke idempotency instead of fixing it. This is not a type-system
// problem, so it isn't fixed as one: this validator instead warns at
// terraform plan/apply time when the configured order doesn't match what
// Foreman will actually impose, so the mismatch is caught immediately
// instead of surfacing as a confusing perpetual diff or a silently wrong
// PXE interface. See:
// https://github.com/terraform-coop/terraform-provider-foreman/issues/126
// https://github.com/terraform-coop/terraform-provider-foreman/pull/203
var _ resource.ResourceWithValidateConfig = &hostResource{}

func (r *hostResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data hostResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if data.Interfaces.IsNull() || data.Interfaces.IsUnknown() {
		return
	}

	var identifiers []string
	for _, elem := range data.Interfaces.Elements() {
		obj, ok := elem.(types.Object)
		if !ok {
			return
		}
		idAttr, ok := obj.Attributes()["identifier"]
		if !ok {
			return
		}
		idVal, ok := idAttr.(types.String)
		if !ok || idVal.IsNull() || idVal.IsUnknown() || idVal.ValueString() == "" {
			// Can't order-check when identifier isn't statically known for
			// every interface (e.g. left for Foreman to assign, or derived
			// from another resource); skip validation rather than guess.
			return
		}
		identifiers = append(identifiers, idVal.ValueString())
	}

	if prev, next, outOfOrder := firstOutOfOrderPair(identifiers); outOfOrder {
		resp.Diagnostics.AddAttributeWarning(
			path.Root("interfaces_attributes"),
			"interfaces_attributes may not be in Foreman's expected order",
			"Foreman stores and returns host network interfaces sorted alphabetically by \"identifier\", and hypervisors "+
				"(confirmed for VMware and libvirt) always use whichever interface is FIRST in that order for PXE boot, "+
				"regardless of the primary/provision flags. This configuration's interfaces_attributes has \""+next+
				"\" after \""+prev+"\", which is out of that order - this can mean the wrong interface is used for PXE "+
				"boot, and/or every subsequent `terraform plan` will show a perpetual reordering diff. Sort "+
				"interfaces_attributes by identifier (ascending) to match Foreman's behavior. See "+
				"https://github.com/terraform-coop/terraform-provider-foreman/issues/126.",
		)
	}
}

// firstOutOfOrderPair returns the first adjacent pair of identifiers that
// violate ascending alphabetical order, or ("", "", false) if ids is
// already sorted.
func firstOutOfOrderPair(ids []string) (prev, next string, found bool) {
	for i := 1; i < len(ids); i++ {
		if ids[i] < ids[i-1] {
			return ids[i-1], ids[i], true
		}
	}
	return "", "", false
}
