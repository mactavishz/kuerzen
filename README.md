# Kuerzen

A simple distributed URL shortener that scales.

## Development

### Prerequisites

- Install [Golang](https://go.dev/dl/) >= 1.23.4
- Install [Docker](https://docs.docker.com/get-docker/)
- Install [Docker Compose](https://docs.docker.com/compose/install/)
- Install [Go dotenv](https://github.com/joho/godotenv) as a command line tool

### Setup

Create a `.env` file in the root of the project and add the following variables, you can use the `.env.example` file as a reference.

- ports:
  - `SHORTENER_PORT` - The port for the shortener service.
  - `REDIRECT_PORT` - The port for the redirect service.
  - `ANALYTICS_PORT` - The port for the analytics service.
- postgres:
  - `POSTGRES_USER` - The username for the postgres database.
  - `POSTGRES_PASSWORD` - The password for the postgres database.
  - `POSTGRES_DB` - The database name for the postgres database.
- influxdb:
  - `DOCKER_INFLUXDB_INIT_MODE` - The mode for the influxdb database.
  - `DOCKER_INFLUXDB_INIT_USERNAME` - The username for the influxdb database.
  - `DOCKER_INFLUXDB_INIT_PASSWORD` - The password for the influxdb database.
  - `DOCKER_INFLUXDB_INIT_ORG` - The organization for the influxdb database.
  - `DOCKER_INFLUXDB_INIT_BUCKET` - The bucket for the influxdb database.
  - `DOCKER_INFLUXDB_INIT_ADMIN_TOKEN` - The admin token for the influxdb database.
- shortener:
  - `SHORTENER_DB_URL` - The database connection url for the shortener service

### Run the services

First, start the database services:

```bash
docker compose up -d
```

Then, start the services:

```bash
godotenv -f .env go run ./cmd/shortener/main.go
godotenv -f .env go run ./cmd/redirect/main.go
godotenv -f .env go run ./cmd/analytics/main.go
```
