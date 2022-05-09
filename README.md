# Foreman Terraform Provider:

Terraform provider to interact with [Foreman](https://www.theforeman.org/).

This project is a component of Project Argo.

## Project Info

This is a fork of the project previously developed, owned, and maintained by
the SRE - Orchestration pod at Wayfair.

This repository uses [`mkdocs`](https://www.mkdocs.org/) for documentation and
Go modules for dependency management.  Dependencies are tracked as part of the
repository.

## Foreman Requirements:

- [Foreman BMC Plugin](https://projects.theforeman.org/projects/smart-proxy/wiki/BMC)
- [ipmitool](https://github.com/ipmitool/ipmitool)

Foreman Smart proxies will need to be provisioned with the Foreman BMC plugin
and have the ipmitool installed.

Currently supported Foreman versions are all >= 1.16 and <= 1.20. Above 1.20
the API was changed with some new required parameters which are not yet
implemented in the provider.

## Requirements:

- [Terraform](https://www.terraform.io/downloads.html) >= 0.12.x
- [Golang](https://golang.org/doc/install) >= 1.13

Follow the setup instructions provided on the install sections of their
respective websites. Windows environments should have a \*nix-style terminal
emulator installed such as [Cygwin](https://www.cygwin.com/) to be compatible
with the `makefile`.

## Provider / Repository Setup

After installing and configuring the toolchain listed in the `Requirements`
section:

1. Clone the repository with `ssh`:

    ```sh
    $ go get -u github.com:HanseMerkur/terraform-provider-foreman
    ```

2. Enter the root directory of the project and install the provider:

    ```sh
    $ go build
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

## [Documentation](https://hansemerkur.github.io/terraform-provider-foreman/)

This repository uses [`mkdocs`](https://www.mkdocs.org/) for documentation.
Follow the installation instructions on
[`mkdocs`](https://www.mkdocs.org/#installation) to get started or use the
auto-generated documentation available on the Github Pages for this project.

The `mkdocs` configuration and associated markdown is auto-generated for the
provider using the `autodoc` package from the utility repository. The
`autodoc` tool uses text templates defined in `templates` and the schema
definitions in the provider to generate all the necessary `mkdocs` files and
resources. The `autodoc` command is located in `cmd/autodoc/main.go`.

To generate and view the entire repository and in-depth provider documentation:

```
$> go build -v -o autodoc $(go list ./cmd/autodoc)
$> mkdir -p docs/{datasources,resources}
$> ./autodoc
$> mkdocs serve
INFO    -  Building documentation...
INFO    -  Cleaning site directory
[I 160402 15:50:43 server:271] Serving on http://127.0.0.1:8000
[I 160402 15:50:43 handlers:58] Start watching changes
[I 160402 15:50:43 handlers:60] Start detecting changes
```
If you see an error `'jinja2' has no attribute 'contextfilter'`, try pinning jinja2 to v3.0.3: `pip install jinja2==3.0.3`

The documentation can then be viewed by accessing localhost in your favorite
browser or viewport.

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

The provider is set to log to the file `terraform-provider-foreman.log` with
all Foreman provider specific log messages sent to this file.  When the
provider is executed, it will create the provider log file in the current
working directory (if it does not exist).  If the log file already exists,
then the logs are *appended* to the existing file.  In the case the
provider cannot create/open the desired log file, the provider defaults to
sending log messages to `stderr`.

The provider uses a level-based logging module that extends the golang
stdlib `log` package.  When the log level is set to a verbosity threshold,
only log messages of that verbosity and higher are sent to the output file.

From most verbose to least verbose:

| Log Level | Description |
| :--- | :--- |
| DEBUG | Intermediate calculations, values. Useful when debugging. |
| TRACE | Function enter/exit notifications |
| INFO | Notifications - not related to suspicious behavior or errors |
| WARNING | Suspcious or error behavior, but the system was able to recover or default/degrade gracefully |
| ERROR | Behavior that causes the program execution to stop |
| NONE | Do not log any output |

The provider's log level defaults to `INFO`, meaning `INFO`, `WARNING`, and
`ERROR` messages are committed to the log file, `DEBUG` and `TRACE` are
ignored.  The log level can be overridden by either setting the
`provider_loglevel` attribute in the provider block of the Terraform module,
or by setting the environment variable `FOREMAN_PROVIDER_LOGLEVEL`.  If both
values are set, `provider_loglevel` takes precedence. You can also override
the Foreman provider's log file using the `FOREMAN_PROVIDER_LOGFILE`
environment variable. A value of `-` preserves the stdlib `log` behavior
and outputs to the `stdlog` stream.

Ex:

Terraform module
```
provider "foreman" {
  ...
  provider_loglevel = "DEBUG"
  provider_logfile  = "terraform-provider-foreman.log"
  ...
}
```

MacOS / Linux
```shell
$> export FOREMAN_PROVIDER_LOGLEVEL="DEBUG"
$> export FOREMAN_PROVIDER_LOGFILE="terraform-provider-foreman.log"
```

Windows
```powershell
> $env:FOREMAN_PROVIDER_LOGLEVEL = "DEBUG"
> $env:FOREMAN_PROVIDER_LOGFILE = "terraform-provider-foreman.log"
```

## Using the Provider:

An example of of usage of this provider is included in this repository under
`./examples`. See the examples for more information.
