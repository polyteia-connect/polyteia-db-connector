# Polyteia DB Connector

Polyteia DB Connector is a tool designed to extract data from SQL databases (PostgreSQL or MySQL), transform the results into Parquet format, and upload them to a Polyteia dataset via API. It is ideal for scheduled, automated data transfers from your internal databases to the Polyteia platform.

---

## Features

- **Database Support:** Connects to PostgreSQL and MySQL databases using DuckDB for efficient querying and Parquet export.
- **Automated Scheduling:** Runs jobs on a configurable cron schedule.
- **Polyteia Integration:** Securely uploads data to Polyteia datasets using API tokens.
- **Flexible Deployment:** Run locally, as a Docker container, or in Kubernetes (Helm chart included).
- **Health Checks:** Exposes a `/healthz` endpoint for liveness/readiness probes.
- **Configurable Logging:** Supports log level and format customization.
- **Secure Configuration:** Environment variable and secret support for sensitive data.

---

## Architecture & Workflow

1. **Configuration:** The connector loads its configuration from environment variables or a `.env` file.
2. **Database Connection:** Uses DuckDB to connect to the source database (PostgreSQL/MySQL) and execute a user-defined SQL query.
3. **Data Export:** The query result is exported as a Parquet file to a temporary location.
4. **Authentication:** The connector authenticates with the Polyteia API using a personal access token.
5. **Upload:** The Parquet file is uploaded to the specified Polyteia dataset.
6. **Scheduling:** The process is triggered on a cron schedule, with retries on failure.
7. **Health Check:** A lightweight HTTP server exposes `/healthz` for monitoring.

---

## Configuration

All configuration is done via environment variables (can be set in a `.env` file or injected as secrets in Kubernetes). Below is a list of supported variables:

| Variable                   | Description                                                                 | Required | Default                        |
|----------------------------|-----------------------------------------------------------------------------|----------|---------------------------------|
| `PERSONAL_ACCESS_TOKEN`    | Polyteia API personal access token.                                         | Yes      | -                               |
| `POLYTEIA_BASE_URL`        | Base URL for Polyteia API.                                                  | No       | https://app.polyteia.com        |
| `DATASET_ID`               | Target Polyteia dataset ID.                                                 | Yes      | -                               |
| `CRON_SCHEDULE`            | Cron expression for job scheduling.                                         | No       | 0 0 * * * (midnight daily)      |
| `LOG_LEVEL`                | Log level: debug, info, warn, error.                                        | No       | info                            |
| `LOG_FORMAT`               | Log format: text or json.                                                   | No       | text                            |
| `HEALTH_CHECK_PORT`        | Port for health check server.                                               | No       | 8080                            |
| `SOURCE_DATABASE_HOST`     | Hostname of the source database.                                            | Yes      | -                               |
| `SOURCE_DATABASE_PORT`     | Port of the source database.                                                | Yes      | -                               |
| `SOURCE_DATABASE_USER`     | Username for the source database.                                           | Yes      | -                               |
| `SOURCE_DATABASE_PASSWORD` | Password for the source database.                                           | No       | -                               |
| `SOURCE_DATABASE_NAME`     | Name of the source database.                                                | Yes      | -                               |
| `SOURCE_DATABASE_TYPE`     | Type of the source database: `postgres` or `mysql`.                         | Yes      | -                               |
| `SOURCE_DATABASE_SQL_QUERY`| SQL query to execute on the source database.                                | Yes      | -                               |

---

## Example `.env` File

```env
PERSONAL_ACCESS_TOKEN=your_polyteia_token
POLYTEIA_BASE_URL=https://app.polyteia.com
DATASET_ID=your_dataset_id
CRON_SCHEDULE=0 0 * * *
LOG_LEVEL=info
LOG_FORMAT=text
HEALTH_CHECK_PORT=8080
SOURCE_DATABASE_HOST=localhost
SOURCE_DATABASE_PORT=5432
SOURCE_DATABASE_USER=dbuser
SOURCE_DATABASE_PASSWORD=dbpassword
SOURCE_DATABASE_NAME=mydb
SOURCE_DATABASE_TYPE=postgres
SOURCE_DATABASE_SQL_QUERY=SELECT * FROM my_table;
```

---

## Usage

### Local (Go)

1. Copy the example `.env` file and fill in your configuration.
2. Run the connector:

```bash
go run ./cmd/connector
```

### Docker

Build the Docker image (or use a published one):

```bash
docker build -t polyteia-db-connector .
```

Run the container with your `.env` file:

```bash
docker run --env-file .env polyteia-db-connector:latest
```

Or pass environment variables directly:

```bash
docker run -e PERSONAL_ACCESS_TOKEN=... -e DATASET_ID=... ... polyteia-db-connector:latest
```

### Kubernetes (Helm)

A Helm chart is provided in `charts/polyteia-db-connector`.

1. Customize `values.yaml` for your environment and secrets.
2. Deploy with Helm:

```bash
helm upgrade --install polyteia-db-connector charts/polyteia-db-connector
```

- Environment variables can be set via `env` or `envFrom` in `values.yaml`.
- Resource requests/limits and network policies are configurable.

---

## Health Check

The connector exposes a health check endpoint at `http://<host>:<HEALTH_CHECK_PORT>/healthz` for liveness/readiness probes.

---

## Versioning

This project follows [Semantic Versioning](https://semver.org/). Releases are published on GitHub and as Docker images. Check the [releases page](https://github.com/polyteia-connect/polyteia-db-connector/releases) for the latest version and changelog.

---

## Contributing

Contributions, issues, and feature requests are welcome! Please use [GitHub Issues](https://github.com/polyteia-connect/polyteia-db-connector/issues/new/choose) to report bugs or suggest enhancements.

---

## License

This project is licensed under the MIT License.

---

## Support

For questions or support, please contact the Polyteia team or open an issue on GitHub.
