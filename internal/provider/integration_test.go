//go:build integration

package provider

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	generated "github.com/terraform-coop/terraform-provider-foreman/goforeman"
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

	return goforeman.NewClient(
		serverURL,
		goforeman.ClientCredentials{Username: username, Password: password},
		goforeman.ClientConfig{TLSInsecure: tlsInsecure},
	)
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
	if !goforeman.IsNotFoundError(err) {
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
	if !goforeman.IsNotFoundError(err) {
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
	if !goforeman.IsNotFoundError(err) {
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
	if !goforeman.IsNotFoundError(err) {
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
	if !goforeman.IsNotFoundError(err) {
		t.Errorf("expected 404 after delete, got: %v", err)
	}
}
