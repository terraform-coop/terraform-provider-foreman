//go:build integration

package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/terraform-coop/terraform-provider-foreman/goforeman"
)

func testClient(t *testing.T) *goforeman.Client {
	t.Helper()

	if os.Getenv("FOREMAN_INTEGRATION_TESTS") == "" {
		t.Skip("Set FOREMAN_INTEGRATION_TESTS=1 to run integration tests")
	}

	hostname := os.Getenv("FOREMAN_SERVER_HOSTNAME")
	if hostname == "" {
		hostname = "localhost"
	}
	username := os.Getenv("FOREMAN_CLIENT_USERNAME")
	if username == "" {
		username = "admin"
	}
	password := os.Getenv("FOREMAN_CLIENT_PASSWORD")
	if password == "" {
		password = "changeme"
	}
	tlsInsecure := os.Getenv("FOREMAN_CLIENT_TLS_INSECURE") == "true"

	scheme := "https"
	if os.Getenv("FOREMAN_CLIENT_SCHEME") == "http" {
		scheme = "http"
	}

	serverURL := url.URL{
		Scheme: scheme,
		Host:   hostname,
	}

	opts := []goforeman.Option{goforeman.WithBasicAuth(username, password)}
	if tlsInsecure {
		opts = append(opts, goforeman.WithTLSInsecure())
	}
	return goforeman.NewClient(serverURL, opts...)
}

func uniqueName(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

// ptr returns a pointer to v, for constructing request structs whose
// optional int64 fields use pointer semantics (nil vs. explicit value) -
// see issue #185.
func ptr[T any](v T) *T {
	return &v
}

func TestIntegration_Domain(t *testing.T) {
	t.Parallel()

	c := testClient(t)
	ctx := context.Background()
	name := uniqueName("intg-domain") + ".test"

	domain, err := c.CreateDomain(ctx, &goforeman.DomainRequest{
		Name:     name,
		Fullname: "Integration Test Domain",
	})
	if err != nil {
		t.Fatalf("create domain: %v", err)
	}
	t.Cleanup(func() {
		_ = c.DeleteDomain(context.Background(), domain.ID)
	})

	if domain.Name != name {
		t.Errorf("domain name: got %q, want %q", domain.Name, name)
	}
	if domain.Fullname != "Integration Test Domain" {
		t.Errorf("domain fullname: got %q, want %q", domain.Fullname, "Integration Test Domain")
	}

	read, err := c.ReadDomain(ctx, domain.ID)
	if err != nil {
		t.Fatalf("read domain: %v", err)
	}
	if read.ID != domain.ID {
		t.Errorf("read domain ID: got %d, want %d", read.ID, domain.ID)
	}

	updated, err := c.UpdateDomain(ctx, domain.ID, &goforeman.DomainRequest{
		Fullname: "Updated Integration Domain",
	})
	if err != nil {
		t.Fatalf("update domain: %v", err)
	}
	if updated.Fullname != "Updated Integration Domain" {
		t.Errorf("updated fullname: got %q, want %q", updated.Fullname, "Updated Integration Domain")
	}

	read2, err := c.ReadDomain(ctx, domain.ID)
	if err != nil {
		t.Fatalf("read updated domain: %v", err)
	}
	if read2.Fullname != "Updated Integration Domain" {
		t.Errorf("read2 fullname: got %q, want %q", read2.Fullname, "Updated Integration Domain")
	}

	if err := c.DeleteDomain(ctx, domain.ID); err != nil {
		t.Fatalf("delete domain: %v", err)
	}

	_, err = c.ReadDomain(ctx, domain.ID)
	if !errors.Is(err, goforeman.ErrNotFound) {
		t.Errorf("expected 404 after delete, got: %v", err)
	}
}

func TestIntegration_Architecture(t *testing.T) {
	t.Parallel()

	c := testClient(t)
	ctx := context.Background()
	name := uniqueName("intg-arch")

	arch, err := c.CreateArchitecture(ctx, &goforeman.ArchitectureRequest{
		Name: name,
	})
	if err != nil {
		t.Fatalf("create architecture: %v", err)
	}
	t.Cleanup(func() {
		_ = c.DeleteArchitecture(context.Background(), arch.ID)
	})

	if arch.Name != name {
		t.Errorf("arch name: got %q, want %q", arch.Name, name)
	}

	read, err := c.ReadArchitecture(ctx, arch.ID)
	if err != nil {
		t.Fatalf("read architecture: %v", err)
	}
	if read.ID != arch.ID {
		t.Errorf("read arch ID: got %d, want %d", read.ID, arch.ID)
	}

	newName := uniqueName("intg-arch-upd")
	updated, err := c.UpdateArchitecture(ctx, arch.ID, &goforeman.ArchitectureRequest{
		Name: newName,
	})
	if err != nil {
		t.Fatalf("update architecture: %v", err)
	}
	if updated.Name != newName {
		t.Errorf("updated name: got %q, want %q", updated.Name, newName)
	}

	read2, err := c.ReadArchitecture(ctx, arch.ID)
	if err != nil {
		t.Fatalf("read updated architecture: %v", err)
	}
	if read2.Name != newName {
		t.Errorf("read2 name: got %q, want %q", read2.Name, newName)
	}

	if err := c.DeleteArchitecture(ctx, arch.ID); err != nil {
		t.Fatalf("delete architecture: %v", err)
	}

	_, err = c.ReadArchitecture(ctx, arch.ID)
	if !errors.Is(err, goforeman.ErrNotFound) {
		t.Errorf("expected 404 after delete, got: %v", err)
	}
}

func TestIntegration_OperatingSystem(t *testing.T) {
	t.Parallel()

	c := testClient(t)
	ctx := context.Background()
	name := uniqueName("intg-os")
	desc := "Integration Test OS"

	osObj, err := c.CreateOperatingSystem(ctx, &goforeman.OperatingSystemRequest{
		Name:        name,
		Major:       "1",
		Description: desc,
	})
	if err != nil {
		t.Fatalf("create operating system: %v", err)
	}
	t.Cleanup(func() {
		_ = c.DeleteOperatingSystem(context.Background(), osObj.ID)
	})

	if osObj.Name != name {
		t.Errorf("os name: got %q, want %q", osObj.Name, name)
	}
	if osObj.Major != "1" {
		t.Errorf("os major: got %q, want %q", osObj.Major, "1")
	}
	if osObj.Description != desc {
		t.Errorf("os description: got %q, want %q", osObj.Description, desc)
	}

	read, err := c.ReadOperatingSystem(ctx, osObj.ID)
	if err != nil {
		t.Fatalf("read operating system: %v", err)
	}
	if read.ID != osObj.ID {
		t.Errorf("read os ID: got %d, want %d", read.ID, osObj.ID)
	}

	newDesc := "Updated Integration OS"
	updated, err := c.UpdateOperatingSystem(ctx, osObj.ID, &goforeman.OperatingSystemRequest{
		Description: newDesc,
	})
	if err != nil {
		t.Fatalf("update operating system: %v", err)
	}
	if updated.Description != newDesc {
		t.Errorf("updated description: got %q, want %q", updated.Description, newDesc)
	}

	read2, err := c.ReadOperatingSystem(ctx, osObj.ID)
	if err != nil {
		t.Fatalf("read updated operating system: %v", err)
	}
	if read2.Description != newDesc {
		t.Errorf("read2 description: got %q, want %q", read2.Description, newDesc)
	}

	if err := c.DeleteOperatingSystem(ctx, osObj.ID); err != nil {
		t.Fatalf("delete operating system: %v", err)
	}

	_, err = c.ReadOperatingSystem(ctx, osObj.ID)
	if !errors.Is(err, goforeman.ErrNotFound) {
		t.Errorf("expected 404 after delete, got: %v", err)
	}
}

func TestIntegration_Hostgroup(t *testing.T) {
	t.Parallel()

	c := testClient(t)
	ctx := context.Background()

	arch, err := c.CreateArchitecture(ctx, &goforeman.ArchitectureRequest{
		Name: uniqueName("intg-hg-arch"),
	})
	if err != nil {
		t.Fatalf("create prerequisite architecture: %v", err)
	}
	t.Cleanup(func() {
		_ = c.DeleteArchitecture(context.Background(), arch.ID)
	})

	domain, err := c.CreateDomain(ctx, &goforeman.DomainRequest{
		Name: uniqueName("intg-hg-domain") + ".test",
	})
	if err != nil {
		t.Fatalf("create prerequisite domain: %v", err)
	}
	t.Cleanup(func() {
		_ = c.DeleteDomain(context.Background(), domain.ID)
	})

	osObj, err := c.CreateOperatingSystem(ctx, &goforeman.OperatingSystemRequest{
		Name:  uniqueName("intg-hg-os"),
		Major: "1",
	})
	if err != nil {
		t.Fatalf("create prerequisite operating system: %v", err)
	}
	t.Cleanup(func() {
		_ = c.DeleteOperatingSystem(context.Background(), osObj.ID)
	})

	name := uniqueName("intg-hg")
	hg, err := c.CreateHostgroup(ctx, &goforeman.HostgroupRequest{
		Name:              name,
		ArchitectureID:    ptr(int64(arch.ID)),
		DomainID:          ptr(int64(domain.ID)),
		OperatingsystemID: ptr(int64(osObj.ID)),
	})
	if err != nil {
		t.Fatalf("create hostgroup: %v", err)
	}
	t.Cleanup(func() {
		_ = c.DeleteHostgroup(context.Background(), hg.ID)
	})

	if hg.Name != name {
		t.Errorf("hostgroup name: got %q, want %q", hg.Name, name)
	}
	if hg.ArchitectureID != int64(arch.ID) {
		t.Errorf("hostgroup architecture_id: got %d, want %d", hg.ArchitectureID, arch.ID)
	}
	if hg.DomainID != int64(domain.ID) {
		t.Errorf("hostgroup domain_id: got %d, want %d", hg.DomainID, domain.ID)
	}
	if hg.OperatingsystemID != int64(osObj.ID) {
		t.Errorf("hostgroup operatingsystem_id: got %d, want %d", hg.OperatingsystemID, osObj.ID)
	}

	read, err := c.ReadHostgroup(ctx, hg.ID)
	if err != nil {
		t.Fatalf("read hostgroup: %v", err)
	}
	if read.ID != hg.ID {
		t.Errorf("read hostgroup ID: got %d, want %d", read.ID, hg.ID)
	}

	// Terraform's Update always resends the resource's complete desired
	// state (not a partial patch), so a real update call re-sends every
	// attribute the plan still has set, not just the one that changed.
	newDesc := "Updated Integration Hostgroup"
	updated, err := c.UpdateHostgroup(ctx, hg.ID, &goforeman.HostgroupRequest{
		Description:       newDesc,
		ArchitectureID:    ptr(int64(arch.ID)),
		DomainID:          ptr(int64(domain.ID)),
		OperatingsystemID: ptr(int64(osObj.ID)),
	})
	if err != nil {
		t.Fatalf("update hostgroup: %v", err)
	}
	if updated.Description != newDesc {
		t.Errorf("updated description: got %q, want %q", updated.Description, newDesc)
	}

	read2, err := c.ReadHostgroup(ctx, hg.ID)
	if err != nil {
		t.Fatalf("read updated hostgroup: %v", err)
	}
	if read2.Description != newDesc {
		t.Errorf("read2 description: got %q, want %q", read2.Description, newDesc)
	}
	if read2.DomainID != int64(domain.ID) {
		t.Errorf("read2 domain_id: got %d, want %d (should be unchanged)", read2.DomainID, domain.ID)
	}

	// Issue #185: omitting an optional FK field from the update request
	// (nil pointer, explicit JSON null) must actually clear it server-side,
	// not silently leave the previous association in place.
	cleared, err := c.UpdateHostgroup(ctx, hg.ID, &goforeman.HostgroupRequest{
		Description:       newDesc,
		ArchitectureID:    ptr(int64(arch.ID)),
		OperatingsystemID: ptr(int64(osObj.ID)),
		// DomainID intentionally omitted (nil) to clear it.
	})
	if err != nil {
		t.Fatalf("update hostgroup clearing domain_id: %v", err)
	}
	if cleared.DomainID != 0 {
		t.Errorf("cleared domain_id: got %d, want 0", cleared.DomainID)
	}

	read3, err := c.ReadHostgroup(ctx, hg.ID)
	if err != nil {
		t.Fatalf("read hostgroup after clearing domain_id: %v", err)
	}
	if read3.DomainID != 0 {
		t.Errorf("read3 domain_id: got %d, want 0 (should have been cleared)", read3.DomainID)
	}

	if err := c.DeleteHostgroup(ctx, hg.ID); err != nil {
		t.Fatalf("delete hostgroup: %v", err)
	}

	_, err = c.ReadHostgroup(ctx, hg.ID)
	if !errors.Is(err, goforeman.ErrNotFound) {
		t.Errorf("expected 404 after delete, got: %v", err)
	}
}

func TestIntegration_Host(t *testing.T) {
	t.Parallel()

	c := testClient(t)
	ctx := context.Background()

	arch, err := c.CreateArchitecture(ctx, &goforeman.ArchitectureRequest{
		Name: uniqueName("intg-host-arch"),
	})
	if err != nil {
		t.Fatalf("create prerequisite architecture: %v", err)
	}
	t.Cleanup(func() {
		_ = c.DeleteArchitecture(context.Background(), arch.ID)
	})

	domain, err := c.CreateDomain(ctx, &goforeman.DomainRequest{
		Name: uniqueName("intg-host-domain") + ".test",
	})
	if err != nil {
		t.Fatalf("create prerequisite domain: %v", err)
	}
	t.Cleanup(func() {
		_ = c.DeleteDomain(context.Background(), domain.ID)
	})

	osObj, err := c.CreateOperatingSystem(ctx, &goforeman.OperatingSystemRequest{
		Name:  uniqueName("intg-host-os"),
		Major: "1",
	})
	if err != nil {
		t.Fatalf("create prerequisite operating system: %v", err)
	}
	t.Cleanup(func() {
		_ = c.DeleteOperatingSystem(context.Background(), osObj.ID)
	})

	hg, err := c.CreateHostgroup(ctx, &goforeman.HostgroupRequest{
		Name:              uniqueName("intg-host-hg"),
		ArchitectureID:    ptr(int64(arch.ID)),
		DomainID:          ptr(int64(domain.ID)),
		OperatingsystemID: ptr(int64(osObj.ID)),
	})
	if err != nil {
		t.Fatalf("create prerequisite hostgroup: %v", err)
	}
	t.Cleanup(func() {
		_ = c.DeleteHostgroup(context.Background(), hg.ID)
	})

	hostName := uniqueName("intg-host")

	// Use Post directly so managed=false and build=false are explicit in the
	// JSON payload (Go's omitempty would otherwise strip them).
	hostReq := map[string]interface{}{
		"name":               hostName,
		"managed":            false,
		"build":              false,
		"architecture_id":    arch.ID,
		"domain_id":          domain.ID,
		"operatingsystem_id": osObj.ID,
		"hostgroup_id":       hg.ID,
	}

	var host goforeman.Host
	if err := c.Post(ctx, "hosts", "host", hostReq, &host); err != nil {
		t.Fatalf("create host: %v", err)
	}
	t.Cleanup(func() {
		_ = c.DeleteHost(context.Background(), host.ID)
	})

	if !strings.Contains(host.Name, hostName) {
		t.Errorf("host name: got %q, want it to contain %q", host.Name, hostName)
	}
	if host.Managed {
		t.Errorf("host managed: got true, want false")
	}
	if host.Build {
		t.Errorf("host build: got true, want false")
	}
	if host.ArchitectureID != int64(arch.ID) {
		t.Errorf("host architecture_id: got %d, want %d", host.ArchitectureID, arch.ID)
	}
	if host.DomainID != int64(domain.ID) {
		t.Errorf("host domain_id: got %d, want %d", host.DomainID, domain.ID)
	}
	if host.OperatingsystemID != int64(osObj.ID) {
		t.Errorf("host operatingsystem_id: got %d, want %d", host.OperatingsystemID, osObj.ID)
	}
	if host.HostgroupID != int64(hg.ID) {
		t.Errorf("host hostgroup_id: got %d, want %d", host.HostgroupID, hg.ID)
	}

	read, err := c.ReadHost(ctx, host.ID)
	if err != nil {
		t.Fatalf("read host: %v", err)
	}
	if read.ID != host.ID {
		t.Errorf("read host ID: got %d, want %d", read.ID, host.ID)
	}

	newComment := "updated by integration test"
	updated, err := c.UpdateHost(ctx, host.ID, &goforeman.HostRequest{
		Comment: newComment,
	})
	if err != nil {
		t.Fatalf("update host: %v", err)
	}
	if updated.Comment != newComment {
		t.Errorf("updated comment: got %q, want %q", updated.Comment, newComment)
	}

	read2, err := c.ReadHost(ctx, host.ID)
	if err != nil {
		t.Fatalf("read updated host: %v", err)
	}
	if read2.Comment != newComment {
		t.Errorf("read2 comment: got %q, want %q", read2.Comment, newComment)
	}

	if err := c.DeleteHost(ctx, host.ID); err != nil {
		t.Fatalf("delete host: %v", err)
	}

	_, err = c.ReadHost(ctx, host.ID)
	if !errors.Is(err, goforeman.ErrNotFound) {
		t.Errorf("expected 404 after delete, got: %v", err)
	}
}

// TestIntegration_EndToEnd builds a complete Foreman environment the way a
// real deployment would be wired together, top to bottom: architecture and
// domain, a subnet with full network information, the provisioning building
// blocks (partition table, installation medium, provisioning template with
// its template kind), an operating system tied to all of them plus its
// default provision template, a fully-wired hostgroup with group
// parameters, domain-scoped and global parameters, and finally a host in
// that hostgroup with multiple interfaces and host parameters. Then it
// exercises the read side (find-by-name lookups, list) and the update
// paths - including removing a host interface, which live-validates the
// Rails _destroy convention (goforeman.AppendDestroyMarkers) against a
// real server. Cleanup is t.Cleanup LIFO, i.e. exact reverse dependency
// order.
func TestIntegration_EndToEnd(t *testing.T) {
	t.Parallel()

	c := testClient(t)
	ctx := context.Background()

	// --- Foundation: architecture + domain ---

	arch, err := c.CreateArchitecture(ctx, &goforeman.ArchitectureRequest{
		Name: uniqueName("e2e-arch"),
	})
	if err != nil {
		t.Fatalf("create architecture: %v", err)
	}
	t.Cleanup(func() { _ = c.DeleteArchitecture(context.Background(), arch.ID) })

	domainName := uniqueName("e2e") + ".example.test"
	domain, err := c.CreateDomain(ctx, &goforeman.DomainRequest{
		Name:     domainName,
		Fullname: "E2E domain " + domainName,
	})
	if err != nil {
		t.Fatalf("create domain: %v", err)
	}
	t.Cleanup(func() { _ = c.DeleteDomain(context.Background(), domain.ID) })

	// --- Network: subnet with full network information ---

	subnet, err := c.CreateSubnet(ctx, &goforeman.SubnetRequest{
		Name:        uniqueName("e2e-subnet"),
		Network:     "198.51.100.0", // TEST-NET-2
		Mask:        "255.255.255.0",
		Gateway:     "198.51.100.1",
		DNSPrimary:  "198.51.100.2",
		From:        "198.51.100.10",
		To:          "198.51.100.100",
		BootMode:    "Static",
		IPam:        "Internal DB",
		NetworkType: "IPv4",
		DomainIDs:   []int64{int64(domain.ID)},
	})
	if err != nil {
		t.Fatalf("create subnet: %v", err)
	}
	// DeleteSubnet itself clears the domain association first - Foreman
	// otherwise deadlocks subnet<->domain deletion in both directions.
	t.Cleanup(func() { _ = c.DeleteSubnet(context.Background(), subnet.ID) })

	if got := subnet.Gateway; got != "198.51.100.1" {
		t.Errorf("subnet gateway: got %q, want %q", got, "198.51.100.1")
	}

	// --- Provisioning building blocks: ptable, medium, template ---

	ptable, err := c.CreatePartitionTable(ctx, &goforeman.PartitionTableRequest{
		Name:     uniqueName("e2e-ptable"),
		OsFamily: "Redhat",
		Layout:   "zerombr\nclearpart --all --initlabel\nautopart\n",
	})
	if err != nil {
		t.Fatalf("create partition table: %v", err)
	}
	t.Cleanup(func() { _ = c.DeletePartitionTable(context.Background(), ptable.ID) })

	medium, err := c.CreateMedium(ctx, &goforeman.MediumRequest{
		Name:     uniqueName("e2e-medium"),
		Path:     "http://mirror.example.test/pub/os/$version/$arch",
		OsFamily: "Redhat",
	})
	if err != nil {
		t.Fatalf("create medium: %v", err)
	}
	t.Cleanup(func() { _ = c.DeleteMedium(context.Background(), medium.ID) })

	// The "script" kind: creating an OS auto-creates os_default_templates
	// for every stock template whose family matches (PXELinux, provision,
	// finish, ...), so those kinds are "already taken" - script has no
	// stock Redhat default and stays free for our template.
	scriptKind, err := c.FindTemplateKindByName(ctx, "script")
	if err != nil || scriptKind == nil {
		t.Fatalf("find 'script' template kind: result=%v err=%v", scriptKind, err)
	}

	template, err := c.CreateProvisioningTemplate(ctx, &goforeman.ProvisioningTemplateRequest{
		Name:           uniqueName("e2e-template"),
		Template:       "#!/bin/bash\n# end-to-end test provision template\necho provisioned\n",
		Snippet:        ptr(false),
		TemplateKindID: ptr(int64(scriptKind.ID)),
	})
	if err != nil {
		t.Fatalf("create provisioning template: %v", err)
	}
	t.Cleanup(func() { _ = c.DeleteProvisioningTemplate(context.Background(), template.ID) })

	// --- Operating system, wired to everything above ---

	osObj, err := c.CreateOperatingSystem(ctx, &goforeman.OperatingSystemRequest{
		Name:                    uniqueName("e2eOS"),
		Major:                   "9",
		Minor:                   "3",
		Family:                  "Redhat",
		Description:             "End-to-end test OS",
		ArchitectureIDs:         []int64{int64(arch.ID)},
		MediumIDs:               []int64{int64(medium.ID)},
		PtableIDs:               []int64{int64(ptable.ID)},
		ProvisioningTemplateIDs: []int64{int64(template.ID)},
	})
	if err != nil {
		t.Fatalf("create operating system: %v", err)
	}
	t.Cleanup(func() { _ = c.DeleteOperatingSystem(context.Background(), osObj.ID) })

	if len(osObj.ArchitectureIDs) != 1 || osObj.ArchitectureIDs[0] != int64(arch.ID) {
		t.Errorf("os architecture_ids: got %v, want [%d]", osObj.ArchitectureIDs, arch.ID)
	}
	// Foreman auto-associates every stock template whose family matches, so
	// ours is one of several - assert membership, not exact contents.
	foundTpl := false
	for _, id := range osObj.ProvisioningTemplateIDs {
		if id == int64(template.ID) {
			foundTpl = true
			break
		}
	}
	if !foundTpl {
		t.Errorf("os provisioning_template_ids: %v does not contain our template %d", osObj.ProvisioningTemplateIDs, template.ID)
	}

	// Make the template the OS's default for the script kind
	// (parent-scoped os_default_templates endpoint).
	defaultTpl, err := c.CreateDefaultTemplate(ctx, osObj.ID, &goforeman.DefaultTemplateRequest{
		ProvisioningTemplateID: ptr(int64(template.ID)),
		TemplateKindID:         ptr(int64(scriptKind.ID)),
	})
	if err != nil {
		t.Fatalf("create os default template: %v", err)
	}
	t.Cleanup(func() { _ = c.DeleteDefaultTemplate(context.Background(), osObj.ID, defaultTpl.ID) })

	// --- Hostgroup wiring the whole stack, with group parameters ---

	hgName := uniqueName("e2e-hg")
	hg, err := c.CreateHostgroup(ctx, &goforeman.HostgroupRequest{
		Name:              hgName,
		Description:       "End-to-end test hostgroup",
		ArchitectureID:    ptr(int64(arch.ID)),
		DomainID:          ptr(int64(domain.ID)),
		SubnetID:          ptr(int64(subnet.ID)),
		OperatingsystemID: ptr(int64(osObj.ID)),
		MediumID:          ptr(int64(medium.ID)),
		PtableID:          ptr(int64(ptable.ID)),
		RootPass:          "e2e-secret-rootpw",
		GroupParametersAttributes: goforeman.BuildParameters(map[string]string{
			"e2e_tier": "integration",
		}),
	})
	if err != nil {
		t.Fatalf("create hostgroup: %v", err)
	}
	t.Cleanup(func() { _ = c.DeleteHostgroup(context.Background(), hg.ID) })

	hgParams, err := goforeman.ParseParameters(hg.Parameters)
	if err != nil {
		t.Fatalf("parse hostgroup parameters: %v", err)
	}
	if hgParams["e2e_tier"] != "integration" {
		t.Errorf("hostgroup parameter e2e_tier: got %q, want %q", hgParams["e2e_tier"], "integration")
	}

	// --- Parameters: domain-scoped and global ---

	domainParam, err := c.CreateParameter(ctx, "domains", domain.ID, &goforeman.ParameterRequest{
		Name:          "e2e_dns_search",
		Value:         domainName,
		ParameterType: "string",
	})
	if err != nil {
		t.Fatalf("create domain parameter: %v", err)
	}
	t.Cleanup(func() { _ = c.DeleteParameter(context.Background(), "domains", domain.ID, domainParam.ID) })

	commonParam, err := c.CreateCommonParameter(ctx, &goforeman.CommonParameterRequest{
		Name:          uniqueName("e2e_global"),
		Value:         "true",
		ParameterType: "boolean",
	})
	if err != nil {
		t.Fatalf("create common parameter: %v", err)
	}
	t.Cleanup(func() { _ = c.DeleteCommonParameter(context.Background(), commonParam.ID) })

	if got := goforeman.RawValueString(commonParam.Value); got != "true" {
		t.Errorf("common parameter value: got %q, want %q", got, "true")
	}

	// --- Host in the hostgroup, two interfaces, host parameters ---

	hostName := uniqueName("e2e-host")
	host, err := c.CreateHost(ctx, &goforeman.HostRequest{
		Name:        hostName,
		HostgroupID: ptr(int64(hg.ID)),
		Managed:     ptr(false),
		Build:       ptr(false),
		Comment:     "end-to-end test host",
		HostParametersAttributes: goforeman.BuildParameters(map[string]string{
			"e2e_role": "worker",
		}),
		InterfacesAttributes: []map[string]interface{}{
			{
				"identifier": "eth0",
				"mac":        "52:54:00:e2:e0:01",
				"ip":         "198.51.100.20",
				"subnet_id":  subnet.ID,
				"domain_id":  domain.ID,
				"primary":    true,
				"managed":    false,
			},
			{
				"identifier": "eth1",
				"mac":        "52:54:00:e2:e0:02",
				"managed":    false,
			},
		},
	})
	if err != nil {
		t.Fatalf("create host: %v", err)
	}
	t.Cleanup(func() { _ = c.DeleteHost(context.Background(), host.ID) })

	if !strings.Contains(host.Name, hostName) {
		t.Errorf("host name: got %q, want it to contain %q", host.Name, hostName)
	}
	if !strings.Contains(host.Name, domainName) {
		t.Errorf("host fqdn: got %q, want the hostgroup's domain %q appended", host.Name, domainName)
	}
	if host.HostgroupID != int64(hg.ID) {
		t.Errorf("host hostgroup_id: got %d, want %d", host.HostgroupID, hg.ID)
	}

	hostParams, err := goforeman.ParseParameters(host.Parameters)
	if err != nil {
		t.Fatalf("parse host parameters: %v", err)
	}
	if hostParams["e2e_role"] != "worker" {
		t.Errorf("host parameter e2e_role: got %q, want %q", hostParams["e2e_role"], "worker")
	}

	readIfaces := func() []map[string]interface{} {
		read, err := c.ReadHost(ctx, host.ID)
		if err != nil {
			t.Fatalf("read host: %v", err)
		}
		var ifaces []map[string]interface{}
		if err := json.Unmarshal(read.Interfaces, &ifaces); err != nil {
			t.Fatalf("parse host interfaces: %v", err)
		}
		return ifaces
	}

	ifaces := readIfaces()
	if len(ifaces) != 2 {
		t.Fatalf("host interfaces after create: got %d, want 2 (%v)", len(ifaces), ifaces)
	}

	// --- Read side: find-by-name lookups and list ---

	if found, err := c.FindHostgroupByName(ctx, hgName); err != nil || found == nil || found.ID != hg.ID {
		t.Errorf("FindHostgroupByName(%q): got %v, err %v", hgName, found, err)
	}
	if found, err := c.FindSubnetByName(ctx, subnet.Name); err != nil || found == nil || found.ID != subnet.ID {
		t.Errorf("FindSubnetByName(%q): got %v, err %v", subnet.Name, found, err)
	}
	if found, err := c.FindParameterByName(ctx, "domains", domain.ID, "e2e_dns_search"); err != nil || found == nil || found.ID != domainParam.ID {
		t.Errorf("FindParameterByName(domain %d, e2e_dns_search): got %v, err %v", domain.ID, found, err)
	}
	hosts, err := c.ListHosts(ctx)
	if err != nil {
		t.Fatalf("list hosts: %v", err)
	}
	foundHost := false
	for _, h := range hosts {
		if h.ID == host.ID {
			foundHost = true
			break
		}
	}
	if !foundHost {
		t.Errorf("ListHosts: created host %d not in the %d returned", host.ID, len(hosts))
	}

	// --- Updates ---

	// Remove eth1: live-validates the Rails accepts_nested_attributes_for
	// _destroy convention - simply omitting the interface would be silently
	// ignored by Foreman; AppendDestroyMarkers adds the explicit marker.
	var keep []map[string]interface{}
	for _, iface := range ifaces {
		if iface["identifier"] == "eth0" {
			keep = append(keep, map[string]interface{}{
				"id":         int64(iface["id"].(float64)),
				"identifier": "eth0",
			})
		}
	}
	if len(keep) != 1 {
		t.Fatalf("expected to find eth0 among interfaces, got %v", ifaces)
	}
	prior := make([]map[string]interface{}, 0, len(ifaces))
	for _, iface := range ifaces {
		prior = append(prior, map[string]interface{}{"id": int64(iface["id"].(float64))})
	}

	newComment := "updated by end-to-end test"
	updated, err := c.UpdateHost(ctx, host.ID, &goforeman.HostRequest{
		Comment:              newComment,
		HostgroupID:          ptr(int64(hg.ID)), // nil pointer FKs mean "clear" - re-send what must survive
		InterfacesAttributes: goforeman.AppendDestroyMarkers(keep, prior),
	})
	if err != nil {
		t.Fatalf("update host: %v", err)
	}
	if updated.Comment != newComment {
		t.Errorf("updated host comment: got %q, want %q", updated.Comment, newComment)
	}

	ifaces = readIfaces()
	if len(ifaces) != 1 {
		t.Errorf("host interfaces after removing eth1: got %d, want 1 (%v)", len(ifaces), ifaces)
	}

	// Subnet network-info update.
	updatedSubnet, err := c.UpdateSubnet(ctx, subnet.ID, &goforeman.SubnetRequest{
		Gateway: "198.51.100.254",
	})
	if err != nil {
		t.Fatalf("update subnet: %v", err)
	}
	if got := updatedSubnet.Gateway; got != "198.51.100.254" {
		t.Errorf("updated subnet gateway: got %q, want %q", got, "198.51.100.254")
	}
}
