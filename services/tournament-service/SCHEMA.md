# Tournament Service Schema

## Database: `tournament_db`

This service is the source of truth for all tournament-related metadata.

### `tournaments` Table

Stores the core tournament information.

```sql
CREATE TABLE tournaments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organizer_id UUID NOT NULL, -- Logical link to User Service
    name VARCHAR(100) NOT NULL,
    description TEXT,
    game VARCHAR(50) NOT NULL, -- 'rocket-league', 'csgo'
    format VARCHAR(20) NOT NULL, -- 'single-elimination'
    start_date TIMESTAMP WITH TIME ZONE,
    status VARCHAR(20) DEFAULT 'draft', -- draft, registration, active, completed
    min_teams INT DEFAULT 2,
    max_teams INT DEFAULT 16
);
```

**Design Choices:**

*   **Status Flow:** The `status` column is the most important field in this table. It drives the logic of the entire tournament flow. For example, teams can only register for a tournament when the `status` is `registration`, and the bracket cannot be generated until the `status` is `active`.

### `registrations` Table

A join table that tracks which teams have registered for which tournaments.

```sql
CREATE TABLE registrations (
    tournament_id UUID REFERENCES tournaments(id),
    team_id UUID NOT NULL, -- Logical link to Team Service
    registered_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    status VARCHAR(20) DEFAULT 'approved',
    PRIMARY KEY (tournament_id, team_id)
);
```

**Design Choices:**

*   **Location:** This table is located in the `tournament-service` because it is primarily used to answer the question, "What teams are registered for this tournament?". This is a tournament-centric view of the data.
*   **Loose Coupling:** The `team_id` is a logical link to the `team-service`. We do not enforce a foreign key constraint to the `teams` table in the `team-service`'s database, as that would create a tight coupling between the two services.
