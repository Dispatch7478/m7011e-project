# User Service Schema

## Database: `user_db`

This service is the source of truth for user-related data that is not directly related to authentication.

### `users` Table

Stores the core user profile.

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

**Design Choices:**

*   **`id`:** The `id` column directly corresponds to the user's ID from Keycloak, serving as the primary key.
*   **Caching:** `username` and `email` are cached from Keycloak for direct access and to reduce direct calls to the authentication service for basic user profile information.