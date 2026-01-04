import http from 'k6/http';
import { check, sleep } from 'k6';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

// --- CONFIGURATION ---
const BASE_URL = 'https://api.ltu-m7011e-4.se/api';
const AUTH_URL = 'https://keycloak.ltu-m7011e-4.se/realms/t-hub/protocol/openid-connect/token';

// !!! ENTER YOUR REAL CREDENTIALS HERE !!!
const USERNAME = 'player2'; 
const PASSWORD = 'testpass123'; 
const CLIENT_ID = 't-hub-frontend'; // This is a guess based on your token ("azp": "t-hub-frontend")

export const options = {
  stages: [
    { duration: '30s', target: 10 },
    { duration: '1m',  target: 50 },
    { duration: '30s', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'],
    http_req_failed: ['rate<0.10'],
  },
};

// This function runs once at the start to get a token
export function setup() {
  const payload = {
    grant_type: 'password',
    client_id: CLIENT_ID,
    username: USERNAME,
    password: PASSWORD,
  };

  const res = http.post(AUTH_URL, payload, {
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
  });

  if (res.status !== 200) {
    throw new Error(`Login failed! Status: ${res.status}. Body: ${res.body}`);
  }

  // extract the access_token
  return res.json().access_token;
}

export default function (token) {
  // The 'token' argument comes from the setup() function
  
  const uniqueId = uuidv4().substring(0, 8);
  
  const params = {
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
  };

  // 1. Scenario: GET Tournaments
  const resTournaments = http.get(`${BASE_URL}/tournaments`, params);
  
  check(resTournaments, {
    'GET tournaments status is 200': (r) => r.status === 200,
  });

  sleep(1);

  // 2. Scenario: GET My Teams
  const resMyTeams = http.get(`${BASE_URL}/teams/me/teams`, params);
  check(resMyTeams, {
    'GET my teams status is 200': (r) => r.status === 200,
  });

  sleep(1);

  // 3. Scenario: Create Team
  const teamPayload = JSON.stringify({
    name: `LoadTest-${uniqueId}`,
    tag: uniqueId.substring(0, 4).toUpperCase(),
    logo_url: null
  });

  const resCreateTeam = http.post(`${BASE_URL}/teams`, teamPayload, params);
  
  // Log unexpected errors
  if (resCreateTeam.status !== 200 && resCreateTeam.status !== 409) {
    console.error(`Create Team failed. Status: ${resCreateTeam.status}. Body: ${resCreateTeam.body}`);
  }
  
  check(resCreateTeam, {
    'Create Team status is 200 or 409': (r) => r.status === 200 || r.status === 409,
  });

  sleep(2);
}