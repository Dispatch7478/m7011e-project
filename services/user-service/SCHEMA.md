# User Service Schema

## Database: `user_db`

This service is the source of truth for user-related data that is not directly related to authentication.

### `users` Table

Stores the core user profile.

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY,
    keycloak_id VARCHAR(255) UNIQUE NOT NULL, -- The link to Auth
    username VARCHAR(50) NOT NULL,            -- Cached from Keycloak for easy joins
    email VARCHAR(255) NOT NULL,              -- Cached from Keycloak
    avatar_url TEXT,
    bio TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

**Design Choices:**

*   **`id` vs `keycloak_id`:** We use the `keycloak_id` as the primary means of identifying a user across the entire system. The `id` column in this table is the same as the `keycloak_id`, and it is used as the primary key for this table. This ensures a single source of truth for user identity.
*   **Caching:** We cache the `username` and `email` from Keycloak in this table. This is a deliberate design choice to improve performance and reduce the need for frequent calls to the Keycloak API. For example, when displaying a user's profile, we can get the username directly from this table instead of having to make a separate API call to Keycloak.

### `team_memberships` Table

A join table that tracks which teams a user is a member of.

```sql
CREATE TABLE team_memberships (
    user_id UUID REFERENCES users(id),
    team_id UUID NOT NULL, -- Logical link to Team Service
    role VARCHAR(20) NOT NULL CHECK (role IN ('captain', 'member')),
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (user_id, team_id)
);
```

**Design Choices:**

*   **Location:** This table is located in the `user-service` because it is primarily used to answer the question, "What teams is this user a member of?". This is a user-centric view of the data.
*   **Loose Coupling:** The `team_id` is a "logical link" to the `team-service`. We do not enforce a foreign key constraint to the `teams` table in the `team-service`'s database. This is because that table lives in a different database, and enforcing a foreign key would create a tight coupling between the two services. This design choice improves the resilience of the system; if the `team-service` is down, the `user-service` can still operate.
