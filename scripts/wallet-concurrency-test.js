import http from 'k6/http';
import { check, sleep } from 'k6';
import crypto from 'k6/crypto';

export const options = {
    vus: 20, // 20 concurrent workers hitting the SAME wallet
    duration: '60s',
};

const BASE_URL = 'http://localhost:8080/api/v1';
const SECRET_KEY = __ENV.AEGIS_PAYSTACK_SECRET_KEY || 'sk_test_123';
const TARGET_USER_ID = '6855c0c7-9ffa-489c-a9df-d71b049b8ea7'; // Targeting a known user

export default function () {
    const payload = JSON.stringify({
        event: 'charge.success',
        data: {
            id: Math.floor(Math.random() * 1000000),
            amount: 5000,
            currency: 'USD',
            reference: `ref_${Math.random()}`,
            metadata: {
                user_id: TARGET_USER_ID,
                transaction_id: `tx_${Math.random()}`,
            },
        },
    });

    const signature = crypto.hmac('sha512', SECRET_KEY, payload, 'hex');

    const params = {
        headers: {
            'Content-Type': 'application/json',
            'x-paystack-signature': signature,
        },
    };

    const res = http.post(`${BASE_URL}/paystack/webhook`, payload, params);

    check(res, {
        'is status 200': (r) => r.status === 200,
    });

    sleep(0.1); // High frequency
}
