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
