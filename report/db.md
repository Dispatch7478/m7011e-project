# Database Schema Overview

This document provides a complete overview of the database schemas for all microservices. It includes an Entity-Relationship Diagram (ERD) to visualize the relationships between tables, as well as the full SQL `CREATE TABLE` statements for each service.

## Entity-Relationship Diagram (ERD)

This diagram shows the tables for each service and the relationships between them.

```mermaid
erDiagram
    %% --- User Service ---
    users {
        UUID id PK
        VARCHAR username
        VARCHAR email
        TIMESTAMPTZ created_at
    }

    %% --- Team Service ---
    teams {
        UUID id PK
        VARCHAR name
        VARCHAR tag
        UUID captain_id "FK to users.id"
        TEXT logo_url
        TIMESTAMPTZ created_at
    }
    team_members {
        UUID team_id PK, FK
        UUID user_id PK, FK
        VARCHAR role
        TIMESTAMPTZ joined_at
    }
    invites {
        UUID id PK
        UUID team_id FK
        UUID inviter_id "FK to users.id"
        VARCHAR invitee_email
        VARCHAR status
        TIMESTAMPTZ expires_at
    }

    %% --- Tournament Service ---
    tournaments {
        UUID id PK
        UUID organizer_id "FK to users.id"
        VARCHAR name
        VARCHAR game
        VARCHAR format
        VARCHAR status
        INT max_participants
    }
    registrations {
        UUID tournament_id PK, FK
        UUID participant_id PK "User or Team ID"
        VARCHAR participant_name
        TIMESTAMPTZ registered_at
    }

    %% --- Bracket Service ---
    brackets {
        UUID id PK
        UUID tournament_id FK
        VARCHAR bracket_type
        VARCHAR status
    }
    matches {
        UUID id PK
        UUID bracket_id FK
        INT round
        UUID player1_id
        UUID player2_id
        INT player1_score
        INT player2_score
        UUID winner_id
        VARCHAR status
    }

    %% --- Relationships ---
    users ||--o{ teams : "is_captain_of"
    users ||--o{ team_members : "is_member_of"
    users ||--o{ tournaments : "is_organizer_of"
    teams ||--|{ team_members : "has"
    teams ||--o{ invites : "sends"
    tournaments ||--|{ registrations : "has"
    tournaments ||--|{ brackets : "has"
    brackets ||--|{ matches : "has"
```

---

## SQL Schemas by Service

### User Service (`user_db`)

For the detailed schema, see [`services/user-service/SCHEMA.md`](../services/user-service/SCHEMA.md).

### Team Service (`team_db`)

For the detailed schema, see [`services/team-service/SCHEMA.md`](../services/team-service/SCHEMA.md).

### Tournament Service (`tournament_db`)

For the detailed schema, see [`services/tournament-service/SCHEMA.md`](../services/tournament-service/SCHEMA.md).

### Bracket Service (`bracket_db`)

For the detailed schema, see [`services/bracket-service/SCHEMA.md`](../services/bracket-service/SCHEMA.md).
