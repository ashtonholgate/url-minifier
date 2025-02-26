# Local Infrastructure Setup Guide

This guide details the process of setting up the local development infrastructure using Docker Compose, including all required services and monitoring tools.

## 1. Docker Compose Configuration

### 1.1 Project Structure
```
docker/
├── config/
│   ├── mongodb/
│   │   └── mongod.conf
│   ├── redis/
│   │   └── redis.conf
│   ├── prometheus/
│   │   └── prometheus.yml
│   ├── grafana/
│   │   ├── datasources/
│   │   │   └── automatic.yml
│   │   └── dashboards/
│   │       ├── dashboard.yml
│   │       └── url_minifier_dashboard.json
│   └── otel-collector/
│       └── config.yml
├── .env.example
└── docker-compose.yml
```

### 1.2 Base Docker Compose Configuration
Create `docker-compose.yml`:
```yaml
version: '3.8'

services:
  mongodb:
    image: mongo:6.0
    ports:
      - "27017:27017"
    volumes:
      - mongodb_data:/data/db
      - ./config/mongodb/mongod.conf:/etc/mongod.conf
    command: ["mongod", "--config", "/etc/mongod.conf"]
    healthcheck:
      test: ["CMD", "mongosh", "--eval", "db.adminCommand('ping')"]
      interval: 10s
      timeout: 5s
      retries: 3

  redis:
    image: redis:7.2
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
      - ./config/redis/redis.conf:/usr/local/etc/redis/redis.conf
    command: ["redis-server", "/usr/local/etc/redis/redis.conf"]
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 3

  unleash:
    image: unleashorg/unleash-server:5.5
    ports:
      - "4242:4242"
    environment:
      DATABASE_URL: postgresql://postgres:${POSTGRES_PASSWORD}@postgres:5432/unleash
      DATABASE_SSL: "false"
      LOG_LEVEL: debug
    depends_on:
      postgres:
        condition: service_healthy

  postgres:
    image: postgres:15
    ports:
      - "5432:5432"
    environment:
      POSTGRES_DB: unleash
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 3

  prometheus:
    image: prom/prometheus:v2.45.0
    ports:
      - "9090:9090"
    volumes:
      - ./config/prometheus/prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/usr/share/prometheus/console_libraries'
      - '--web.console.templates=/usr/share/prometheus/consoles'
    healthcheck:
      test: ["CMD", "wget", "-q", "--spider", "http://localhost:9090/-/healthy"]
      interval: 10s
      timeout: 5s
      retries: 3

  grafana:
    image: grafana/grafana:10.0.0
    ports:
      - "3100:3000"
    volumes:
      - ./config/grafana/datasources:/etc/grafana/provisioning/datasources
      - ./config/grafana/dashboards:/etc/grafana/provisioning/dashboards
      - grafana_data:/var/lib/grafana
    environment:
      GF_SECURITY_ADMIN_PASSWORD: ${GRAFANA_PASSWORD}
    depends_on:
      prometheus:
        condition: service_healthy

  otel-collector:
    image: otel/opentelemetry-collector:0.90.1
    command: ["--config=/etc/otel-collector/config.yml"]
    volumes:
      - ./config/otel-collector/config.yml:/etc/otel-collector/config.yml
    ports:
      - "4317:4317"   # OTLP gRPC
      - "4318:4318"   # OTLP HTTP
      - "8888:8888"   # metrics
    depends_on:
      prometheus:
        condition: service_healthy

volumes:
  mongodb_data:
  redis_data:
  postgres_data:
  prometheus_data:
  grafana_data:
```

## 2. Service Configuration Files

### 2.1 MongoDB Configuration
Create `config/mongodb/mongod.conf`:
```yaml
storage:
  dbPath: /data/db
systemLog:
  destination: file
  path: /var/log/mongodb/mongod.log
  logAppend: true
net:
  port: 27017
  bindIp: 0.0.0.0
```

### 2.2 Redis Configuration
Create `config/redis/redis.conf`:
```
bind 0.0.0.0
port 6379
protected-mode yes
maxmemory 256mb
maxmemory-policy allkeys-lru
```

### 2.3 Prometheus Configuration
Create `config/prometheus/prometheus.yml`:
```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'url-minifier'
    static_configs:
      - targets: ['host.docker.internal:8080']

  - job_name: 'otel-collector'
    static_configs:
      - targets: ['otel-collector:8888']
```

### 2.4 Grafana Configuration
Create `config/grafana/datasources/automatic.yml`:
```yaml
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
```

### 2.5 OpenTelemetry Collector Configuration
Create `config/otel-collector/config.yml`:
```yaml
receivers:
  otlp:
    protocols:
      grpc:
      http:

processors:
  batch:

exporters:
  prometheus:
    endpoint: "0.0.0.0:8889"
  logging:
    verbosity: detailed

service:
  pipelines:
    metrics:
      receivers: [otlp]
      processors: [batch]
      exporters: [prometheus, logging]
```

## 3. Environment Variables
Create `.env.example`:
```env
POSTGRES_PASSWORD=your_secure_password_here
GRAFANA_PASSWORD=your_secure_password_here
UNLEASH_API_TOKEN=your_api_token_here
```

## 4. Setup Instructions

1. Create the directory structure:
   ```bash
   mkdir -p docker/{config/{mongodb,redis,prometheus,grafana/{datasources,dashboards},otel-collector}}
   ```

2. Copy all configuration files to their respective directories as shown in the project structure.

3. Create a `.env` file from `.env.example`:
   ```bash
   cp docker/.env.example docker/.env
   ```

4. Update the `.env` file with secure passwords and tokens.

5. Start the infrastructure:
   ```bash
   cd docker
   docker compose up -d
   ```

6. Verify services:
   ```bash
   docker compose ps
   ```

## 5. Access Points

- MongoDB: localhost:27017
- Redis: localhost:6379
- Unleash: http://localhost:4242
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3100
- OpenTelemetry Collector:
  - OTLP gRPC: localhost:4317
  - OTLP HTTP: localhost:4318
  - Metrics: localhost:8888

## 6. Health Checks

All services include health checks to ensure proper startup order and monitoring. You can check the status of all services with:
```bash
docker compose ps
```

## 7. Data Persistence

All data is persisted using named volumes:
- `mongodb_data`: MongoDB data
- `redis_data`: Redis data
- `postgres_data`: PostgreSQL data
- `prometheus_data`: Prometheus metrics
- `grafana_data`: Grafana dashboards and settings

## 8. Monitoring Setup Verification

1. Access Grafana (http://localhost:3100)
   - Default username: admin
   - Password: From GRAFANA_PASSWORD in .env

2. Verify Prometheus data source is connected
   - Go to Configuration > Data Sources
   - Check Prometheus connection status

3. Import the URL Minifier dashboard
   - The dashboard will be automatically provisioned from the configuration

## 9. Troubleshooting

Common issues and solutions:

1. Port conflicts:
   - Check if ports are already in use: `netstat -an | grep <port>`
   - Modify the port mapping in docker-compose.yml if needed

2. Service dependencies:
   - Services are configured with health checks and proper startup order
   - Use `docker compose logs <service>` to check for specific service issues

3. Data persistence:
   - If data isn't persisting, verify volume mounts: `docker volume ls`
   - Check volume permissions: `docker volume inspect <volume_name>`
