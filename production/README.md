# Kuerzen Production Deployment Guide

This guide covers deploying Kuerzen URL shortener to a production Docker Swarm cluster.

## Prerequisites

- A Docker Swarm cluster with at least 3 nodes (1 manager, 2 workers)
- At least 8 GB RAM available across all your swarm nodes

## Quick Start

In case you don not have a Docker Swarm cluster available, you can quickly set one up using Docker Desktop or any cloud provider.

### 1. Prepare Docker Swarm Cluster

Ensure you have Docker installed on all nodes and initialize the swarm. If you are using Docker Desktop, you can enable Swarm mode directly from the settings.

```bash
# On manager node
docker swarm init

# On worker nodes (use token from manager)
docker swarm join --token <token> <manager-ip>:2377
```

### 2. Configure Environment

```bash
# Copy and edit production environment file
cp .prod.env.example .prod.env
# Update the environment variables as needed
vim .prod.env
```

### 3. Build Images

Run the following command to build all production images:

```bash
# login to Docker Hub if needed
docker login
# Build images for production
./build-images.sh <docker-hub-username>
```

### 4. Deploy Application Stack

```bash
# Deploy the stack
DOCKER_HUB_USER=<docker-hub-username> docker stack deploy -c compose.yml kuerzen-app
```

### 5. Deploy Monitoring Stack (Optional)

```bash
# Deploy monitoring stack
docker stack deploy -c compose.monitoring.yml kuerzen-monitoring 
```

### 6. Access the Application

Assuming you have a domain pointing to your Docker Swarm cluster named `kuerzen.example.com`, you can access the application at:

```bash
# Create a shortened URL
curl -X POST -H "Content-Type: application/json" -d '{"url": "https://www.google.com"}' http://kuerzen.example.com/create

# Redirect to a shortened URL
curl http://kuerzen.example.com/<shortened-id>

# Access the Grafana dashboard
http://kuerzen.example.com:3000
```

### 7. Access the Analytics Dashboard

At the moment, the analytics dashboard is not directly accessible. You need to add a manual port mapping to the service:

```bash
docker service update \
  --publish-add published=8086,target=8086,mode=ingress \
  kuerzen-app_analytics-db
```

After this, you can access the analytics dashboard at:

```bash
http://kuerzen.example.com:8086
```

To remove the port mapping later, you can run:

```bash
docker service update \
  --publish-rm published=8086 \ 
    kuerzen-app_analytics-db
```

### Scale the services

You can scale out any service except the database services. For example:

```bash
# Scale api-gateway service to 3 replicas
docker service scale kuerzen-app_api-gateway=3

# Scale shortener service to 3 replicas
docker service scale kuerzen-app_shortener=3

# Scale analytics service to 2 replicas
docker service scale kuerzen-app_analytics=2

# Scale redirector service to 2 replicas
docker service scale kuerzen-app_redirector=2

# To scale down, just change the number of replicas, such as:
docker service scale kuerzen-app_api-gateway=1
```

## Caveats

All services communicate using `http` protocol for simplicity.
