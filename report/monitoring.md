# Monitoring decisions and graphs

## 1. Kubernetes Metrics Scraping & Visualization Flow

```mermaid
flowchart LR
  subgraph K8s[Kubernetes Cluster]
    subgraph NSAPP[t-hub-dev namespace]
      S1[team-service metrics]:::svc
      S2[user-service metrics]:::svc
      S3[tournament-service metrics]:::svc
      S4[bracket-service metrics]:::svc
      GW[api-gateway]:::svc
    end

    subgraph NSMON[monitoring namespace]
      P[Prometheus scrape jobs]:::mon
      G[Grafana Dashboards + Alerts]:::mon
    end
  end

  P -->|scrape /metrics| S1
  P -->|scrape /metrics| S2
  P -->|scrape /metrics| S3
  P -->|scrape /metrics| S4

  G -->|PromQL queries| P
  U[(Developer / Student)] -->|view dashboards| G

  classDef svc fill:#1f2937,stroke:#93c5fd,color:#e5e7eb
  classDef mon fill:#111827,stroke:#34d399,color:#e5e7eb
```

## 2. HTTP Request → Metrics Collection Flow

```mermaid
flowchart TB
  R["Incoming HTTP Request"]
  R --> M["metrics middleware (in_flight, duration, requests_total)"]
  M --> H["Route Handler / business logic"]
  H --> Resp["HTTP Response"]
  M -->|updates metrics| REG["Prometheus client registry"]
  REG -->|exposed at /metrics| METRICS["/metrics endpoint"]
```
## 3. Prometheus and Grafana
We used prometheus to scrape the services and grafana for some simple graphs and alerts. To access prometheus directly one needs to portforward to the service. But we decided to expose grafana with ingress. We know that this is not best practice but having a stable access point made development easier. The dashboards can be found at k8s/infra-charts/monitoring/dashboards/ and the alerts at k8s/infra-charts/monitoring/templates/prometheus-rules-configmap.yaml

## 4. No logging or tracing tools used
Some basic logging is provided by rancher/argocd and that is the only thing we used when working on this project.
