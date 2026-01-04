# Dynamic Web System Motivation: T-Hub

The T-Hub Tournament Platform is classified as a **dynamic web system** because it generates content in real-time based on user interactions and database state, rather than serving static HTML files.

## Key Dynamic Characteristics

*   **State Management:** Tournaments transition through complex states (e.g., `draft` → `registration_open` → `ongoing` → `completed`), which dynamically triggers different system behaviors and UI updates across the application.

*   **User-Specific Content:** The frontend dynamically renders views based on the logged-in user's identity and roles. For example, it shows "My Teams" lists, "Organizer" controls for specific users, and "Admin" pages. This is achieved by inspecting OIDC claims provided by Keycloak after authentication.

*   **Concurrency & Interactivity:** The system is designed to handle concurrent write operations, such as multiple users registering for a tournament simultaneously. It uses database transactions to maintain data integrity and prevent race conditions dynamically.

*   **Event-Driven Updates:** The architecture uses a message broker (RabbitMQ) to publish events (e.g., `TournamentStatusUpdated`). This design allows for decoupled, asynchronous updates across the system, setting the foundation for future real-time features like live bracket updates without requiring user-initiated refreshes.
