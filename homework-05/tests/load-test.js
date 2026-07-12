import http from 'k6/http';
import { sleep, check } from 'k6';

const BASE_URL = 'http://arch.homework';

export const options = {
    stages: [
        { duration: '1m', target: 20 },
        { duration: '8m', target: 20 },
        { duration: '1m', target: 0 },
    ],
};

export default function () {
    let list = http.get(`${BASE_URL}/api/v1/users`);
    check(list, { 'list status is 200': r => r.status === 200 });

    let payload = JSON.stringify({
        username: `user-${__VU}-${Date.now()}`,
        firstName: 'Test',
        lastName: 'User',
        email: `user-${__VU}-${Date.now()}@test.local`,
        phone: '+79990000000'
    });

    let create = http.post(`${BASE_URL}/api/v1/users`, payload, {
        headers: { 'Content-Type': 'application/json' },
    });

    check(create, { 'create status is 201': r => r.status === 201 });

    http.get(`${BASE_URL}/api/v1/users/not-found-${Date.now()}`);

    sleep(1);
}