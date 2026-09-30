# Polyteia DB Connector 🚀

Der Polyteia DB Connector ist ein Tool, das Daten aus SQL-Datenbanken (PostgreSQL, MySQL oder MS SQL Server) extrahiert,
die Ergebnisse in das CSV-Format umwandelt und über eine API in ein Polyteia-Dataset hochlädt. Es eignet sich ideal für
geplante, automatisierte Datenübertragungen aus Ihren internen Datenbanken auf die Polyteia-Plattform.

---

## ✨ Funktionen

- **Datenbankunterstützung:** Verbindet sich mit PostgreSQL-, MySQL- und MS SQL Server-Datenbanken über native Go
  SQL-Treiber für effizientes Abfragen und CSV-Export.
- **Automatisierte Planung:** Führt Jobs nach einem konfigurierbaren Cron-Zeitplan aus.
- **Polyteia-Integration:** Lädt Daten sicher mit Personal Access Keys in Polyteia-Datasets hoch.
- **Flexibles Deployment:** Lokal, als Docker-Container oder in Kubernetes (Helm-Chart inbegriffen) ausführbar.
- **Startprüfungen:** Prüft beim Start Datenbankzugriff, Zugangsdaten und Dataset-Berechtigungen und protokolliert das
  Ergebnis jeder Prüfung.
- **Health Checks:** Stellt einen Endpunkt unter `/healthz` für Liveness- und Readiness-Probes bereit.
- **Konfigurierbare Protokollierung:** Unterstützt verschiedene Log-Level und -Formate.
- **Sichere Konfiguration:** Umgebungsvariablen und Secret-Injection für sensible Daten.

---

## 🏗️ Architektur & Workflow

1. **Konfiguration:** Der Connector lädt seine Konfiguration aus Umgebungsvariablen oder einer `.env`-Datei.
2. **Startprüfungen:** Bevor ein Job geplant wird, prüft der Connector, ob Quell-Datenbank, Polyteia-API und
   Ziel-Dataset erreichbar und nutzbar sind. Schlägt eine Prüfung fehl, beendet er sich (siehe
   [Startprüfungen](#-startprüfungen)).
3. **Datenbankverbindung:** Native Go SQL-Treiber werden verwendet, um sich direkt mit der Quell-Datenbank (
   PostgreSQL/MySQL/MS SQL Server) zu verbinden und eine benutzerdefinierte SQL-Abfrage auszuführen.
4. **Datenexport:** Das Abfrageergebnis wird als CSV-Datei in ein temporäres Verzeichnis exportiert, mit effizientem
   Speichermanagement für große Datensätze.
5. **Authentifizierung:** Der Connector tauscht seinen Personal Access Key gegen eine kurzlebige Sitzung für die
   konfigurierte Organisation.
6. **Upload:** Die CSV-Datei wird in das angegebene Polyteia-Dataset hochgeladen und ersetzt dessen Daten. Anschließend
   wartet der Connector, bis der Import abgeschlossen ist, und markiert den Job als fehlgeschlagen, wenn der Import
   fehlschlägt.
7. **Zeitplanung:** Der Prozess wird nach einem Cron-Zeitplan ausgelöst, mit Wiederholungsversuchen bei Fehlern.
8. **Health Check:** Ein leichtgewichtiger HTTP-Server stellt den Endpunkt `/healthz` für die Überwachung bereit.

---

## ⚙️ Konfiguration

Die gesamte Konfiguration erfolgt über Umgebungsvariablen (diese können in einer `.env`-Datei oder als Secrets in
Kubernetes gesetzt werden). Nachfolgend eine Liste der unterstützten Variablen:

| Variable                    | Beschreibung                                                            | Erforderlich | Standardwert                   |
|-----------------------------|-------------------------------------------------------------------------|--------------|--------------------------------|
| `PERSONAL_ACCESS_TOKEN`     | Polyteia Personal Access Key (`pak_...`).                               | Ja           | –                              |
| `POLYTEIA_BASE_URL`         | Basis-URL für die Polyteia-API.                                         | Nein         | <https://web.polyteia.de>      |
| `POLYTEIA_ORGANIZATION_ID`  | ID der Organisation, zu der das Dataset gehört (`org_...`).             | Ja\*         | –                              |
| `POLYTEIA_ORGANIZATION_SLUG`| Slug der Organisation, zu der das Dataset gehört.                       | Ja\*         | –                              |
| `DATASET_ID`                | ID des Ziel-Polyteia-Datasets.                                          | Ja           | –                              |
| `INGEST_POLL_INTERVAL`      | Wie oft der Import-Status nach einem Upload abgefragt wird.             | Nein         | 5s                             |
| `INGEST_TIMEOUT`            | Wie lange nach einem Upload auf den Abschluss des Imports gewartet wird. | Nein        | 30m                            |
| `CRON_SCHEDULE`             | Cron-Ausdruck für die Job-Planung.                                      | Nein         | 0 0 ** * (Mitternacht täglich) |
| `LOG_LEVEL`                 | Log-Level: debug, info, warn, error.                                    | Nein         | info                           |
| `LOG_FORMAT`                | Log-Format: text oder json.                                             | Nein         | text                           |
| `HEALTH_CHECK_PORT`         | Port für den Health-Check-Server.                                       | Nein         | 8080                           |
| `SOURCE_DATABASE_HOST`      | Hostname der Quell-Datenbank.                                           | Ja           | –                              |
| `SOURCE_DATABASE_PORT`      | Port der Quell-Datenbank.                                               | Ja           | –                              |
| `SOURCE_DATABASE_USER`      | Benutzername für die Quell-Datenbank.                                   | Ja           | –                              |
| `SOURCE_DATABASE_PASSWORD`  | Passwort für die Quell-Datenbank.                                       | Nein         | –                              |
| `SOURCE_DATABASE_NAME`      | Name der Quell-Datenbank.                                               | Ja           | –                              |
| `SOURCE_DATABASE_TYPE`      | Typ der Quell-Datenbank: `postgres`, `mysql`, `mssql` oder `sqlserver`. | Ja           | –                              |
| `SOURCE_DATABASE_SQL_QUERY` | SQL-Abfrage, die auf der Quell-Datenbank ausgeführt wird.               | Ja           | –                              |

\* Genau eine der Variablen `POLYTEIA_ORGANIZATION_ID` oder `POLYTEIA_ORGANIZATION_SLUG` muss gesetzt sein.

Der Benutzer des Personal Access Keys muss lizenziertes Mitglied der Organisation sein, das Dataset bearbeiten dürfen und
die Auftragsverarbeitungsbedingungen der Lösung des Datasets bestätigt haben. Die Organisation muss Personal Access Keys
erlauben.

---

> [!TIP]
> Unter [Polyteia Docs](https://docs.polyteia.com/konto/zugriffsschlussel-pak) finden Sie Informationen, wie Sie
> Personal Access Keys erstellen.

## 📝 Beispiel `.env`-Datei

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
> Die SQL-Abfrage verwendet die Standard-SQL-Syntax für den ausgewählten Datenbanktyp. Da der Connector sich direkt mit
> der Datenbank über native Treiber verbindet, können Sie Standard-Abfragen ohne spezielle Präfixe verwenden. Zum
> Beispiel:
>
> - PostgreSQL: `SELECT * FROM my_table;` oder `SELECT * FROM schema.my_table;`
> - MySQL: `SELECT * FROM my_table;` oder `SELECT * FROM database.my_table;`
> - MS SQL Server: `SELECT * FROM dbo.my_table;` oder `SELECT * FROM schema.my_table;`>

---

## 🚀 Verwendung

### 🖥️ Lokal (Go)

1. Kopieren Sie die Beispiel-`.env`-Datei und füllen Sie Ihre Konfiguration aus.
2. Führen Sie den Connector aus:

```bash
go run ./cmd/connector
```

### 🐳 Docker

Erstellen Sie das Docker-Image (oder nutzen Sie ein veröffentlichtes):

```bash
docker build -t polyteia-db-connector .
```

Starten Sie den Container mit Ihrer `.env`-Datei:

```bash
docker run --env-file .env polyteia-db-connector:latest
```

Alternativ können Sie die Umgebungsvariablen direkt übergeben:

```bash
docker run -e PERSONAL_ACCESS_TOKEN=... -e POLYTEIA_ORGANIZATION_SLUG=... -e DATASET_ID=... ... polyteia-db-connector:latest
```

### ☸️ Kubernetes (Helm)

Ein Helm-Chart wird unter `charts/polyteia-db-connector` bereitgestellt.

1. Passen Sie [values.yaml](./charts/polyteia-db-connector/values.yaml) für Ihre Umgebung und Secrets an.
2. Deployment mit Helm:

```bash
helm upgrade --install polyteia-db-connector charts/polyteia-db-connector
```

- Umgebungsvariablen können über `env` oder `envFrom` in `values.yaml` gesetzt werden.
- Ressourcenanforderungen/-limits und Netzwerkrichtlinien sind konfigurierbar.

---

## ✅ Startprüfungen

Beim Start führt der Connector die folgenden Prüfungen aus, bevor ein Job geplant wird, und protokolliert, wohin er sich
verbunden hat und das Ergebnis jeder Prüfung, egal ob erfolgreich oder fehlgeschlagen:

| Prüfung                 | Was geprüft wird                                                                                   |
|-------------------------|----------------------------------------------------------------------------------------------------|
| `source database`       | Die Quell-Datenbank ist erreichbar und akzeptiert die konfigurierten Zugangsdaten.                 |
| `source query`          | Die Datenbank kann `SOURCE_DATABASE_SQL_QUERY` parsen und die verwendeten Tabellen existieren. Die Abfrage wird nicht ausgeführt. Für MS SQL Server übersprungen. |
| `personal access key`   | Der Personal Access Key ist gültig und für die konfigurierte Organisation nutzbar.                 |
| `dataset access`        | Das Dataset existiert und der Benutzer des Keys darf es sehen.                                     |
| `upload permission`     | Der Benutzer des Keys darf das Dataset bearbeiten, was für Uploads nötig ist.                      |
| `data processing terms` | Der Benutzer des Keys hat die Auftragsverarbeitungsbedingungen der Lösung des Datasets bestätigt.  |

Alle Prüfungen laufen, auch wenn eine fehlschlägt; Prüfungen, die von einer fehlgeschlagenen abhängen, werden mit einer
Warnung übersprungen. Eine fehlgeschlagene Prüfung wird mit dem Grund und einem Hinweis auf die zu prüfende Einstellung
protokolliert. Schlägt eine Prüfung fehl, beendet sich der Connector mit Exit-Code 1, sodass Fehlkonfigurationen sofort
auffallen (z. B. als Pod im Crash-Loop in Kubernetes) und nicht erst beim ersten Job.

Beispielausgabe:

```text
level=INFO msg="Startup check passed" check="source database" type=postgres host=localhost port=5432 database=app user=connector
level=INFO msg="Startup check passed" check="source query"
level=INFO msg="Startup check passed" check="personal access key" base_url=https://web.polyteia.de organization_id=org_... scope=organization
level=INFO msg="Startup check passed" check="dataset access" dataset_id=ds_... dataset_name="Mein Dataset" solution_id=sol_...
level=INFO msg="Startup check passed" check="upload permission" dataset_id=ds_...
level=ERROR msg="Startup check failed" check="data processing terms" error="the user of the personal access key has not acknowledged the data processing terms of solution sol_..., ..."
level=ERROR msg="Connector is not ready, fix the failed checks above and restart" error="startup checks failed: [data processing terms]"
```

---

## ❤️ Health Check

Der Connector stellt einen Health-Check-Endpunkt unter `http://<host>:<HEALTH_CHECK_PORT>/healthz` für Liveness- und
Readiness-Probes bereit.

---

## 🏷️ Versionierung

Dieses Projekt folgt [Semantic Versioning](https://semver.org/). Releases werden auf GitHub und als Docker-Images
veröffentlicht. Schauen Sie auf der [Releases-Seite](https://github.com/polyteia-connect/polyteia-db-connector/releases)
nach der aktuellen Version und dem Changelog.

---

## 🤝 Mitwirken

Beiträge, Issues und Feature-Anfragen sind herzlich willkommen! Bitte nutzen
Sie [GitHub Issues](https://github.com/polyteia-connect/polyteia-db-connector/issues/new/choose), um Fehler zu melden
oder Verbesserungen vorzuschlagen.

### Wie Sie mitwirken können 🛠️

Wir freuen uns über Ihren Input! Um mitzuwirken, folgen Sie bitte diesen Schritten:

1. **Klonen** Sie das Repository lokal (falls noch nicht geschehen):

   ```bash
   git clone https://github.com/polyteia-connect/polyteia-db-connector.git
   cd polyteia-db-connector
   ```

2. **Erstellen Sie einen neuen Branch** (direkt im Repository) für Ihr Feature oder Bugfix:

   ```bash
   git checkout -b my-feature-branch
   ```

3. **Nehmen Sie Ihre Änderungen vor** und committen Sie diese mit aussagekräftigen Nachrichten.
4. **Pushen** Sie Ihren Branch ins Repository:

   ```bash
   git push origin my-feature-branch
   ```

5. **Öffnen Sie einen Pull Request** gegen den `main`-Branch. Bitte beschreiben Sie Ihre Änderungen klar und verweisen
   Sie auf zugehörige Issues.
6. Warten Sie auf Review und Feedback. Wir arbeiten mit Ihnen zusammen, um Ihren PR zu mergen!

---

## 📄 Lizenz

Dieses Projekt ist unter der MIT License lizenziert.

---

## 💬 Support

Bei Fragen oder Support kontaktieren Sie bitte das Polyteia-Team oder eröffnen Sie ein Issue auf GitHub.
