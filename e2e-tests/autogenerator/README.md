# Terraform provider Foreman parser

This Go program imports the Foreman providers code base and enables users to fetch schemas for both resources and data sources.
For each, the attributes are printed out to have a better overview.

## Motivation

The code was created in context of end-to-end tests for the Foreman provider. It serves as the base tool for an autogenerator of Terraform files that can then be used to test all APIs in Foreman and Katello.

In the end, this tool should create a bunch of \*.tf files with random data, which can then be aimed at a Foreman E2E test instance via `terraform apply`.

