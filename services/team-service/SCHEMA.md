# Team Service Schema

## Database: `team_db`

This service is the source of truth for all team-related data.

### `teams` Table

Stores the core team information.

```sql
CREATE TABLE teams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    tag VARCHAR(10) NOT NULL, -- e.g. [TSM]
    captain_id UUID NOT NULL, -- Logical link to User Service
    logo_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

**Design Choices:**

*   **Source of Truth:** This service and this table are the single source of truth for the concept of a "Team." Any other service that needs to know about a team should query this service.
*   **Captaincy:** We store the `captain_id` to quickly check permissions for actions like editing the team name or inviting new members. This is a logical link to the `users` table in the `user-service`.

### `invites` Table

Stores information about invitations sent to users to join a team.

```sql
CREATE TABLE invites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID REFERENCES teams(id),
    inviter_id UUID NOT NULL, -- User who sent it
    invitee_email VARCHAR(255) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending', -- pending, accepted, rejected
    expires_at TIMESTAMP WITH TIME ZONE
);
```

**Design Choices:**

*   **Invite by Email:** We are inviting users by email. This allows us to invite users who may not have an account on the platform yet. When the invitee accepts the invitation, we can then link the invitation to their user account.
*   **Status Flow:** The `status` column drives the logic of the invitation system. For example, a user cannot accept an invitation that has already been accepted or rejected.
