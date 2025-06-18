# Kuerzen

A simple distributed URL shortener that scales.

## Development

### Prerequisites

- Install [Golang](https://go.dev/dl/) >= 1.23.4
- Install [Docker](https://docs.docker.com/get-docker/)
- Install [Docker Compose](https://docs.docker.com/compose/install/)

### Setup

#### Go Workspace

Since there are some dependencies between the services, it's best to create a Go workspace, run the following command in the root of the project:

```bash
go work init
go work use ./store
go work use ./shortener
go work use ./redirector
go work use ./analytics
```

Because go workspace is only for local development, the `go.work` file is not included in the repository.

### Environment Variables

Create a `.env` file in the root of the project and add the following variables, you can use the `.env.example` file as a reference.

- ports:
  - `SHORTENER_PORT` - The port for the shortener service.
  - `REDIRECTOR_PORT` - The port for the redirect service.
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
- miscellaneous:
  - `KUERZEN_DB_URL` - The database connection url for the shortener service
  - `KUERZEN_HOST` - The host for the application

### Start the services

Run the following command to start all the services:

```bash
docker compose up -d
```

It might take a while to start the services, as the services are dependent on each other. After the services are started, you can access the API Gateway at `http[s]://localhost`.

All the go services are using hot reload using [Air](https://github.com/air-verse/air), so you can change the code and the changes will be reflected immediately.

### End the services

Run the following command to stop and remove all running services:

```bash
docker compose down
```

### API Endpoints

#### Health Check

```bash
curl -X GET http://localhost/health
```

#### URL Shortening

```bash
curl -X POST http://localhost/create \
-H "Content-Type: application/json" \
-d '{"url": "https://www.google.com"}'
```

#### URL Redirecting

```bash
curl -X GET http://localhost/[shorten_id]
```

#### Analytics

```bash
curl -X GET http://localhost/events
```
