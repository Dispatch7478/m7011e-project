## How does the security model look like? What is the request flow for secure connections?

Keycloak is a component we use to manage authentication. We rely on keycloak to ensure that each of our users data is safely handled. Keycloak also allow access to multiple security features such as multi factor authentication or ensuring passwords are more complex that could be used to further improve user security. We also utilize keycloak groups for RBAC, where different roles are allowed to make different changes, for example admin and observer roles. The generated JWT from keycloak is used to create bearer token and periodically refresh the JWT if it is about to expire to ensure stability. 

The api gateway provide some security as it acts as the single entrypoint for all client requests, meaning it verifies the JWT token and authenticates each user and their request to ensure the user has the authorization to make those requests before acting on them in the services. 

Each service we have can also gather information from the token to see if the user is allowed to make specific changes. 

## Request flow for secure connections:
	

### User login and keycloak:   

As the user enters the site they need to authenticate themselves with keycloak in order to continue. When successful the frontend receives a JWT token from keycloak allowing the user to continue

### API Gateway 

Once the user is logged in they can start to do requests to the backend, those requests go through the api gateway which checks with keycloak if the JWT is valid (from an authenticated user), if valid a header is created from the token.

### Microservices

The gateway sends a message with the headers to the microservice that the request points to. The microservice then uses the “headers” to check more fine grained permissions, such as if the user is allowed to delete a team etc.


## Threat Protection/Mitigation

### Protection Against SQL Injection
The system is protected against SQL Injection attacks through the use of **Parameterized Queries** in all Go microservices.
* **Implementation:** We use the `pgx` driver for PostgreSQL. Instead of concatenating user input directly into SQL strings (which is vulnerable), we use placeholder values (e.g., `$1`, `$2`).
* **Effect:** The database treats user input strictly as data, never as executable code, effectively neutralizing SQL injection attempts.

### Protection Against Cross-Site Scripting (XSS)
* **Frontend Defense:** The platform is built on **Vue.js**, which automatically escapes all data bindings by default. This prevents malicious scripts injected into user profiles or team names from executing in other users' browsers.
* **Backend Defense:** The API Gateway enforces strict Content-Type headers (`application/json`), preventing browsers from interpreting API responses as executable scripts.

## Secure Communication & Certificate Management

### HTTPS and TLS Termination
All data in transit between the client and the cluster is encrypted using HTTPS.
* **TLS Termination:** Secure connections are terminated at the **Ingress Controller** (Traefik). Traffic inside the cluster (between microservices) occurs over HTTP to optimize performance, but is isolated within the private Kubernetes network.

### Certificate Management
We implement automated certificate management to ensure continuous security without manual intervention.
* **Tooling:** We use **cert-manager** in Kubernetes combined with the **Let's Encrypt** Certificate Authority.
* **Automatic Renewal:**  `cert-manager` monitors the expiration date of our TLS certificates. When a certificate is close to expiring (e.g., within 30 days), it automatically requests a new one from Let's Encrypt and updates the Kubernetes Secret used by the Ingress Controller.