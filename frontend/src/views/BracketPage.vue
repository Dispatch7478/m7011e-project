<template>
  <div class="bracket-page">
    <h1>Bracket for {{ tournament.name }}</h1>
    <div v-if="matches.length > 0" class="bracket-container">
      <!-- Left Bracket -->
      <div class="bracket-side left">
        <div class="round" v-for="round in leftRounds" :key="'left-' + round.number">
          <h2>{{ round.name }}</h2>
          <div class="match" v-for="match in round.matches" :key="match.id">
            <div class="team">{{ getTeamName(match.team_a_id) }}<span>{{ match.score_a }}</span></div>
            <div class="team">{{ getTeamName(match.team_b_id) }}<span>{{ match.score_b }}</span></div>
          </div>
        </div>
      </div>

      <!-- Final -->
      <div class="bracket-final">
        <div class="round" v-if="finalRound">
          <h2>Final</h2>
          <div class="match final-match">
             <div class="team">{{ getTeamName(finalRound.matches[0].team_a_id) }}<span>{{ finalRound.matches[0].score_a }}</span></div>
             <div class="team">{{ getTeamName(finalRound.matches[0].team_b_id) }}<span>{{ finalRound.matches[0].score_b }}</span></div>
          </div>
        </div>
      </div>

      <!-- Right Bracket -->
      <div class="bracket-side right">
        <div class="round" v-for="round in rightRounds" :key="'right-' + round.number">
          <h2>{{ round.name }}</h2>
          <div class="match" v-for="match in round.matches" :key="match.id">
            <div class="team"><span>{{ match.score_a }}</span>{{ getTeamName(match.team_a_id) }}</div>
            <div class="team"><span>{{ match.score_b }}</span>{{ getTeamName(match.team_b_id) }}</div>
          </div>
        </div>
      </div>
    </div>
    <div v-else>
      <p>Loading bracket...</p>
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
      teams: [],
    };
  },
  computed: {
    rounds() {
      if (this.matches.length === 0) return [];
      const roundsData = {};
      this.matches.forEach(match => {
        if (!roundsData[match.round_number]) {
          roundsData[match.round_number] = { number: match.round_number, matches: [], name: '' };
        }
        roundsData[match.round_number].matches.push(match);
      });
      
      let roundsArray = Object.values(roundsData).sort((a, b) => a.number - b.number);
      
      // Assign names to rounds based on total number of rounds
      if(roundsArray.length === 5) { // 32 teams
        roundsArray[0].name = 'Round of 32';
        roundsArray[1].name = 'Round of 16';
        roundsArray[2].name = 'Quarterfinals';
        roundsArray[3].name = 'Semifinals';
      } else if (roundsArray.length === 4) { // 16 teams
         roundsArray[0].name = 'Round of 16';
         roundsArray[1].name = 'Quarterfinals';
         roundsArray[2].name = 'Semifinals';
      } else if (roundsArray.length === 3) { // 8 teams
        roundsArray[0].name = 'Quarterfinals';
        roundsArray[1].name = 'Semifinals';
      }

      return roundsArray;
    },
    leftRounds() {
      const allRounds = this.rounds;
      if (allRounds.length < 2) return [];

      const mid = Math.ceil(allRounds[0].matches.length / 2);
      let left = allRounds.slice(0, -1).map(round => {
        const leftMatches = round.matches.filter(m => m.match_number <= mid);
        return {...round, matches: leftMatches};
      });
      return left;
    },
    rightRounds() {
       const allRounds = this.rounds;
      if (allRounds.length < 2) return [];

      const mid = Math.ceil(allRounds[0].matches.length / 2);
      let right = allRounds.slice(0, -1).map(round => {
        const rightMatches = round.matches.filter(m => m.match_number > mid).sort((a,b) => a.match_number - b.match_number);
        return {...round, matches: rightMatches};
      });
      return right.reverse();
    },
    finalRound() {
      if (this.rounds.length === 0) return null;
      return this.rounds[this.rounds.length - 1];
    }
  },
  methods: {
    async fetchBracket() {
      const tournamentId = this.$route.params.id;
      try {
        const [tournamentResponse, teamsResponse, bracketResponse] = await Promise.all([
          securedApi.get(`/api/tournaments/${tournamentId}`),
          securedApi.get(`/api/tournaments/${tournamentId}/teams`),
          securedApi.get(`/api/tournaments/${tournamentId}/bracket`)
        ]);

        this.tournament = tournamentResponse.data;
        this.teams = teamsResponse.data;
        this.matches = bracketResponse.data.matches;

      } catch (error) {
        console.error('Failed to fetch bracket data:', error);
      }
    },
    getTeamName(teamId) {
      if (!teamId) return 'TBD';
      const team = this.teams.find(t => t.id === teamId);
      return team ? team.name : 'Unknown Team';
    }
  },
  mounted() {
    this.fetchBracket();
  }
}
</script>

<style scoped>
.bracket-page {
  padding: 20px;
}
.bracket-container {
  display: flex;
  justify-content: space-between;
  overflow-x: auto;
}
.bracket-side {
  display: flex;
  gap: 30px;
}
.bracket-side.right {
  direction: rtl; 
}
.bracket-side.right .round {
  direction: ltr;
}
.bracket-final {
  display: flex;
  align-items: center;
}
.round {
  display: flex;
  flex-direction: column;
  justify-content: space-around;
  gap: 30px;
}
.match {
  background: #fff;
  border: 1px solid #ccc;
  border-radius: 5px;
  width: 200px;
  text-align: left;
}
.team {
  padding: 8px 12px;
  display: flex;
  justify-content: space-between;
  border-bottom: 1px solid #eee;
}
.team:last-child {
  border-bottom: none;
}
.team span {
  font-weight: bold;
}
.bracket-side.right .team span {
  margin-right: 10px;
}
.bracket-side.left .team span {
  margin-left: 10px;
}
.final-match {
    margin: 0 20px;
}
</style>
