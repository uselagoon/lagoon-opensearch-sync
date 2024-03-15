# Lagoon Opensearch Sync

[![Go Reference](https://pkg.go.dev/badge/github.com/uselagoon/lagoon-opensearch-sync.svg)](https://pkg.go.dev/github.com/uselagoon/lagoon-opensearch-sync)
[![Release](https://github.com/uselagoon/lagoon-opensearch-sync/actions/workflows/release.yaml/badge.svg)](https://github.com/uselagoon/lagoon-opensearch-sync/actions/workflows/release.yaml)
[![coverage](https://raw.githubusercontent.com/uselagoon/lagoon-opensearch-sync/badges/.badges/main/coverage.svg)](https://github.com/uselagoon/lagoon-opensearch-sync/actions/workflows/coverage.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/uselagoon/lagoon-opensearch-sync)](https://goreportcard.com/report/github.com/uselagoon/lagoon-opensearch-sync)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/uselagoon/lagoon-opensearch-sync/badge)](https://securityscorecards.dev/viewer/?uri=github.com/uselagoon/lagoon-opensearch-sync)
[![OpenSSF Best Practices](https://www.bestpractices.dev/projects/8659/badge)](https://www.bestpractices.dev/projects/8659)

This tool/service synchronises Opensearch with Lagoon.
This means that it sets up the required roles and permissions based on Lagoon groups and projects.

## Prerequisites

Create a Keycloak client with the `query-groups` realm management role, and client credential authorization enabled.

See, for example, the [lagoon realm export](https://github.com/uselagoon/lagoon/tree/main/services/keycloak).

## How to use

This tool is designed to run in a Kubernetes deployment in the same namespace as a lagoon-core chart.
It will eventually be rolled into the lagoon-core chart.

The deployment requires:

1. An image from this repository.

2. These environment variables:

| Name                             | Description                                                | Example                                                                     |
| ---                              | ---                                                        | ---                                                                         |
| `DEBUG`                          | Verbose logging (not required, default `false`).           | `true`                                                                      |
| `API_DB_ADDRESS`                 | Internal service name of the API DB.                       | `lagoon-core-api-db`                                                        |
| `API_DB_PASSWORD`                | Password to the API DB.                                    |                                                                             |
| `KEYCLOAK_BASE_URL`              | HTTP URL to the internal keycloak service.                 | `http://lagoon-core-keycloak:8080/`                                         |
| `OPENSEARCH_BASE_URL`            | HTTPS URL to the internal Opensearch service.              | `https://opensearch-cluster-coordinating.opensearch.svc.cluster.local:9200` |
| `OPENSEARCH_CA_CERTIFICATE`      | Opensearch CA certificate in PEM format.                   |                                                                             |
| `OPENSEARCH_DASHBOARDS_BASE_URL` | HTTP URL to the internal Dashboards service.               | `http://opensearch-dashboards.opensearch-dashboards.svc.cluster.local:5601` |
| `KEYCLOAK_CLIENT_ID`             | Client ID of `lagoon-opensearch-sync` Keycloak client.     |                                                                             |
| `KEYCLOAK_CLIENT_SECRET`         | Client secret of `lagoon-opensearch-sync` Keycloak client. |                                                                             |
| `OPENSEARCH_ADMIN_PASSWORD`      | Password for the Opensearch `admin` user.                  |                                                                             |

3. Command `/lagoon-opensearch-sync`.

## Advanced usage

This tool can be used to debug Opensearch/Lagoon integration.
For debugging commands see `/lagoon-opensearch-sync --help`.
