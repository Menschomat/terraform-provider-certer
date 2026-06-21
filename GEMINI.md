# Gemini Developer Guide - terraform-provider-certer

This document provides context, architectural guidelines, and development workflows for AI assistants (like Gemini) working on this codebase.

---

## Codebase Overview

This repository contains the Go source code for the **Certer Terraform Provider** (`terraform-provider-certer`). This provider is built using the modern [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework) (not the legacy SDKv2).

It manages certificate configurations, API keys, and retrieves certificate/key PEM data from a Certer control plane server.

### Project Layout

```
.
├── .github/
│   └── workflows/
│       └── release.yml         # GitHub Actions automated release workflow
├── client/                     # Go client library for the Certer API
│   ├── client.go               # HTTP client methods and type definitions
│   └── client_test.go          # Unit tests for the API client
├── internal/
│   └── provider/               # Terraform provider implementation
│       ├── provider.go         # Provider initialization & schema
│       ├── provider_test.go    # Metadata and basic unit tests
│       ├── certificate_resource.go    # certer_certificate resource
│       ├── api_key_resource.go        # certer_api_key resource
│       └── certificate_data_source.go # certer_certificate_data data source
├── .goreleaser.yml             # GoReleaser configuration for provider builds
├── go.mod                      # Go module definition
├── go.sum                      # Go dependency checksums
├── main.go                     # Entrypoint serving the provider binary
└── README.md                   # User documentation for the provider
```

---

## API & Data Models

The provider interacts with a Certer instance via a standard bearer-token JSON API.

### 1. Certificate Configurations
* **Resource Type**: `certer_certificate`
* **API Endpoints**:
  * `GET /api/v1/config/certificates` - List all certificate configs
  * `POST /api/v1/config/certificates` - Create a new certificate config
  * `PUT /api/v1/config/certificates/{primary}` - Update an existing config
  * `DELETE /api/v1/config/certificates/{primary}` - Remove a config

### 2. API Key Configurations
* **Resource Type**: `certer_api_key`
* **API Endpoints**:
  * `GET /api/v1/config/api_keys` - List all configured API keys
  * `POST /api/v1/config/api_keys` - Add an API key
  * `PUT /api/v1/config/api_keys` - Update an API key
  * `DELETE /api/v1/config/api_keys?token={token}` - Revoke an API key

### 3. Certificate Material (Data Source)
* **Data Source**: `certer_certificate_data`
* **API Endpoints**:
  * `GET /api/v1/certificates` - Retrieve issued certificate PEM blocks and private keys

---

## Development Workflow

### Building and Testing

1. **Run Unit Tests**:
   ```bash
   go test ./...
   ```

2. **Build the Provider**:
   ```bash
   go build -o terraform-provider-certer
   ```

### Local Testing

To test the compiled provider locally with standard Terraform:

1. Copy the built binary to the local plugin cache:
   ```bash
   mkdir -p ~/.terraform.d/plugins/gitea.dmz.k8s.menscho.space/m0space/certer/1.0.0/darwin_arm64/
   cp terraform-provider-certer ~/.terraform.d/plugins/gitea.dmz.k8s.menscho.space/m0space/certer/1.0.0/darwin_arm64/terraform-provider-certer_v1.0.0
   ```
2. Reference the provider in a test `.tf` configuration file:
   ```hcl
   terraform {
     required_providers {
       certer = {
         source  = "gitea.dmz.k8s.menscho.space/m0space/certer"
         version = "1.0.0"
       }
     }
   }
   ```
3. Initialize the test configuration:
   ```bash
   terraform init
   ```
