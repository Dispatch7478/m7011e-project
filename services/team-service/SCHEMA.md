# Team Service Schema

## Database: `team_db`

This service is the source of truth for all team-related data.

### `teams` Table

Stores the core team information. A team's name and tag must be unique.

```sql
CREATE TABLE teams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    tag VARCHAR(10) NOT NULL,
    captain_id UUID NOT NULL, -- Keycloak user id
    logo_url TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

ALTER TABLE teams
  ADD CONSTRAINT teams_name_unique UNIQUE (name),
  ADD CONSTRAINT teams_tag_unique UNIQUE (tag);
```

### `team_members` Table

A join table that links users to teams and defines their role.

```sql
CREATE TABLE team_members (
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    user_id UUID NOT NULL, -- Keycloak user id
    role VARCHAR(20) DEFAULT 'member', -- captain, member, admin
    joined_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (team_id, user_id)
);
```

### `invites` Table

Stores invitations sent to users to join a team.

```sql
CREATE TABLE invites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID REFERENCES teams(id),
    inviter_id UUID NOT NULL,
    invitee_email VARCHAR(255) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    expires_at TIMESTAMPTZ
);
```