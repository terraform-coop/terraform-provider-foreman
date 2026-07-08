# Foreman Terraform Provider

Terraform provider to interact with [Foreman](https://www.theforeman.org/)
and, partly, [Katello](https://theforeman.org/plugins/katello/).

Use the provider from the official **Terraform registry**:
[registry.terraform.io/providers/terraform-coop/foreman](https://registry.terraform.io/providers/terraform-coop/foreman/latest).

This is a fork of the project previously developed, owned, and maintained by
the SRE - Orchestration pod at Wayfair.

Resource/data-source documentation is rendered directly by the Terraform
Registry from the `docs/` directory in this repository — see the
[registry listing](https://registry.terraform.io/providers/terraform-coop/foreman/latest/docs).

**Example use-cases** of this provider are included in this repository under `./examples`.
See the examples for more information.

## Changes in 0.6.x
Starting with `v0.6.0` some (breaking) changes require an update of Terraform manifests.

* The host `build` argument was removed (`0.6.0`) and is replaced by `set_build_flag`. (`0.6.1`)
  * The reason behind this change is complex and was thoroughly discussed in https://github.com/terraform-coop/terraform-provider-foreman/discussions/125
  * Using the argument does one thing: it tells Foreman to set the `build` flag for a host. It defaults to `false`, setting it to `true` causes the host to be re-installed on next boot (network-based installation).
* The `method` argument is re-introduced as `provision_method`. It can be either `build` (network-based) or `image` (image-based).
  * Both options require different additional arguments, e.g the image to be used. See `examples/host/`.
* The host `name` argument was considered for deprecation (`0.6.0`). 
  * The `name` attribute has issues based on the "append_domain_name" setting in Foreman. It causes "inconsistent plan" errors when you give it a shortname as value, Terraform receives an FQDN back, and the `name` attribute is then used in variables in other places in your Terraform manifests.
  * As an alternative, the `shortname` argument can be used instead. It is meant for the hostname without the domain part. If you use `name` as input argument, `shortname` will be filled by the provider automatically.
  * To get the host's FQDN from the provider, use the read-only attribute `fqdn`. (`0.6.1`)
  * **Use `shortname` and `fqdn` as variables in your manifests**! Example: `other_server = foreman_host.other_server.fqdn`. This will prevent you from running into inconsistent plans.





## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 0.12.x
- [Golang](https://golang.org/doc/install) >= 1.13

Follow the setup instructions provided on the install sections of their
respective websites. Windows environments should have a \*nix-style terminal
emulator installed such as [Cygwin](https://www.cygwin.com/) to be compatible
with the `makefile`.

### Foreman Requirements

The following tools might be useful to control power of bare-metal hosts through proxy in your setup:

- [Foreman BMC Plugin](https://projects.theforeman.org/projects/smart-proxy/wiki/BMC)
- [ipmitool](https://github.com/ipmitool/ipmitool)

Foreman Smart proxies will need to be provisioned with the Foreman BMC plugin
and have the ipmitool installed.

In case you are still using an older version of Foreman with disabled organizations and locations (< 1.21), you need to disable organizations and locations in the provider by setting `organization_id` and `location_id` to a value < 0.


## Provider / Repository Setup

After installing and configuring the toolchain listed in the `Requirements`
section:

1. Clone the repository with `ssh`:

    ```sh
    $ go get -u github.com:terraform-coop/terraform-provider-foreman
    ```

2. Enter the root directory of the project and install the provider:

    ```sh
    $ export CGO_ENABLED=0
    $ go build -trimpath -ldflags '-s -w'
    ```

    **NOTE:** See the Third-party Plugins section on Terraform's website over
    [here](https://www.terraform.io/docs/configuration/providers.html#third-party-plugins)

3. Initialize Terraform and verify the provider is recognized by terraform:

    ```sh
    $ cd ./examples/verify_provider
    $ terraform init
    $ terraform --version
    ```

    You should see the `foreman` provider in the output like in the listing
    below.  Other providers may be listed if you have already configured Terraform.
    Your version info may be different depending on the version of Terraform you
    installed as part of the Requirements.

    ```
    Terraform v0.12.15
    + provider.foreman (unversioned)
    ```

    **NOTE:** Some builds of Terraform will require subdirectories underneath
    `terraform.d/plugins` organized by operating system and architecture.
    If this is the case, create the directory (if it doesn't exist) and then
    place the plugin within that directory.  If your `terraform init` failed with
    the following message `Provider "foreman" not available for installation`,
    then this is likely the case.  Read the error message and create the correct
    subdirectory.  For 64 bit Windows, this will be
    `terraform.d/plugins/windows_amd64`.  So in step 2, confirm the provider
    binary is located at `terraform.d/plugins/windows_amd64/terraform-provider-foreman.exe`
    and then try step 3 again.

## Documentation

Rendered documentation is available on the
[Terraform Registry](https://registry.terraform.io/providers/terraform-coop/foreman/latest/docs),
which renders it directly from the `docs/` directory committed to this
repository — no separate hosting or build step is required to view it.

`docs/` is generated from the provider's schema plus the real `.tf` examples
under `examples/` using [`terraform-plugin-docs`](https://github.com/hashicorp/terraform-plugin-docs)
(`tfplugindocs`). After changing a resource/data-source schema or its
example, regenerate the docs and commit the result:

```
$> make docs
```

CI (`docs` job in `.github/workflows/test.yml`) fails the build if `docs/`
is out of date, so this must be run and committed alongside any schema
change.

## Using the Go client library (goforeman)

The Foreman API client this provider is built on lives in
[`goforeman/`](./goforeman) as its own Go module,
`github.com/terraform-coop/terraform-provider-foreman/goforeman`, usable
by any Go program without pulling in the provider's terraform-plugin
dependency tree:

```go
import "github.com/terraform-coop/terraform-provider-foreman/goforeman"

client := goforeman.NewClient(serverURL,
    goforeman.WithBasicAuth("admin", "changeme"),
    goforeman.WithTaxonomy(orgID, locID),
)
host, err := client.FindHostByName(ctx, "web01.example.com")
```

It deliberately absorbs the Foreman API's sharp edges (URL namespace
routing, taxonomy placement, async task polling, polymorphic parameter
values, the `_destroy` deletion convention, and more) — see the package
documentation in [`goforeman/doc.go`](./goforeman/doc.go) for the full
list. Library releases are tagged `goforeman/vX.Y.Z` (Go's nested-module
tag format), independently of the provider's `vX.Y.Z` releases:

```
$> go get github.com/terraform-coop/terraform-provider-foreman/goforeman@goforeman/v0.1.0
```

Most of the client is regenerated from `apidoc/v2.json` by
`tools/gen/client` (same `make generate` as the provider); the
hand-written files are the allowlisted ones in `.gitignore`.

## Logging

**NOTE:** When developing, it may be useful to setup terraform logging. A full
list of Terraform environment variables can be found
[here](https://www.terraform.io/docs/configuration/environment-variables.html).
At minimum, it is advised to set the log level to `DEBUG` like so:

MacOS / Linux
```sh
$ export TF_LOG=DEBUG
```

Windows
```powershell
> $env:TF_LOG = "DEBUG"
```

The provider logs through [`tflog`](https://developer.hashicorp.com/terraform/plugin/log/writing),
the standard Terraform Plugin Framework logging library. Log output is
controlled entirely by Terraform's own `TF_LOG`/`TF_LOG_PROVIDER` environment
variables (see the link above) and goes to Terraform's normal log stream —
there is no separate provider-specific log file to configure.

## Migrating from the pre-rewrite provider

The `provider_loglevel` and `provider_logfile` provider-block arguments (and
their `FOREMAN_PROVIDER_LOGLEVEL`/`FOREMAN_PROVIDER_LOGFILE` environment
variable equivalents) from the old custom file-based logger no longer exist.
Remove them from your provider block if present — `terraform plan`/`apply`
will otherwise fail with an "Unsupported argument" error — and use `TF_LOG`
as described above instead.

The `foreman_global_parameter` resource and data source were renamed to
`foreman_commonparameter`. Update your configuration's resource/data source
type accordingly; existing state can be migrated with
[`terraform state mv`](https://developer.hashicorp.com/terraform/cli/commands/state/mv),
e.g. `terraform state mv foreman_global_parameter.example foreman_commonparameter.example`.

Every other provider-block argument (`server_hostname`, `server_protocol`,
`client_username`/`FOREMAN_CLIENT_USERNAME`,
`client_password`/`FOREMAN_CLIENT_PASSWORD`, `client_tls_insecure`,
`client_auth_negotiate`, `organization_id`, `location_id`) is unchanged.

## Known limitations

**Creating many `foreman_host` resources at once can hit a Foreman-side race
condition.** ([#192](https://github.com/terraform-coop/terraform-provider-foreman/issues/192))
Under Terraform's default parallelism, some hosts in a large batch may fail
to create with no logged error, and a subsequent `terraform apply` then
fails with `Name has already been taken` for those hosts even though they
don't appear in the Foreman web UI or `hammer` — the host row was partially
created before something in Foreman's own request handling (most likely
contention in its orchestration providers: DHCP/DNS/TFTP record creation)
failed. `POST /api/hosts` is a plain synchronous call with no async task to
wait on (confirmed against `apidoc/v2.json` and this provider's client code),
so this isn't a case of the provider returning before Foreman has actually
finished — the race is on Foreman's side under concurrent host creation. If
you hit this, apply with a lower parallelism for the hosts in question, e.g.
`terraform apply -parallelism=1`, or `-parallelism=<N>` tuned to what your
Foreman instance can handle concurrently.
