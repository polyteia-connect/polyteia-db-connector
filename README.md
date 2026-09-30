# Polyteia DB Connector 🚀

Polyteia DB Connector is a tool designed to extract data from SQL databases (PostgreSQL, MySQL, or MS SQL Server),
transform the results
into CSV format, and upload them to a Polyteia dataset via API. It is ideal for scheduled, automated data transfers
from your internal databases to the Polyteia platform.

---

## ✨ Features

- **Database Support:** Connects to PostgreSQL, MySQL, and MS SQL Server databases using native Go SQL drivers for
  efficient querying and CSV export.
- **Automated Scheduling:** Runs jobs on a configurable cron schedule.
- **Polyteia Integration:** Securely uploads data to Polyteia datasets using personal access keys.
- **Flexible Deployment:** Run locally, as a Docker container, or in Kubernetes (Helm chart included).
- **Startup Checks:** Verifies database access, credentials and dataset permissions on startup and logs the result of
  every check.
- **Health Checks:** Exposes a `/healthz` endpoint for liveness/readiness probes.
- **Configurable Logging:** Supports log level and format customization.
- **Secure Configuration:** Environment variable and secret support for sensitive data.

---

## 🏗️ Architecture & Workflow

1. **Configuration:** The connector loads its configuration from environment variables or a `.env` file.
2. **Startup Checks:** Before scheduling any job, the connector checks that the source database, the Polyteia API and
   the target dataset are reachable and usable. It exits if any check fails (see [Startup Checks](#-startup-checks)).
3. **Database Connection:** Uses native Go SQL drivers to connect directly to the source database (PostgreSQL/MySQL/MS
   SQL Server) and execute a user-defined SQL query.
4. **Data Export:** The query result is exported as a CSV file to a temporary location with efficient memory management
   for large datasets.
5. **Authentication:** The connector exchanges its personal access key for a short-lived session scoped to the
   configured organization.
6. **Upload:** The CSV file is uploaded to the specified Polyteia dataset, replacing its data. The connector then waits
   until the dataset has finished ingesting the file and fails the job if the ingest fails.
7. **Scheduling:** The process is triggered on a cron schedule, with retries on failure.
8. **Health Check:** A lightweight HTTP server exposes `/healthz` for monitoring.

---

## ⚙️ Configuration

All configuration is done via environment variables (can be set in a `.env` file or injected as secrets in Kubernetes).
Below is a list of supported variables:

| Variable                    | Description                                                                | Required | Default                    |
|-----------------------------|----------------------------------------------------------------------------|----------|----------------------------|
| `PERSONAL_ACCESS_TOKEN`     | Polyteia personal access key (`pak_...`).                                  | Yes      | -                          |
| `POLYTEIA_BASE_URL`         | Base URL for Polyteia API.                                                 | No       | <https://web.polyteia.de>  |
| `POLYTEIA_ORGANIZATION_ID`  | ID of the organization the dataset belongs to (`org_...`).                 | Yes\*    | -                          |
| `POLYTEIA_ORGANIZATION_SLUG`| Slug of the organization the dataset belongs to.                           | Yes\*    | -                          |
| `DATASET_ID`                | Target Polyteia dataset ID.                                                | Yes      | -                          |
| `INGEST_POLL_INTERVAL`      | How often to check the ingest status after an upload.                      | No       | 5s                         |
| `INGEST_TIMEOUT`            | How long to wait for the ingest to finish after an upload.                 | No       | 30m                        |
| `CRON_SCHEDULE`             | Cron expression for job scheduling.                                        | No       | 0 0 ** * (midnight daily)  |
| `LOG_LEVEL`                 | Log level: debug, info, warn, error.                                       | No       | info                       |
| `LOG_FORMAT`                | Log format: text or json.                                                  | No       | text                       |
| `HEALTH_CHECK_PORT`         | Port for health check server.                                              | No       | 8080                       |
| `SOURCE_DATABASE_HOST`      | Hostname of the source database.                                           | Yes      | -                          |
| `SOURCE_DATABASE_PORT`      | Port of the source database.                                               | Yes      | -                          |
| `SOURCE_DATABASE_USER`      | Username for the source database.                                          | Yes      | -                          |
| `SOURCE_DATABASE_PASSWORD`  | Password for the source database.                                          | No       | -                          |
| `SOURCE_DATABASE_NAME`      | Name of the source database.                                               | Yes      | -                          |
| `SOURCE_DATABASE_TYPE`      | Type of the source database: `postgres`, `mysql`, `mssql`, or `sqlserver`. | Yes      | -                          |
| `SOURCE_DATABASE_SQL_QUERY` | SQL query to execute on the source database.                               | Yes      | -                          |

\* Exactly one of `POLYTEIA_ORGANIZATION_ID` or `POLYTEIA_ORGANIZATION_SLUG` must be set.

The user owning the personal access key must be a licensed member of the organization, be allowed to edit the dataset
and have acknowledged the data processing terms of the dataset's solution. The organization must allow personal access
keys.

---

> [!TIP]
> You can checkout [Polyteia Docs](https://docs.polyteia.com/platform-docs/de/account/personal-access-keys-pak) for
> information about how to create personal access keys.

## 📝 Example `.env` File

```env
PERSONAL_ACCESS_TOKEN=pak_your_personal_access_key
POLYTEIA_BASE_URL=https://web.polyteia.de
POLYTEIA_ORGANIZATION_SLUG=your_organization_slug
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

> [!NOTE]
> The SQL query uses standard SQL syntax for the selected database type. Since the connector connects directly to the
> database using native drivers, you can use standard queries without any special prefixes. For example:
>
> - PostgreSQL: `SELECT * FROM my_table;` or `SELECT * FROM schema.my_table;`
> - MySQL: `SELECT * FROM my_table;` or `SELECT * FROM database.my_table;`
> - MS SQL Server: `SELECT * FROM dbo.my_table;` or `SELECT * FROM schema.my_table;`

---

## 🚀 Usage

### 🖥️ Local (Go)

1. Copy the example `.env` file and fill in your configuration.
2. Run the connector:

```bash
go run ./cmd/connector
```

### 🐳 Docker

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
docker run -e PERSONAL_ACCESS_TOKEN=... -e POLYTEIA_ORGANIZATION_SLUG=... -e DATASET_ID=... ... polyteia-db-connector:latest
```

### ☸️ Kubernetes (Helm)

A Helm chart is provided in `charts/polyteia-db-connector`.

1. Customize [values.yaml](./charts/polyteia-db-connector/values.yaml) for your environment and secrets.
2. Deploy with Helm:

```bash
helm upgrade --install polyteia-db-connector charts/polyteia-db-connector
```

- Environment variables can be set via `env` or `envFrom` in `values.yaml`.
- Resource requests/limits and network policies are configurable.

---

## ✅ Startup Checks

When it starts, the connector runs the following checks before scheduling any job and logs where it connected and the
result of every check, whether it passed or failed:

| Check                   | What is checked                                                                                   |
|-------------------------|---------------------------------------------------------------------------------------------------|
| `source database`       | The source database is reachable and accepts the configured credentials.                          |
| `source query`          | The database can parse `SOURCE_DATABASE_SQL_QUERY` and the tables it uses exist. The query is not run. Skipped for MS SQL Server. |
| `personal access key`   | The personal access key is valid and can be used for the configured organization.                 |
| `dataset access`        | The dataset exists and the key's user can view it.                                                |
| `upload permission`     | The key's user is allowed to edit the dataset, which uploads require.                             |
| `data processing terms` | The key's user has acknowledged the data processing terms of the dataset's solution.              |

All checks run even if one fails; checks that depend on a failed one are skipped with a warning. A failed check is
logged with the reason and a hint about which setting to look at. If any check fails, the connector exits with code 1,
so misconfigurations show up immediately (e.g. as a crash-looping pod in Kubernetes) instead of when the first job
runs.

Example output:

```text
level=INFO msg="Startup check passed" check="source database" type=postgres host=localhost port=5432 database=app user=connector
level=INFO msg="Startup check passed" check="source query"
level=INFO msg="Startup check passed" check="personal access key" base_url=https://web.polyteia.de organization_id=org_... scope=organization
level=INFO msg="Startup check passed" check="dataset access" dataset_id=ds_... dataset_name="My dataset" solution_id=sol_...
level=INFO msg="Startup check passed" check="upload permission" dataset_id=ds_...
level=ERROR msg="Startup check failed" check="data processing terms" error="the user of the personal access key has not acknowledged the data processing terms of solution sol_..., ..."
level=ERROR msg="Connector is not ready, fix the failed checks above and restart" error="startup checks failed: [data processing terms]"
```

---

## ❤️ Health Check

The connector exposes a health check endpoint at `http://<host>:<HEALTH_CHECK_PORT>/healthz` for liveness/readiness
probes.

---

## 🏷️ Versioning

This project follows [Semantic Versioning](https://semver.org/). Releases are published on GitHub and as Docker images.
Check the [releases page](https://github.com/polyteia-connect/polyteia-db-connector/releases) for the latest version and
changelog.

---

## 🤝 Contributing

Contributions, issues, and feature requests are welcome! Please
use [GitHub Issues](https://github.com/polyteia-connect/polyteia-db-connector/issues/new/choose) to report bugs or
suggest enhancements.

### How to Contribute 🛠️

We love your input! To contribute, please follow these steps:

1. **Clone** the repository locally (if you haven't already):

   ```bash
   git clone https://github.com/polyteia-connect/polyteia-db-connector.git
   cd polyteia-db-connector
   ```

2. **Create a new branch** (directly on the repository) for your feature or bugfix:

   ```bash
   git checkout -b my-feature-branch
   ```

3. **Make your changes** and commit them with clear messages.
4. **Push** your branch to the repository:

   ```bash
   git push origin my-feature-branch
   ```

5. **Open a Pull Request** against the `main` branch. Please provide a clear description of your changes and reference
   any related issues.
6. Wait for review and feedback. We'll work with you to get your PR merged!

---

## 📄 License

This project is licensed under the MIT License.

---

## 💬 Support

For questions or support, please contact the Polyteia team or open an issue on GitHub.
