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
            <input type="number" v-model="scoreA" min="0"/>
          </div>
          <div class="vs">VS</div>
          <div class="team-input">
            <label>{{ getParticipantName(selectedMatch.player2_id) }}</label>
            <input type="number" v-model="scoreB" min="0"/>
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

      // Logic: Matches in the first half of the draw go left
      // This logic assumes match_number is ordered 1..N
      return allRounds.slice(0, -1).map(round => {
        const mid = Math.ceil(round.matches.length / 2);
        // Take the first half of matches
        const leftMatches = round.matches
          .filter(m => m.match_number <= mid)
          .sort((a,b) => a.match_number - b.match_number);
        return {...round, matches: leftMatches};
      });
    },
    rightRounds() {
       const allRounds = this.rounds;
      if (allRounds.length < 2) return [];

      return allRounds.slice(0, -1).map(round => {
        const mid = Math.ceil(round.matches.length / 2);
        // Take the second half of matches
        const rightMatches = round.matches
          .filter(m => m.match_number > mid)
          .sort((a,b) => a.match_number - b.match_number); // Ascending order
        return {...round, matches: rightMatches};
      }); // Right side should not be reversed here (testing)
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