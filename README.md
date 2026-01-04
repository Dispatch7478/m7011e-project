# T-Hub Tournament Platform

![Go](https://img.shields.io/badge/Go-1.25-blue.svg)
![Vue.js](https://img.shields.io/badge/Vue.js-3.0-green.svg)
![Kubernetes](https://img.shields.io/badge/K8s-GitOps-326ce5.svg)
![RabbitMQ](https://img.shields.io/badge/RabbitMQ-Event--Driven-orange.svg)

T-Hub is a modern, microservices-based web application for creating and managing esports tournaments. It provides a platform for users to register, create teams, join tournaments, and compete in automatically generated brackets.

## Key Features

- **User Authentication:** Secure registration and login provided by Keycloak (OIDC).
- **High-Concurrency Registration:** Uses pessimistic database locking to handle race conditions when multiple users attempt to claim the final tournament slots simultaneously.
- **Team Management:** Users can create, manage, and invite members to their teams.
- **Tournament Creation:** Organizers can create and configure public or private tournaments.
- **Automated Bracket Generation:** Brackets are automatically created once registration is complete.
- **Observability:** Full stack monitoring with Prometheus (metrics) and Grafana (dashboards).

---

## Architecture Overview

T-Hub is built with a microservices architecture. Services communicate synchronously via REST APIs through a central API Gateway.

### Event-Driven Backbone
An asynchronous event bus (RabbitMQ) is implemented to decouple services.
* **Current Implementation:** The `Tournament Service` publishes domain events (e.g., `TournamentCreated`) to the exchange.
* **Design Intent:** This architecture ensures extensibility. Future services (e.g., Notifications, Analytics) can subscribe to these events to trigger actions without modifying the core tournament service logic.

### Service Breakdown
-   **`Frontend`**: Vue.js SPA.
-   **`API Gateway`**: Single entry point handling routing, auth validation, and rate limiting.
-   **`Tournament Service`**: Core logic for lifecycle and capacity management. Implements transaction locking for data integrity.
-   **`User Service`** & **`Team Service`**: Domain-specific management services.
-   **`Bracket Service`**: Manages match logic.

All services are containerized with Docker and deployed to a Kubernetes cluster.

## Key Technologies

- **Backend:** Go (`gorilla/mux`, `echo`)
- **Frontend:** Vue.js
- **Database:** PostgreSQL (one per service)
- **Authentication:** Keycloak
- **API Gateway:** Go
- **Deployment:** Kubernetes, Helm
- **CI/CD & GitOps:** GitHub Actions, Argo CD
- **Monitoring:** Prometheus, Grafana
- **Messaging:** RabbitMQ


## 📂 Repository Structure

```text
.
├── frontend/               # Vue.js Single Page Application
├── k8s/                    # Kubernetes Infrastructure
│   ├── argo-apps/          # ArgoCD Application manifests
│   ├── charts/             # Helm Charts for microservices
│   └── infra-charts/       # Helm Charts for dependencies (Keycloak, Postgres, RabbitMQ)
├── report/                 # Architecture, Security, and Database documentation
├── services/               # Backend Microservices (Go)
│   ├── api-gateway/        # Routing and Auth middleware
│   ├── bracket-service/    # Bracket generation logic
│   ├── team-service/       # Team management logic
│   ├── tournament-service/ # Core tournament logic & Event Publishing
│   └── user-service/       # User profile management
└── tests/                  # Load tests (k6) and integration scripts
```

---

## Getting Started / Local Development

To run the project locally, you will need to have Go, Node.js, and `kubectl` (with a configured cluster context) installed.

### Backend Services

Each backend service is a standalone Go application. To run a service, navigate to its directory and run the server. For example, to run the `tournament-service`:

```bash
cd services/tournament-service
go run .
```

You will need to do this for each backend service (`api-gateway`, `user-service`, etc.). Note that for local testing, you will likely need to use `kubectl port-forward` to connect to cluster services like the database if you are not running one locally.

### Frontend

The frontend is a Vue.js application.

1.  Navigate to the `frontend` directory:
    ```bash
    cd frontend
    ```
2.  Install dependencies:
    ```bash
    npm install
    ```
3.  Start the development server:
    ```bash
    npm run dev
    ```
The frontend will then be available at `http://localhost:5173`.

---

## Testing & Quality Assurance

### Unit Testing
Each service includes Go unit tests utilizing `pgxmock` for database mocking.

```bash
cd services/tournament-service
go test -v ./...
```

### Load testing
In the folder `/tests` there are two scripts to stress test the system with concurrent users (default 50). Running the tools requires the tool `k6`

```bash
k6 run tests/race_join_load.js
```

---

## Deployment

This project uses a GitOps workflow for deployment. All Kubernetes manifests and configurations are stored in this Git repository.

- **Argo CD** monitors the repository for changes.
- When changes are pushed to the `main` branch (e.g., updating a service's image tag in its Helm chart `values.yaml`), Argo CD automatically syncs the changes and applies them to the Kubernetes cluster, triggering a rolling update of the affected service.
- Infrastructure components like Keycloak and the monitoring stack are deployed via Helm charts located in `k8s/infra-charts/`.

## API Documentation

Each service has its own OpenAPI (Swagger) documentation located in its respective directory (e.g., `services/tournament-service/swagger.yaml`). This documentation details the available endpoints, request/response schemas, and parameters for each service.

## Monitoring

- **Prometheus:** Collects metrics from all backend services.
- **Grafana:** Provides dashboards for visualizing system metrics. The Grafana instance for this project can be found at its configured domain and requires login credentials.
