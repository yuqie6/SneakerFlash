import http from 'k6/http';
import { check, fail } from 'k6';
import { Rate } from 'k6/metrics';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8000/api/v1';
const USER_PREFIX = __ENV.USER_PREFIX || 'perf_u';
const USER_COUNT = Number(__ENV.USER_COUNT || '3000');
const USER_PASSWORD = __ENV.USER_PASSWORD || 'PerfTest#123';
const USER_BATCH = Number(__ENV.USER_BATCH || '200');
const START_DELAY_SEC = Number(__ENV.START_DELAY_SEC || '5');
const PRODUCT_STOCK = Number(__ENV.PRODUCT_STOCK || '500');
const PRODUCT_PRICE = Number(__ENV.PRODUCT_PRICE || '1999');

const successRate = new Rate('seckill_success_rate');
const businessFailRate = new Rate('seckill_business_fail_rate');
const httpErrorRate = new Rate('http_error_rate');

export const options = {
  scenarios: {
    seckill: {
      executor: 'constant-arrival-rate',
      rate: Number(__ENV.RATE || '200'),
      timeUnit: '1s',
      duration: __ENV.DURATION || '30s',
      preAllocatedVUs: Number(__ENV.PRE_ALLOCATED_VUS || __ENV.RATE || '200'),
      maxVUs: Number(__ENV.MAX_VUS || Math.max(Number(__ENV.RATE || '200') * 2, 400)),
    },
  },
  setupTimeout: __ENV.SETUP_TIMEOUT || '5m',
  thresholds: {
    http_error_rate: ['rate<0.02'],
    http_req_duration: ['p(95)<800'],
  },
};

function postJson(path, payload, headers = {}) {
  return http.post(`${BASE_URL}${path}`, JSON.stringify(payload), {
    headers: {
      'Content-Type': 'application/json',
      ...headers,
    },
  });
}

function createProduct(token, startTime) {
  const uniqueName = `perf-sneaker-${Date.now()}`;
  const res = postJson(
    '/products',
    {
      name: uniqueName,
      price: PRODUCT_PRICE,
      stock: PRODUCT_STOCK,
      start_time: startTime,
      image: '',
    },
    { Authorization: `Bearer ${token}` },
  );

  const body = safeJson(res);
  if (!(res.status === 200 && body?.code === 200 && body.data?.id)) {
    fail(`创建商品失败: status=${res.status} body=${res.body}`);
  }
  return body.data.id;
}

function safeJson(res) {
  try {
    return res.json();
  } catch (err) {
    return null;
  }
}

function batchLogin(userNames) {
  const tokens = new Map();
  const missing = [];
  for (let i = 0; i < userNames.length; i += USER_BATCH) {
    const slice = userNames.slice(i, i + USER_BATCH);
    const responses = http.batch(
      slice.map(name => [
        'POST',
        `${BASE_URL}/login`,
        JSON.stringify({ user_name: name, user_password: USER_PASSWORD }),
        { headers: { 'Content-Type': 'application/json' } },
      ]),
    );

    responses.forEach((res, idx) => {
      const body = safeJson(res);
      if (res.status === 200 && body?.code === 200 && body.data?.access_token) {
        tokens.set(slice[idx], body.data.access_token);
      } else {
        missing.push(slice[idx]);
      }
    });
  }
  return { tokens, missing };
}

function batchRegister(userNames) {
  for (let i = 0; i < userNames.length; i += USER_BATCH) {
    const slice = userNames.slice(i, i + USER_BATCH);
    const responses = http.batch(
      slice.map(name => [
        'POST',
        `${BASE_URL}/register`,
        JSON.stringify({ user_name: name, user_password: USER_PASSWORD }),
        { headers: { 'Content-Type': 'application/json' } },
      ]),
    );
    responses.forEach((res, idx) => {
      const body = safeJson(res);
      const ok = res.status === 200 || body?.code === 10001;
      if (!ok) {
        fail(`注册失败: user=${slice[idx]} status=${res.status} body=${res.body}`);
      }
    });
  }
}

export function setup() {
  const startTime = new Date(Date.now() + START_DELAY_SEC * 1000).toISOString();
  const userNames = Array.from({ length: USER_COUNT }, (_, i) => `${USER_PREFIX}_${i}`);

  const loginResult = batchLogin(userNames);
  if (loginResult.missing.length > 0) {
    batchRegister(loginResult.missing);
    const retry = batchLogin(loginResult.missing);
    retry.tokens.forEach((token, name) => loginResult.tokens.set(name, token));
  }

  const tokens = userNames.map(name => {
    const token = loginResult.tokens.get(name);
    if (!token) {
      fail(`无法获取 token: user=${name}`);
    }
    return token;
  });

  const productId = createProduct(tokens[0], startTime);
  return { tokens, productId };
}

export default function run(data) {
  const token = data.tokens[(__ITER || 0) % data.tokens.length];
  const res = postJson('/seckill', { product_id: data.productId }, { Authorization: `Bearer ${token}` });
  const body = safeJson(res);
  const ok = res.status === 200 && body?.code === 200;

  successRate.add(ok);
  businessFailRate.add(res.status === 200 && body?.code !== 200);
  httpErrorRate.add(res.status >= 400);

  check(res, {
    'http 200': r => r.status === 200,
  });
}
