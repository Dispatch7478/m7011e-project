import http from 'k6/http';
import { check } from 'k6';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

// --- CONFIGURATION ---
const BASE_URL = 'https://api.ltu-m7011e-4.se/api';
const AUTH_URL = 'https://keycloak.ltu-m7011e-4.se/realms/t-hub/protocol/openid-connect/token';

// !!! ENTER YOUR CREDENTIALS HERE !!!
const USERNAME = 'player2'; 
const PASSWORD = 'testpass123'; 
const CLIENT_ID = 't-hub-frontend'; 

export const options = {
  scenarios: {
    race_condition: {
      executor: 'per-vu-iterations',
      vus: 50,              // 50 concurrent users
      iterations: 1,        // Each user runs once
      maxDuration: '30s',
    },
  },
};

// --- SETUP PHASE (Runs Once) ---
export function setup() {
  // 1. LOGIN TO GET FRESH TOKEN
  const loginPayload = {
    grant_type: 'password',
    client_id: CLIENT_ID,
    username: USERNAME,
    password: PASSWORD,
  };

  const resLogin = http.post(AUTH_URL, loginPayload, {
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
  });

  if (resLogin.status !== 200) {
    throw new Error(`Login failed! Status: ${resLogin.status}. Body: ${resLogin.body}`);
  }
  const token = resLogin.json().access_token;
  
  // Headers for subsequent requests
  const params = {
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
  };

  // 2. CREATE TOURNAMENT (Small limit: 4 slots)
  const createPayload = JSON.stringify({
    name: `RaceTest-${uuidv4().substring(0, 5)}`,
    description: "Testing race conditions",
    game: "StressPong",
    format: "single-elimination",
    participant_type: "team", 
    min_participants: 2,
    max_participants: 4,      // THE LIMIT: Only 4 should succeed
    start_date: new Date(Date.now() + 86400000).toISOString(), 
  });

  const resCreate = http.post(`${BASE_URL}/tournaments`, createPayload, params);

  if (resCreate.status !== 201) {
    throw new Error(`Setup failed! Could not create tournament. Status: ${resCreate.status} ${resCreate.body}`);
  }

  const tournamentId = resCreate.json().id;
  console.log(`\n🏆 Tournament Created! ID: ${tournamentId}`);

  // 3. OPEN TOURNAMENT
  const openPayload = JSON.stringify({ status: 'registration_open' });
  const resOpen = http.patch(`${BASE_URL}/tournaments/${tournamentId}/status`, openPayload, params);

  if (resOpen.status !== 200) {
     throw new Error(`Setup failed! Could not open tournament. Status: ${resOpen.status} ${resOpen.body}`);
  }
  
  console.log(`🔓 Tournament Opened. Race starting...`);
  console.log(`Expected: 4 Successes, 46 Failures (409 Conflict)\n`);
  
  // Pass the ID and the Token to the VUs
  return { tournamentId, token };
}

// --- VIRTUAL USER LOGIC ---
export default function (data) {
  const tournamentId = data.tournamentId;
  const token = data.token; // Use the token from setup
  const fakeTeamId = uuidv4();
  
  const params = {
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
  };

  const payload = JSON.stringify({
    team_id: fakeTeamId,     
    name: `Racer-${fakeTeamId.substring(0,5)}`
  });

  // THE ATTACK: Try to register
  const res = http.post(`${BASE_URL}/tournaments/${tournamentId}/register`, payload, params);

  // We log failures that AREN'T "Tournament Full" errors
  if (res.status !== 201 && res.status !== 409) {
    console.error(`Unexpected Error: ${res.status} - ${res.body}`);
  }

  check(res, {
    'Status is 201 (Joined) or 409 (Full)': (r) => r.status === 201 || r.status === 409,
    'Got In! (201)': (r) => r.status === 201,
  });
}