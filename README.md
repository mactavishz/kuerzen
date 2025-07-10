# Kuerzen

A simple distributed URL shortener that scales with comprehensive monitoring and observability.

## Features

- **Distributed Architecture**: Microservices design with API Gateway, URL shortener, redirector, and analytics services
- **High Performance**: Built with Go and Fiber for fast HTTP handling
- **Observability**: Complete monitoring stack with Grafana, Prometheus, Loki, and Alloy
- **Analytics**: Real-time analytics with InfluxDB for URL creation and redirect events
- **Load Shedding**: Automatic load shedding based on CPU and memory thresholds
- **Health Checks**: Comprehensive health monitoring for all services
- **Local Development Support**: Full containerization with hot-reload for development

## Development

### Prerequisites

- Install [Golang](https://go.dev/dl/) >= 1.23.4
- Install [Docker](https://docs.docker.com/get-docker/)
- Install [Docker Compose](https://docs.docker.com/compose/install/)
- Install [Protocol Buffer Compiler](https://grpc.io/docs/protoc-installation/)
- Install [Make](https://www.gnu.org/software/make/) or [GNU Make](https://www.gnu.org/software/make/)

### Setup

#### Go Workspace

Since there are some dependencies between the services, it's best to create a Go workspace, run the following command in the root of the project:

```bash
go work init
go work use ./store
go work use ./middleware
go work use ./shortener
go work use ./redirector
go work use ./analytics
go work use ./retries
```

Because go workspace is only for local development, the `go.work` file is not included in the repository.

### Environment Variables

- Create a `.env` file in the root of the project, you can use the `.env.example` file as a reference.
- Create a `.app.env` file in the root of the project, you can use the `.app.env.example` file as a reference.
- Create a `.monitoring.env` file in the `monitoring` directory, you can use the `.monitoring.env.example` file as a reference.

#### What are the differences between `.env`, `.app.env`, and `.monitoring.env`?

- `.env`: Contains environment variables for substitution in the Docker Compose files.
- `.app.env`: Contains environment variables for the application services (shortener, redirector, analytics, databases).
- `.monitoring.env`: Contains environment variables for the monitoring stack (Grafana, Prometheus, Loki, Alloy).

#### Caveats

Make sure you set the following environment variables the same in both `.env` and `.app.env` files:

- `SHORTENER_PORT`
- `REDIRECTOR_PORT`
- `ANALYTICS_PORT`

### Generate Protobuf Code

Currently, only the analytics service has a protobuf file, so if you need to generate the code for the analytics service (e.g. when you update the protobuf file), run the following command in the `analytics` directory:

```bash
make gen
```

### Start the application services

Run the following command to start all the services:

```bash
docker compose up -d
```

It might take a while to start the services, as the services are dependent on each other. After the services are started, you can access the API Gateway at `http[s]://localhost`.

All the go services are using hot reload using [Air](https://github.com/air-verse/air), so you can change the code and the changes will be reflected immediately.

### Start the monitoring services

Run the following command to start the monitoring stack:

```bash
docker compose -f compose.monitoring.yml up -d
```

### Stop all the services

Run the following command to stop and remove all running services:

```bash
docker compose down
docker compose -f compose.monitoring.yml down
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

### Analytics

The analytics data is stored in InfluxDB, after the services are started, you can access the data in the InfluxDB UI at `http://localhost:8086` using the credentials defined in the `.env` file.

You can use the following query to get the data:

```flux
from(bucket: "kuerzen_analytics")
  |> range(start: -30m, stop: now())
  |> pivot(
      rowKey:["_time"],
      columnKey: ["_field"],
      valueColumn: "_value"
  )
```

### Monitoring

Please refer to the [monitoring/README.md](monitoring/README.md) for details.

## Testing

### Unit Tests

Run the following command in the root of the project to run all unit tests:

```bash
./test.sh
```

## Production Deployment

Please refer to the [production/README.md](production/README.md) for details on how to deploy the application in production in a Docker Swarm cluster.
