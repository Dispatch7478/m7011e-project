<template>
  <div class="bracket-page">
    <header class="bracket-header">
      <h1>Bracket: {{ tournament.name }}</h1>
      <router-link :to="{ name: 'Tournaments' }" class="back-link">← Back to Tournaments</router-link>
    </header>

    <div v-if="matches.length > 0" class="bracket-container">
      <div class="bracket-side left">
        <div class="round" v-for="round in leftRounds" :key="'left-' + round.number">
          <h2 class="round-title">{{ round.name }}</h2>
          <div class="match" v-for="match in round.matches" :key="match.id" @click="reportScore(match)">
            <div class="participant" :class="{ 'winner': isWinner(match, match.player1_id) }">
              <span class="name">{{ getParticipantName(match.player1_id) }}</span>
              <span class="score">{{ match.score_a || '-' }}</span>
            </div>
            <div class="participant" :class="{ 'winner': isWinner(match, match.player2_id) }">
              <span class="name">{{ getParticipantName(match.player2_id) }}</span>
              <span class="score">{{ match.score_b || '-' }}</span>
            </div>
          </div>
        </div>
      </div>

      <div class="bracket-final">
        <div class="round" v-if="finalRound">
          <h2 class="round-title">Final</h2>
          <div class="match final-match" @click="reportScore(finalRound.matches[0])">
             <div class="participant" :class="{ 'winner': isWinner(finalRound.matches[0], finalRound.matches[0].player1_id) }">
               <span class="name">{{ getParticipantName(finalRound.matches[0].player1_id) }}</span>
               <span class="score">{{ finalRound.matches[0].score_a || '-' }}</span>
             </div>
             <div class="participant" :class="{ 'winner': isWinner(finalRound.matches[0], finalRound.matches[0].player2_id) }">
               <span class="name">{{ getParticipantName(finalRound.matches[0].player2_id) }}</span>
               <span class="score">{{ finalRound.matches[0].score_b || '-' }}</span>
             </div>
          </div>
        </div>
      </div>

      <div class="bracket-side right">
        <div class="round" v-for="round in rightRounds" :key="'right-' + round.number">
          <h2 class="round-title">{{ round.name }}</h2>
          <div class="match" v-for="match in round.matches" :key="match.id" @click="reportScore(match)">
            <div class="participant" :class="{ 'winner': isWinner(match, match.player1_id) }">
              <span class="score">{{ match.score_a || '-' }}</span>
              <span class="name">{{ getParticipantName(match.player1_id) }}</span>
            </div>
            <div class="participant" :class="{ 'winner': isWinner(match, match.player2_id) }">
              <span class="score">{{ match.score_b || '-' }}</span>
              <span class="name">{{ getParticipantName(match.player2_id) }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
    
    <div v-else-if="loading" class="loading-state">
      <p>Loading bracket data...</p>
    </div>
    
    <div v-else class="empty-state">
      <p>No bracket generated yet.</p>
     
    </div>
    <div v-if="showScoreModal" class="modal-overlay">
      <div class="modal-content">
        <h3>Report Score</h3>
        <p>Enter the final score for this match.</p>
        
        <div class="score-inputs">
          <div class="team-input">
            <label>{{ getParticipantName(selectedMatch.player1_id) }}</label>
            <input type="number" v-model="scoreA" min="0">
          </div>
          <div class="vs">VS</div>
          <div class="team-input">
            <label>{{ getParticipantName(selectedMatch.player2_id) }}</label>
            <input type="number" v-model="scoreB" min="0">
          </div>
        </div>

        <div class="modal-actions">
          <button @click="closeModal" class="btn-cancel">Cancel</button>
          <button @click="submitScore" class="btn-confirm">Submit Result</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import securedApi from '@/axios-auth.js';

export default {
  name: 'BracketPage',
  data() {
    return {
      tournament: {},
      matches: [],
      participants: [], // Renamed from teams
      loading: true,
      showScoreModal: false,
      selectedMatch: null,
      scoreA: 0,
      scoreB: 0,
      currentUserId: null,
    };
  },
  computed: {

    rounds() {
      if (this.matches.length === 0) return [];
      const roundsData = {};
      
      this.matches.forEach(match => {
        // Handle "round_number" (Go default often outputs snake_case keys if tagged, or PascalCase if not)
        // We will assume you tagged them as json:"round" or json:"round_number". 
        // Adjust "match.round" based on your actual API response.
        const rNum = match.round || match.round_number; 
        
        if (!roundsData[rNum]) {
          roundsData[rNum] = { number: rNum, matches: [], name: '' };
        }
        roundsData[rNum].matches.push(match);
      });
      
      let roundsArray = Object.values(roundsData).sort((a, b) => a.number - b.number);
      
      // Sort matches within each round to ensure consistent order
      roundsArray.forEach(round => {
        round.matches.sort((a, b) => a.match_number - b.match_number);
      });
      
      // Dynamic Round Naming
      const totalRounds = roundsArray.length;
      roundsArray.forEach((round, index) => {
        // Reverse index: 0 = Final, 1 = Semis, etc. (if we count backwards from end)
        const roundFromFinal = totalRounds - index; 
        if (roundFromFinal === 1) round.name = 'Final';
        else if (roundFromFinal === 2) round.name = 'Semifinals';
        else if (roundFromFinal === 3) round.name = 'Quarterfinals';
        else round.name = `Round ${index + 1}`;
      });

      return roundsArray;
    },
    leftRounds() {
      // Split rounds for visual display (Standard Tree)
      const allRounds = this.rounds;
      if (allRounds.length < 2) return [];

      return allRounds.slice(0, -1).map(round => {
        const mid = Math.ceil(round.matches.length / 2);
        const leftMatches = round.matches.slice(0, mid);
        return {...round, matches: leftMatches};
      });
    },
    rightRounds() {
       const allRounds = this.rounds;
      if (allRounds.length < 2) return [];

      return allRounds.slice(0, -1).map(round => {
        const mid = Math.ceil(round.matches.length / 2);
        // Take the second half of matches and reverse their order for mirrored display
        const rightMatches = round.matches.slice(mid)
          .sort((a,b) => b.match_number - a.match_number); // Descending order
        return {...round, matches: rightMatches};
      });
    },
    finalRound() {
      if (this.rounds.length === 0) return null;
      return this.rounds[this.rounds.length - 1];
    }
  },
  methods: {
    async fetchBracket() {
      const tournamentId = this.$route.params.id;
      this.loading = true;
      try {
        // Parallel requests to Tournament Service (for names) and Bracket Service (for matches)
        const [tournamentResponse, participantsResponse, bracketResponse] = await Promise.all([
          securedApi.get(`/api/tournaments/${tournamentId}`),
          // Changed from /teams to /participants
          securedApi.get(`/api/tournaments/${tournamentId}/participants`),
          // Expecting the bracket service to return { matches: [...] }
          securedApi.get(`/api/brackets/${tournamentId}`)
        ]);

        this.tournament = tournamentResponse.data;
        this.participants = participantsResponse.data;
        this.matches = bracketResponse.data.matches || [];

      } catch (error) {
        console.error('Failed to fetch bracket data:', error);
      } finally {
        this.loading = false;
      }
    },
    getParticipantName(id) {
      if (!id) return 'TBD';
      const p = this.participants.find(p => p.id === id);
      return p ? p.name : 'Unknown';
    },
    isWinner(match, participantId) {
       return match.status === 'completed' && match.winner_id === participantId;
    },
    reportScore(match) {
      // 1. Prevent clicking on empty/completed matches
      if (!match.player1_id || !match.player2_id) return;
      if (match.status === 'completed') {
          alert("This match is already completed.");
          return;
      }

      // Permission check so that only the organizer or players can report scores.
      const isOrganizer = this.currentUserId === this.tournament.organizer_id;
      const isPlayerA = this.currentUserId === match.player1_id;
      const isPlayerB = this.currentUserId === match.player2_id;

      if (!isOrganizer && !isPlayerA && !isPlayerB) {
        console.log("User is spectator");
        return; 
      }

      this.selectedMatch = match;
      this.scoreA = 0;
      this.scoreB = 0;
      this.showScoreModal = true;
    },
    closeModal() {
      this.showScoreModal = false;
      this.selectedMatch = null;
    },
    async submitScore() {
      if (!this.selectedMatch) return;

      // Determine winner based on score
      let winnerId = null;
      if (this.scoreA > this.scoreB) winnerId = this.selectedMatch.player1_id;
      else if (this.scoreB > this.scoreA) winnerId = this.selectedMatch.player2_id;
      else {
        alert("Draws are not allowed in elimination brackets.");
        return;
      }

      try {
        // Call the new endpoint
        await securedApi.post(`/api/brackets/matches/${this.selectedMatch.id}/result`, {
          score_a: String(this.scoreA),
          score_b: String(this.scoreB),
          winner_id: winnerId
        });

        this.closeModal();
        this.fetchBracket(); // Refresh to see the winner 
        alert("Score updated!");
      } catch (error) {
        console.error("Failed to submit score:", error);
        alert("Failed to submit score.");
      }
    }
  },
  created() {
    if (this.$keycloak && this.$keycloak.tokenParsed) {
      this.currentUserId = this.$keycloak.tokenParsed.sub;
    }
  },
  mounted() {
    this.fetchBracket();
  }
}
</script>

<style scoped>
.bracket-page { padding: 20px; text-align: center; }
.bracket-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 20px; }
.back-link { text-decoration: none; color: #007bff; font-weight: bold; }

.bracket-container {
  display: flex;
  justify-content: center;
  align-items: center;
  overflow-x: auto;
  padding: 20px 0;
}

.bracket-side { display: flex; gap: 40px; }

/* Right Bracket Reversal */
.bracket-side.right { direction: rtl; } 
.bracket-side.right .round { direction: ltr; } /* Keep text LTR */

.bracket-final { margin: 0 40px; }

.round {
  display: flex;
  flex-direction: column;
  justify-content: space-around;
  gap: 20px;
}

.round-title { font-size: 0.9rem; color: #666; text-transform: uppercase; margin-bottom: 10px; }

.match {
  background: #fff;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  width: 220px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.05);
  cursor: pointer;
  transition: transform 0.2s, box-shadow 0.2s;
}

.match:hover { transform: translateY(-2px); box-shadow: 0 4px 6px rgba(0,0,0,0.1); border-color: #007bff; }

.participant {
  padding: 8px 12px;
  display: flex;
  justify-content: space-between;
  border-bottom: 1px solid #f3f4f6;
  font-size: 0.9rem;
}
.participant:last-child { border-bottom: none; }

.participant.winner { background-color: #f0fdf4; color: #15803d; font-weight: bold; }
.participant.winner .score { color: #15803d; }

.name { white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.score { font-weight: bold; color: #374151; }

.loading-state, .empty-state { color: #6b7280; margin-top: 50px; font-size: 1.1rem; }

.modal-overlay {
  position: fixed; top: 0; left: 0; width: 100%; height: 100%;
  background: rgba(0,0,0,0.5);
  display: flex; justify-content: center; align-items: center;
  z-index: 1000;
}
.modal-content {
  background: white; padding: 30px; border-radius: 8px;
  width: 400px; text-align: center;
}
.score-inputs {
  display: flex; justify-content: space-between; align-items: center;
  margin: 20px 0;
}
.team-input { display: flex; flex-direction: column; width: 40%; }
.team-input input { padding: 8px; text-align: center; font-size: 1.2em; }
.vs { font-weight: bold; color: #888; }
.modal-actions { display: flex; justify-content: space-between; gap: 10px; }
.btn-cancel, .btn-confirm { padding: 10px 20px; border-radius: 4px; cursor: pointer; border: none;}
.btn-confirm { background-color: #28a745; color: white; }
.btn-cancel { background-color: #ccc; }
</style>