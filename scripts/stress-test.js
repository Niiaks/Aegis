import http from 'k6/http';
import { check, sleep } from 'k6';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

export const options = {
    stages: [
        { duration: '30s', target: 200 }, // Ramp up to 200 users
        { duration: '1m', target: 200 },  // Stay at 200 users
        { duration: '30s', target: 0 },   // Ramp down
    ],
    thresholds: {
        http_req_duration: ['p(95)<500'], // 95% of requests must complete below 500ms
        http_req_failed: ['rate<0.01'],   // Error rate should be less than 1%
    },
};

const BASE_URL = 'http://localhost:8080/api/v1';

export default function () {
    const idempotencyKey = uuidv4();
    const payload = JSON.stringify({
        email: `test-user-${__VU}@example.com`,
        amount: 10000, // 100.00 in minor units
        currency: 'USD',
        status: 'pending',
        type: 'payment_intent',
        metadata: {
            user_id: '6855c0c7-9ffa-489c-a9df-d71b049b8ea7',
            custom_field: 'stress-test',
        },
    });

    const params = {
        headers: {
            'Content-Type': 'application/json',
            'Idempotency-Key': idempotencyKey,
        },
    };

    const res = http.post(`${BASE_URL}/transactions/payment-intent`, payload, params);

    check(res, {
        'is status 200': (r) => r.status === 200,
        'has payment url': (r) => r.body && r.json().data && r.json().data.authorization_url !== undefined,
    });

    // sleep(1); // Removed for max RPS test
}
