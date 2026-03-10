import http from 'k6/http';
import { check } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://127.0.0.1:8000';
const MODE = __ENV.MODE || 'ramp';
const LOGIN_RATE = Number(__ENV.LOGIN_RATE || 30);
const AUTH_READ_RATE = Number(__ENV.AUTH_READ_RATE || 50);
const REFRESH_RATE = Number(__ENV.REFRESH_RATE || 20);
const DURATION = __ENV.DURATION || '3m';
const PRE_ALLOCATED_VUS = Number(__ENV.PRE_ALLOCATED_VUS || 50);
const MAX_VUS = Number(__ENV.MAX_VUS || 500);
const P95_MS = Number(__ENV.P95_MS || 200);
const P99_MS = Number(__ENV.P99_MS || 500);
const FAIL_RATE = Number(__ENV.FAIL_RATE || 0.001);
const LOGIN_EMAIL = __ENV.LOGIN_EMAIL || '';
const LOGIN_PASSWORD = __ENV.LOGIN_PASSWORD || '';
const ACCESS_TOKEN = __ENV.ACCESS_TOKEN || '';
const REFRESH_TOKEN = __ENV.REFRESH_TOKEN || '';

function selectedScenarios() {
  const raw = __ENV.SCENARIOS || '';
  if (raw) {
    return raw
      .split(',')
      .map((value) => value.trim())
      .filter(Boolean);
  }

  if (LOGIN_EMAIL && LOGIN_PASSWORD) {
    return ['login', 'read', 'refresh'];
  }
  if (ACCESS_TOKEN && REFRESH_TOKEN) {
    return ['read', 'refresh'];
  }
  if (ACCESS_TOKEN) {
    return ['read'];
  }

  return ['login', 'read', 'refresh'];
}

const ENABLED_SCENARIOS = selectedScenarios();

function hasScenario(name) {
  return ENABLED_SCENARIOS.includes(name);
}

function buildScenario(name, rate) {
  if (MODE === 'steady') {
    return {
      executor: 'constant-arrival-rate',
      exec: name,
      rate,
      timeUnit: '1s',
      duration: DURATION,
      preAllocatedVUs: PRE_ALLOCATED_VUS,
      maxVUs: MAX_VUS,
    };
  }

  return {
    executor: 'ramping-arrival-rate',
    exec: name,
    startRate: Math.max(1, Math.floor(rate / 10)),
    timeUnit: '1s',
    preAllocatedVUs: PRE_ALLOCATED_VUS,
    maxVUs: MAX_VUS,
    stages: [
      { target: Math.max(5, Math.floor(rate * 0.25)), duration: '1m' },
      { target: Math.max(10, Math.floor(rate * 0.5)), duration: '1m' },
      { target: rate, duration: '1m' },
      { target: rate, duration: '1m' },
    ],
  };
}

function jsonParams(token, tags) {
  const headers = { 'Content-Type': 'application/json' };
  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }
  return { headers, tags };
}

function requireCredentials() {
  if (!LOGIN_EMAIL || !LOGIN_PASSWORD) {
    throw new Error('Set LOGIN_EMAIL and LOGIN_PASSWORD, or provide ACCESS_TOKEN/REFRESH_TOKEN.');
  }
}

function loginOnce() {
  requireCredentials();
  const response = http.post(
    `${BASE_URL}/v1/auth/login/email-password`,
    JSON.stringify({ email: LOGIN_EMAIL, password: LOGIN_PASSWORD }),
    jsonParams('', { endpoint: 'auth.login', profile: 'setup-login' })
  );

  check(response, {
    'setup login status is 200': (res) => res.status === 200,
  });

  if (response.status !== 200) {
    throw new Error(`setup login failed with status ${response.status}`);
  }

  const body = response.json();
  return {
    accessToken: body.accessToken,
    refreshToken: body.refreshToken,
  };
}

export const options = {
  scenarios: {
    ...(hasScenario('login')
      ? { loginByPassword: buildScenario('loginByPassword', LOGIN_RATE) }
      : {}),
    ...(hasScenario('read')
      ? { authenticatedRead: buildScenario('authenticatedRead', AUTH_READ_RATE) }
      : {}),
    ...(hasScenario('refresh')
      ? { refreshTokenFlow: buildScenario('refreshTokenFlow', REFRESH_RATE) }
      : {}),
  },
  thresholds: {
    http_req_failed: [`rate<${FAIL_RATE}`],
    http_req_duration: [`p(95)<${P95_MS}`, `p(99)<${P99_MS}`],
  },
};

export function setup() {
  if (!hasScenario('read') && !hasScenario('refresh')) {
    return {
      accessToken: ACCESS_TOKEN,
      refreshToken: REFRESH_TOKEN,
    };
  }

  if (ACCESS_TOKEN) {
    return {
      accessToken: ACCESS_TOKEN,
      refreshToken: REFRESH_TOKEN,
    };
  }
  return loginOnce();
}

export function loginByPassword() {
  requireCredentials();
  const response = http.post(
    `${BASE_URL}/v1/auth/login/email-password`,
    JSON.stringify({ email: LOGIN_EMAIL, password: LOGIN_PASSWORD }),
    jsonParams('', { endpoint: 'auth.login', profile: 'login' })
  );

  check(response, {
    'login status is 200': (res) => res.status === 200,
  });
}

export function authenticatedRead(data) {
  if (!data.accessToken) {
    throw new Error('No access token available. Provide ACCESS_TOKEN or allow setup() to login.');
  }

  const responseInfo = http.get(
    `${BASE_URL}/v1/user/info`,
    jsonParams(data.accessToken, { endpoint: 'user.info', profile: 'auth-read' })
  );

  check(responseInfo, {
    'user info status is 200': (res) => res.status === 200,
  });

  const responsePrivate = http.post(
    `${BASE_URL}/v1/test/private`,
    JSON.stringify({}),
    jsonParams(data.accessToken, { endpoint: 'test.private', profile: 'auth-read' })
  );

  check(responsePrivate, {
    'private test status is 200': (res) => res.status === 200,
  });
}

export function refreshTokenFlow(data) {
  if (!data.refreshToken) {
    throw new Error('No refresh token available. Provide REFRESH_TOKEN or allow setup() to login.');
  }

  const response = http.post(
    `${BASE_URL}/v1/auth/refresh-token`,
    JSON.stringify({ refreshToken: data.refreshToken }),
    jsonParams('', { endpoint: 'auth.refresh', profile: 'refresh' })
  );

  check(response, {
    'refresh status is 200': (res) => res.status === 200,
  });
}
