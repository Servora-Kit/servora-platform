import http from 'k6/http';
import { check } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://127.0.0.1:8000';
const MODE = __ENV.MODE || 'ramp';
const RATE = Number(__ENV.RATE || 200);
const DURATION = __ENV.DURATION || '3m';
const PRE_ALLOCATED_VUS = Number(__ENV.PRE_ALLOCATED_VUS || 50);
const MAX_VUS = Number(__ENV.MAX_VUS || 500);
const P95_MS = Number(__ENV.P95_MS || 200);
const P99_MS = Number(__ENV.P99_MS || 500);
const FAIL_RATE = Number(__ENV.FAIL_RATE || 0.001);

function buildScenario(name) {
  if (MODE === 'steady') {
    return {
      executor: 'constant-arrival-rate',
      exec: name,
      rate: RATE,
      timeUnit: '1s',
      duration: DURATION,
      preAllocatedVUs: PRE_ALLOCATED_VUS,
      maxVUs: MAX_VUS,
    };
  }

  return {
    executor: 'ramping-arrival-rate',
    exec: name,
    startRate: Math.max(1, Math.floor(RATE / 10)),
    timeUnit: '1s',
    preAllocatedVUs: PRE_ALLOCATED_VUS,
    maxVUs: MAX_VUS,
    stages: [
      { target: Math.max(10, Math.floor(RATE * 0.25)), duration: '1m' },
      { target: Math.max(20, Math.floor(RATE * 0.5)), duration: '1m' },
      { target: RATE, duration: '1m' },
      { target: RATE, duration: '1m' },
    ],
  };
}

export const options = {
  scenarios: {
    baseline: buildScenario('baseline'),
  },
  thresholds: {
    http_req_failed: [`rate<${FAIL_RATE}`],
    http_req_duration: [`p(95)<${P95_MS}`, `p(99)<${P99_MS}`],
  },
};

export function baseline() {
  const response = http.post(
    `${BASE_URL}/v1/test/test`,
    JSON.stringify({}),
    {
      headers: { 'Content-Type': 'application/json' },
      tags: { endpoint: 'test.test', profile: 'baseline' },
    }
  );

  check(response, {
    'baseline status is 200': (res) => res.status === 200,
  });
}
