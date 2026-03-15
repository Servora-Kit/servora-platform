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
const REFRESH_TOKENS = (__ENV.REFRESH_TOKENS || '')
  .split(',')
  .map((value) => value.trim())
  .filter(Boolean);

function selectedScenarios() {
  const raw = __ENV.SCENARIOS || '';
  if (raw) {
    return raw
      .split(',')
      .map((value) => value.trim())
      .filter(Boolean);
  }

  if (LOGIN_EMAIL && LOGIN_PASSWORD) {
    return ['login', 'read'];
  }
  if (ACCESS_TOKEN) {
    return ['read'];
  }
  if (REFRESH_TOKEN || REFRESH_TOKENS.length > 0) {
    return ['refresh'];
  }

  return ['login', 'read'];
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

function scenarioThresholds(profile) {
  return {
    [`http_req_failed{profile:${profile}}`]: [`rate<${FAIL_RATE}`],
    [`http_req_duration{profile:${profile}}`]: [`p(95)<${P95_MS}`, `p(99)<${P99_MS}`],
  };
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

function refreshSetupData() {
  const data = {
    accessToken: ACCESS_TOKEN,
    refreshToken: REFRESH_TOKEN,
    refreshTokens: REFRESH_TOKENS,
  };

  if (data.refreshTokens.length > 0 || data.refreshToken) {
    return data;
  }

  if (LOGIN_EMAIL && LOGIN_PASSWORD) {
    const tokenPair = loginOnce();
    return {
      accessToken: tokenPair.accessToken,
      refreshToken: tokenPair.refreshToken,
      refreshTokens: [],
    };
  }

  return data;
}

function currentRefreshToken(data) {
  if (data.refreshTokens && data.refreshTokens.length > 0) {
    return data.refreshTokens[(__VU - 1) % data.refreshTokens.length];
  }
  return data.refreshToken;
}

function issueRefreshTokenForThisIteration() {
  const response = http.post(
    `${BASE_URL}/v1/auth/login/email-password`,
    JSON.stringify({ email: LOGIN_EMAIL, password: LOGIN_PASSWORD }),
    jsonParams('', { endpoint: 'auth.login', profile: 'refresh-setup' })
  );

  check(response, {
    'refresh setup login status is 200': (res) => res.status === 200,
  });

  if (response.status !== 200) {
    throw new Error(`refresh setup login failed with status ${response.status}`);
  }

  return response.json().refreshToken;
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
    ...(hasScenario('login') ? scenarioThresholds('login') : {}),
    ...(hasScenario('read') ? scenarioThresholds('auth-read') : {}),
    ...(hasScenario('refresh') ? scenarioThresholds('refresh') : {}),
  },
};

export function setup() {
  if (!hasScenario('read') && !hasScenario('refresh')) {
    return {
      accessToken: ACCESS_TOKEN,
      refreshToken: REFRESH_TOKEN,
      refreshTokens: REFRESH_TOKENS,
    };
  }

  if (hasScenario('refresh') && !hasScenario('read')) {
    return refreshSetupData();
  }

  if (ACCESS_TOKEN) {
    return {
      accessToken: ACCESS_TOKEN,
      refreshToken: REFRESH_TOKEN,
      refreshTokens: REFRESH_TOKENS,
    };
  }

  const tokenPair = loginOnce();
  return {
    accessToken: tokenPair.accessToken,
    refreshToken: tokenPair.refreshToken,
    refreshTokens: REFRESH_TOKENS,
  };
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
  let refreshToken = currentRefreshToken(data);
  if (!refreshToken && LOGIN_EMAIL && LOGIN_PASSWORD) {
    refreshToken = issueRefreshTokenForThisIteration();
  }

  if (!refreshToken) {
    throw new Error(
      'No refresh token available. Provide REFRESH_TOKEN, REFRESH_TOKENS, or LOGIN_EMAIL/LOGIN_PASSWORD.'
    );
  }

  const response = http.post(
    `${BASE_URL}/v1/auth/refresh-token`,
    JSON.stringify({ refreshToken }),
    jsonParams('', { endpoint: 'auth.refresh', profile: 'refresh' })
  );

  check(response, {
    'refresh status is 200': (res) => res.status === 200,
  });
}
