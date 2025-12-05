# Bracket Service Schema

## Database: `bracket_db`

This service is responsible for the complex logic of generating, managing, and updating the tournament brackets.

### `brackets` Table

Stores the root of a bracket for a specific tournament.

```sql
CREATE TABLE brackets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tournament_id UUID UNIQUE NOT NULL, -- Link to Tournament Service
    type VARCHAR(20) NOT NULL -- 'single-elimination'
);
```

### `matches` Table

The nodes in the bracket tree. Each row represents a single match between two teams.

```sql
CREATE TABLE matches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bracket_id UUID REFERENCES brackets(id),
    round_number INT NOT NULL,   -- 1 = Round of 16, 2 = Quarterfinals, etc.
    match_number INT NOT NULL,   -- Position in the round (1, 2, 3...)
    
    -- The Participants (Teams)
    team_a_id UUID, -- Can be NULL if TBD
    team_b_id UUID, -- Can be NULL if TBD
    
    -- The Result
    score_a INT DEFAULT 0,
    score_b INT DEFAULT 0,
    winner_id UUID, -- Populated when match finishes
    
    -- The Link to the Next Match (The "Tree" Structure)
    next_match_id UUID, -- The match the winner advances to
    next_match_slot VARCHAR(1) -- 'A' or 'B' (does winner become Team A or Team B?)
);
```

**Design Choices:**

*   **Self-Referencing Logic:** The `next_match_id` is the key to the entire bracket progression system. When a winner is declared for a match, the service can look up the `next_match_id` and update the appropriate team slot (`team_a_id` or `team_b_id`) with the `winner_id`. This is how the bracket "flows" automatically as matches are completed.
*   **Nulls are Okay:** The `team_a_id`, `team_b_id`, and `winner_id` fields are all nullable. This is essential, as the participants and winner of a match are unknown until the preceding matches have been played. The initial state of a bracket will have many nulls, which get filled in as the tournament progresses.
*   **Decoupled from Tournament Logic:** This service only cares about the structure of the bracket and the scores. It does not need to know about the tournament's name, description, or rules. It receives the list of participating teams from the `tournament-service` when the bracket is first generated.
