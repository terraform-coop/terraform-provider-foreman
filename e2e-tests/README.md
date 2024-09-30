# Terraform provider Foreman E2E tests

This folder contains end-to-end tests for the Terraform provider "foreman".

The aim is to enable developers to bootstrap their own Foreman/Katello test instance and then create
a bunch of Terraform resources in it. Based on the implemented Terraform resources in the provider,
everything can be tested, including the data sources.

Under `tf/` there are some examples available. These can be extended in the future and include or replace
the existing Terraform examples from the `examples` folder.


## Setting up a test instance of Foreman and Katello on Alma Linux 9

Currently a setup script for Alma Linux 9 is provided in this repo. Run this on your test machine to install
Foreman and Katello with the login admin/admine2e. Please refer to the script `bootstrap-script-on-host-alma9.sh` for more details.

The reference test machine used for this instance is a 16 vCPU, 32 GB RAM and 360 GB disk VM.
Installation requires around 12 minutes to complete (from logging into an empty server to running the bootstrap script to completion).

Using a machine with lower resources will block the Katello installation scenario, but might be sufficient for Foreman-only.

Using Ubuntu 22.04 also works, but only with Foreman. For Katello you need an OS from the RHEL family (as is Alma Linux).


## Auto generation of resources and data sources

This undertaking is WIP and shall allow to generate \*.tf files for all resources automatically.
See the code in folder "autogenerator".
