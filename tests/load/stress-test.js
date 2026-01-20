import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const STRESS_DURATION = __ENV.STRESS_DURATION || '10s';
const WORKERS = __ENV.WORKERS || '2';

export const options = {
    stages: [
        { duration: '30s', target: 10 },
        { duration: '2m', target: 20 },
        { duration: '1m', target: 30 },
        { duration: '30s', target: 0 },
    ],
    thresholds: {
        http_req_failed: ['rate<0.1'],
    },
};

export function setup() {
    const healthRes = http.get(`${BASE_URL}/health`);
    check(healthRes, {
        'health check passed': (r) => r.status === 200,
    });

    console.log(`Target: ${BASE_URL}`);
    console.log(`Stress duration per request: ${STRESS_DURATION}`);
    console.log(`CPU workers per request: ${WORKERS}`);
    
    return { baseUrl: BASE_URL };
}

export default function (data) {
    const url = `${data.baseUrl}/stress?duration=${STRESS_DURATION}&workers=${WORKERS}`;
    
    const res = http.get(url, {
        timeout: '60s',
    });

    check(res, {
        'status is 200': (r) => r.status === 200,
    });

    sleep(1);
}
