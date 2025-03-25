# gcectl: Google Cloud Compute Engine Commands

[![Go](https://github.com/haru-256/gcectl/actions/workflows/go.yml/badge.svg)](https://github.com/haru-256/gcectl/actions/workflows/go.yml)
[![Rust](https://github.com/haru-256/gcectl/actions/workflows/rust.yml/badge.svg)](https://github.com/haru-256/gcectl/actions/workflows/rust.yml)

A project that combines a Go CLI tool for managing Google Cloud Compute Engine instances and Terraform configurations for infrastructure provisioning.

## Description

This repository contains:

### Go CLI Tool

A command-line utility for managing Google Cloud Compute Engine VMs with the following features:

* Listing VM instances
* Turning instances on/off
* Setting machine types
* Managing schedule policies
* Status monitoring

### Terraform Configurations

Infrastructure as Code for creating and managing GCE resources:

* VM instance provisioning
* Network configuration
* Schedule policy management
* State management setup

*## Directory Structure

```sh
.
├── .github/                # GitHub-specific configurations
│   ├── workflows/          # GitHub Actions workflows
│   └── pull_request_template.md
│
├── go/                     # Go CLI tool source code
│   ├── cmd/                # Command implementations
│   │   ├── list.go         # VM listing command
│   │   ├── off.go          # VM shutdown command
│   │   ├── on.go           # VM startup command
│   │   ├── root.go         # Root command definition
│   │   └── set/            # Commands for setting VM properties
│   ├── pkg/                # Core packages
│   │   ├── config/         # Configuration handling
│   │   ├── gce/            # GCE API interactions
│   │   ├── log/            # Logging utilities
│   │   └── utils/          # Utility functions
│   ├── main.go             # CLI entry point
│   └── gcectl.yaml   # Configuration example
│
├── terraform/              # Terraform configurations
│   ├── environments/       # Environment-specific configs
│   │   └── dev/            # Development environment
│   ├── modules/            # Reusable Terraform modules
│   │   ├── gce/            # GCE instance module
│   │   └── tfstate_gcs_bucket/ # GCS bucket for Terraform state
│   └── scripts/            # Helper scripts for Terraform
│
└── .tool-versions          # Tool version specifications
```

The Go CLI tool provides a convenient interface for managing GCE instances, while the Terraform configurations enable infrastructure provisioning and management following best practices.
*
