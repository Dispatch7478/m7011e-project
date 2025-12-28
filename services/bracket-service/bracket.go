package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

type BracketHandler struct {
	DB                   *pgxpool.Pool
	RMQ                  *Service
	TournamentServiceURL string
}

// Struct to parse participants from Tournament Service
type Participant struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Match struct {
    ID           string  `json:"id"`
    TournamentID string  `json:"tournament_id"`
    Round        int     `json:"round"`
    MatchNumber  int     `json:"match_number"`
    Player1ID    *string `json:"player1_id"`
    Player2ID    *string `json:"player2_id"`
    NextMatchID  *string `json:"next_match_id"`
    Status       string  `json:"status"`
    ScoreA       *string `json:"score_a"` 
    ScoreB       *string `json:"score_b"`
    WinnerID     *string `json:"winner_id"`
}

func (h *BracketHandler) GenerateBracket(c echo.Context) error {
	tournamentID := c.QueryParam("tournament_id")
	if tournamentID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "tournament_id is required"})
	}

	// 1. Fetch Participants from Tournament Service
	resp, err := http.Get(fmt.Sprintf("%s/tournaments/%s/participants", h.TournamentServiceURL, tournamentID))
	if err != nil || resp.StatusCode != 200 {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch participants"})
	}
	defer resp.Body.Close()

	var participants []Participant
	if err := json.NewDecoder(resp.Body).Decode(&participants); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to decode participants"})
	}

	count := len(participants)
	if count < 2 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Not enough participants to generate a bracket"})
	}

	// 2. Shuffle Participants
	// Note: In Go 1.20+, the global random source is automatically seeded, so we don't need "time" to seed it.
	rand.Shuffle(count, func(i, j int) { participants[i], participants[j] = participants[j], participants[i] })

	// 3. Calculate Bracket Depth
	power := math.Ceil(math.Log2(float64(count)))
	rounds := int(power)

	// 4. Generate Matches
	tx, err := h.DB.Begin(context.Background())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "DB Transaction failed"})
	}
	defer tx.Rollback(context.Background())

	// Map to keep track of created matches to link next_match_id
	matchMap := make(map[string]string)

	for r := rounds; r >= 1; r-- {
		matchesInRound := int(math.Pow(2, float64(rounds-r))) 
		
		for m := 1; m <= matchesInRound; m++ {
			var nextMatchID *string
			
			// If not the final, find the next match ID
			if r < rounds {
				nextRoundMatchNum := (m + 1) / 2
				key := fmt.Sprintf("%d-%d", r+1, nextRoundMatchNum)
				if id, exists := matchMap[key]; exists {
					nextMatchID = &id
				}
			}

			var p1, p2 *string
			status := "scheduled"
			
			if r == 1 {
				idx1 := (m - 1) * 2
				idx2 := idx1 + 1

				if idx1 < count {
					p1 = &participants[idx1].ID
				}
				if idx2 < count {
					p2 = &participants[idx2].ID
				}

				// Handle BYE
				if p1 != nil && p2 == nil {
					status = "completed"
				}
			}

			var matchID string
			err := tx.QueryRow(context.Background(), `
				INSERT INTO matches (tournament_id, round, match_number, player1_id, player2_id, next_match_id, status)
				VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id
			`, tournamentID, r, m, p1, p2, nextMatchID, status).Scan(&matchID)

			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save match"})
			}

			matchMap[fmt.Sprintf("%d-%d", r, m)] = matchID
		}
	}

	if err := tx.Commit(context.Background()); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit bracket"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Bracket generated successfully", "rounds": fmt.Sprintf("%d", rounds)})
}

func GetParticipantsHandler(db *pgxpool.Pool) echo.HandlerFunc {
	return func(c echo.Context) error {
		tournamentID := c.Param("id")

		query := `
			SELECT participant_id, participant_name 
			FROM registrations 
			WHERE tournament_id = $1 AND status = 'approved'
		`
		rows, err := db.Query(context.Background(), query, tournamentID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch participants"})
		}
		defer rows.Close()

		type Participant struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}
		participants := []Participant{}

		for rows.Next() {
			var p Participant
			if err := rows.Scan(&p.ID, &p.Name); err != nil {
				continue
			}
			participants = append(participants, p)
		}

		return c.JSON(http.StatusOK, participants)
	}
}


func (h *BracketHandler) GetBracket(c echo.Context) error {
	tournamentID := c.Param("tournament_id") // Matches the :tournament_id in main.go

	// 1. Query Matches
    // We explicitly select columns to match your struct fields
	query := `
		SELECT id, tournament_id, round, match_number, 
               player1_id, player2_id, next_match_id, status,
               COALESCE(score_a, ''), COALESCE(score_b, '')
		FROM matches 
		WHERE tournament_id = $1
        ORDER BY round DESC, match_number ASC
	`
	rows, err := h.DB.Query(context.Background(), query, tournamentID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch bracket"})
	}
	defer rows.Close()

	matches := []Match{}
	for rows.Next() {
		var m Match
        // Temporary vars for scores to handle potential NULLs safely
        var sA, sB string 

		err := rows.Scan(
            &m.ID, &m.TournamentID, &m.Round, &m.MatchNumber, 
            &m.Player1ID, &m.Player2ID, &m.NextMatchID, &m.Status,
            &sA, &sB,
        )
		if err != nil {
            // Log error but continue? Or return error. 
            // Ideally log it: log.Printf("Scan error: %v", err)
			continue
		}
        m.ScoreA = &sA
        m.ScoreB = &sB
		matches = append(matches, m)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"matches": matches,
	})
}


type ResultRequest struct {
	ScoreA   string `json:"score_a"`
	ScoreB   string `json:"score_b"`
	WinnerID string `json:"winner_id"`
}

func (h *BracketHandler) UpdateMatchResult(c echo.Context) error {
	matchID := c.Param("match_id")
	var req ResultRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	ctx := context.Background()

	// 1. Start Transaction (Critical for integrity)
	tx, err := h.DB.Begin(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "DB Error"})
	}
	defer tx.Rollback(ctx)

	// 2. Fetch Current Match to get 'NextMatchID' and 'MatchNumber'
	var nextMatchID *string
	var matchNum int
	err = tx.QueryRow(ctx, `SELECT next_match_id, match_number FROM matches WHERE id = $1`, matchID).Scan(&nextMatchID, &matchNum)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Match not found"})
	}

	// 3. Update Current Match
	_, err = tx.Exec(ctx, `
		UPDATE matches 
		SET score_a = $1, score_b = $2, winner_id = $3, status = 'completed' 
		WHERE id = $4`,
		req.ScoreA, req.ScoreB, req.WinnerID, matchID,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update match result"})
	}

	// 4. Advance Winner to Next Match (if not the final)
	if nextMatchID != nil {
		// Logic: If MatchNumber is Odd (1,3,5), winner goes to Player1 slot of next match.
		//        If MatchNumber is Even (2,4,6), winner goes to Player2 slot.
		updateField := "player1_id"
		if matchNum%2 == 0 {
			updateField = "player2_id"
		}

		query := fmt.Sprintf("UPDATE matches SET %s = $1 WHERE id = $2", updateField)
		_, err = tx.Exec(ctx, query, req.WinnerID, *nextMatchID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to advance winner"})
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Commit failed"})
	}

	// 5. Publish Event (for other services)
	// _ = h.RMQ.Publish("events.match.completed", ...)

	return c.JSON(http.StatusOK, map[string]string{"message": "Match updated"})
}