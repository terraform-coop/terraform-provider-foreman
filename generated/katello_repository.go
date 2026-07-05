package generated

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

type ForemanKatelloRepositoryRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Label       string `json:"label,omitempty"`
	ProductID   int    `json:"product_id,omitempty"`
	ContentType string `json:"content_type,omitempty"`
	URL         string `json:"url,omitempty"`
	GpgKeyID    int    `json:"gpg_key_id,omitempty"`
	// Unprotected/IgnoreGlobalProxy/MirrorOnSync/VerifySslOnSync are
	// pointers, not plain bool+omitempty: Go's zero value for bool is
	// false, so plain omitempty can never send an explicit "false" - these
	// flags could be enabled but never disabled again via Terraform.
	Unprotected                   *bool    `json:"unprotected"`
	ChecksumType                  string   `json:"checksum_type,omitempty"`
	IgnoreGlobalProxy             *bool    `json:"ignore_global_proxy"`
	IgnorableContent              []string `json:"ignorable_content,omitempty"`
	DownloadPolicy                string   `json:"download_policy,omitempty"`
	DownloadConcurrency           int      `json:"download_concurrency,omitempty"`
	MirrorOnSync                  *bool    `json:"mirror_on_sync"`
	MirroringPolicy               string   `json:"mirroring_policy,omitempty"`
	VerifySslOnSync               *bool    `json:"verify_ssl_on_sync"`
	UpstreamUsername              string   `json:"upstream_username,omitempty"`
	UpstreamPassword              string   `json:"upstream_password,omitempty"`
	HttpProxyPolicy               string   `json:"http_proxy_policy,omitempty"`
	HttpProxyID                   int      `json:"http_proxy_id,omitempty"`
	DebReleases                   string   `json:"deb_releases,omitempty"`
	DebComponents                 string   `json:"deb_components,omitempty"`
	DebArchitectures              string   `json:"deb_architectures,omitempty"`
	DockerUpstreamName            string   `json:"docker_upstream_name,omitempty"`
	DockerTagsWhitelist           string   `json:"docker_tags_whitelist,omitempty"`
	AnsibleCollectionRequirements string   `json:"ansible_collection_requirements,omitempty"`
}

type ForemanKatelloRepository struct {
	ForemanObject
	Description                   string   `json:"description"`
	Label                         string   `json:"label"`
	ProductID                     int      `json:"product_id"`
	ContentType                   string   `json:"content_type"`
	URL                           string   `json:"url"`
	GpgKeyID                      int      `json:"gpg_key_id"`
	Unprotected                   bool     `json:"unprotected"`
	ChecksumType                  string   `json:"checksum_type"`
	IgnoreGlobalProxy             bool     `json:"ignore_global_proxy"`
	IgnorableContent              []string `json:"ignorable_content"`
	DownloadPolicy                string   `json:"download_policy"`
	DownloadConcurrency           int      `json:"download_concurrency"`
	MirrorOnSync                  bool     `json:"mirror_on_sync"`
	MirroringPolicy               string   `json:"mirroring_policy"`
	VerifySslOnSync               bool     `json:"verify_ssl_on_sync"`
	UpstreamUsername              string   `json:"upstream_username"`
	UpstreamPassword              string   `json:"upstream_password"`
	HttpProxyPolicy               string   `json:"http_proxy_policy"`
	HttpProxyID                   int      `json:"http_proxy_id"`
	DebReleases                   string   `json:"deb_releases"`
	DebComponents                 string   `json:"deb_components"`
	DebArchitectures              string   `json:"deb_architectures"`
	DockerUpstreamName            string   `json:"docker_upstream_name"`
	DockerTagsWhitelist           string   `json:"docker_tags_whitelist"`
	AnsibleCollectionRequirements string   `json:"ansible_collection_requirements"`
}

func (c *ForemanClient) CreateForemanKatelloRepository(ctx context.Context, req *ForemanKatelloRepositoryRequest) (*ForemanKatelloRepository, error) {
	var resp ForemanKatelloRepository
	err := c.Post(ctx, "katello/repositories", "repository", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *ForemanClient) ReadForemanKatelloRepository(ctx context.Context, id int) (*ForemanKatelloRepository, error) {
	var resp ForemanKatelloRepository
	err := c.Get(ctx, fmt.Sprintf("katello/repositories/%d", id), &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *ForemanClient) UpdateForemanKatelloRepository(ctx context.Context, id int, req *ForemanKatelloRepositoryRequest) (*ForemanKatelloRepository, error) {
	var resp ForemanKatelloRepository
	err := c.Put(ctx, fmt.Sprintf("katello/repositories/%d", id), "repository", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *ForemanClient) DeleteForemanKatelloRepository(ctx context.Context, id int) error {
	return c.Delete(ctx, fmt.Sprintf("katello/repositories/%d", id))
}

func (c *ForemanClient) QueryForemanKatelloRepository(ctx context.Context, name string) (*ForemanKatelloRepository, error) {
	var response QueryResponse
	err := c.Get(ctx, fmt.Sprintf("katello/repositories?search=name=\"%s\"", url.QueryEscape(name)), &response)
	if err != nil {
		return nil, err
	}
	if len(response.Results) == 0 {
		return nil, nil
	}
	var obj ForemanKatelloRepository
	if err := json.Unmarshal(response.Results[0], &obj); err != nil {
		return nil, err
	}
	return &obj, nil
}
