# Polyteia DB Connector 🚀

Der Polyteia DB Connector ist ein Tool, das Daten aus SQL-Datenbanken (PostgreSQL, MySQL oder MS SQL Server) extrahiert,
die Ergebnisse in das CSV-Format umwandelt und über eine API in ein Polyteia-Dataset hochlädt. Es eignet sich ideal für
geplante, automatisierte Datenübertragungen aus Ihren internen Datenbanken auf die Polyteia-Plattform.

---

## ✨ Funktionen

- **Datenbankunterstützung:** Verbindet sich mit PostgreSQL-, MySQL- und MS SQL Server-Datenbanken über native Go
  SQL-Treiber für effizientes Abfragen und CSV-Export.
- **Automatisierte Planung:** Führt Jobs nach einem konfigurierbaren Cron-Zeitplan aus.
- **Polyteia-Integration:** Lädt Daten sicher in Polyteia-Datasets hoch, indem API-Token verwendet werden.
- **Flexibles Deployment:** Lokal, als Docker-Container oder in Kubernetes (Helm-Chart inbegriffen) ausführbar.
- **Health Checks:** Stellt einen Endpunkt unter `/healthz` für Liveness- und Readiness-Probes bereit.
- **Konfigurierbare Protokollierung:** Unterstützt verschiedene Log-Level und -Formate.
- **Sichere Konfiguration:** Umgebungsvariablen und Secret-Injection für sensible Daten.

---

## 🏗️ Architektur & Workflow

1. **Konfiguration:** Der Connector lädt seine Konfiguration aus Umgebungsvariablen oder einer `.env`-Datei.
2. **Datenbankverbindung:** Native Go SQL-Treiber werden verwendet, um sich direkt mit der Quell-Datenbank (
   PostgreSQL/MySQL/MS SQL Server) zu verbinden und eine benutzerdefinierte SQL-Abfrage auszuführen.
3. **Datenexport:** Das Abfrageergebnis wird als CSV-Datei in ein temporäres Verzeichnis exportiert, mit effizientem
   Speichermanagement für große Datensätze.
4. **Authentifizierung:** Der Connector authentifiziert sich über ein Personal Access Token bei der Polyteia-API.
5. **Upload:** Die CSV-Datei wird in das angegebene Polyteia-Dataset hochgeladen.
6. **Zeitplanung:** Der Prozess wird nach einem Cron-Zeitplan ausgelöst, mit Wiederholungsversuchen bei Fehlern.
7. **Health Check:** Ein leichtgewichtiger HTTP-Server stellt den Endpunkt `/healthz` für die Überwachung bereit.

---

## ⚙️ Konfiguration

Die gesamte Konfiguration erfolgt über Umgebungsvariablen (diese können in einer `.env`-Datei oder als Secrets in
Kubernetes gesetzt werden). Nachfolgend eine Liste der unterstützten Variablen:

| Variable                    | Beschreibung                                                            | Erforderlich | Standardwert                   |
|-----------------------------|-------------------------------------------------------------------------|--------------|--------------------------------|
| `PERSONAL_ACCESS_TOKEN`     | Personal Access Token für die Polyteia-API.                             | Ja           | –                              |
| `POLYTEIA_BASE_URL`         | Basis-URL für die Polyteia-API.                                         | Nein         | <https://app.polyteia.com>     |
| `DATASET_ID`                | ID des Ziel-Polyteia-Datasets.                                          | Ja           | –                              |
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

---

> [!TIP]
> Unter [Polyteia Docs](https://docs.polyteia.com/konto/zugriffsschlussel-pak) finden Sie Informationen, wie Sie
> Personal Access Keys erstellen.

## 📝 Beispiel `.env`-Datei

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
docker run -e PERSONAL_ACCESS_TOKEN=... -e DATASET_ID=... ... polyteia-db-connector:latest
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
