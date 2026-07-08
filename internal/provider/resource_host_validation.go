package provider

import (
	"context"

	"github.com/terraform-coop/terraform-provider-foreman/goforeman"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Warns at plan/apply time when interfaces_attributes isn't in the
// ascending-by-identifier order Foreman will impose server-side (see
// goforeman.FirstOutOfOrderIdentifier for the underlying Foreman behavior
// and its PXE-boot consequences). Warning instead of a type-level fix
// (unordered Set) is deliberate: the SDKv2 provider tried the Set approach
// and reverted it - it broke idempotency instead of fixing it. See
// https://github.com/terraform-coop/terraform-provider-foreman/issues/126
// and pull/203.
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

	if prev, next, outOfOrder := goforeman.FirstOutOfOrderIdentifier(identifiers); outOfOrder {
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
