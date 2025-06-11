# Kuerzen

A simple distributed URL shortener that scales.

## Development

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
