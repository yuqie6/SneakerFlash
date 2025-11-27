import http from 'k6/http';
import { check, fail } from 'k6';
import exec from 'k6/execution';
import { Rate } from 'k6/metrics';
import { SharedArray } from 'k6/data';
import papa from 'https://jslib.k6.io/papaparse/5.1.1/index.js';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8000/api/v1';
const USER_PREFIX = __ENV.USER_PREFIX || 'perf_u';
const USER_COUNT = Number(__ENV.USER_COUNT || '3000');
const USER_PASSWORD = __ENV.USER_PASSWORD || 'PerfTest#123';
const USER_BATCH = Number(__ENV.USER_BATCH || '200');
const START_DELAY_SEC = Number(__ENV.START_DELAY_SEC || '3');
let failLogRemain = Number(__ENV.FAIL_LOG_LIMIT || '20'); // 仅 VU1 打印有限条失败日志，避免刷屏
const TOKEN_STRATEGY = (__ENV.TOKEN_STRATEGY || 'round_robin').toLowerCase(); // round_robin | random
const PRODUCT_STOCK = Number(__ENV.PRODUCT_STOCK || '500');
const PRODUCT_PRICE = Number(__ENV.PRODUCT_PRICE || '1999');
const TOKEN_CSV = __ENV.TOKEN_CSV || '';
const USE_RAMP = (__ENV.USE_RAMP || '').toLowerCase() === 'true';
const RAMP_STAGES = __ENV.RAMP_STAGES || '30s:800,30s:1200,30s:1500';
const START_RATE = Number(__ENV.START_RATE || __ENV.RATE || '200');

const successRate = new Rate('seckill_success_rate');
const businessFailRate = new Rate('seckill_business_fail_rate');
const httpErrorRate = new Rate('http_error_rate');

const csvTokens = TOKEN_CSV
  ? new SharedArray('tokens-csv', () => {
      const text = open(TOKEN_CSV);
      const parsed = papa.parse(text.trim(), { header: true });
      const tokens = parsed.data
        .map(row => row.token || row.Token || row.TOKEN || row.access_token || row.AccessToken || row.ACCESS_TOKEN)
        .filter(t => typeof t === 'string' && t.length > 0)
        .map(t => t.trim());
      if (tokens.length === 0) {
        throw new Error('token.csv 未解析到 token 列，请确保包含表头 token 或 access_token');
      }
      return tokens;
    })
  : null;

function parseRampStages(raw) {
  return raw
    .split(',')
    .map(s => s.trim())
    .filter(Boolean)
    .map(item => {
      const [duration, target] = item.split(':');
      return { duration, target: Number(target) };
    });
}

const commonScenario = {
  timeUnit: '1s',
  preAllocatedVUs: Number(__ENV.PRE_ALLOCATED_VUS || __ENV.RATE || '200'),
  maxVUs: Number(__ENV.MAX_VUS || Math.max(Number(__ENV.RATE || '200') * 2, 400)),
};

export const options = USE_RAMP
  ? {
      scenarios: {
        seckill: {
          executor: 'ramping-arrival-rate',
          startRate: START_RATE,
          stages: parseRampStages(RAMP_STAGES),
          ...commonScenario,
        },
      },
      setupTimeout: __ENV.SETUP_TIMEOUT || '5m',
      thresholds: {
        http_error_rate: ['rate<0.02'],
        http_req_duration: ['p(95)<800'],
      },
    }
  : {
      scenarios: {
        seckill: {
          executor: 'constant-arrival-rate',
          rate: Number(__ENV.RATE || '200'),
          duration: __ENV.DURATION || '30s',
          ...commonScenario,
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

function getJson(path, headers = {}) {
  return http.get(`${BASE_URL}${path}`, { headers });
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
  console.log(`秒杀开始时间: ${startTime}（延迟 ${START_DELAY_SEC}s）`);

  // 若提供 TOKEN_CSV，则直接读取 token 列，跳过注册/登录
  if (TOKEN_CSV) {
    const productId = createProduct(csvTokens[0], startTime);
    // 记录初始库存
    const detailRes = getJson(`/product/${productId}`);
    const detailBody = safeJson(detailRes);
    if (detailRes.status === 200 && detailBody?.data) {
      console.log(`初始库存: product_id=${productId} stock=${detailBody.data.stock}`);
    } else {
      console.log(`获取商品详情失败: status=${detailRes.status} body=${detailRes.body}`);
    }
    return { productId };
  }

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

  // 记录初始库存，便于对比压测后库存变化（仅日志，不影响流程）
  const detailRes = getJson(`/product/${productId}`);
  const detailBody = safeJson(detailRes);
  if (detailRes.status === 200 && detailBody?.data) {
    console.log(`初始库存: product_id=${productId} stock=${detailBody.data.stock}`);
  } else {
    console.log(`获取商品详情失败: status=${detailRes.status} body=${detailRes.body}`);
  }

  return { tokens, productId };
}

export default function run(data) {
  const tokenList = TOKEN_CSV ? csvTokens : data.tokens;
  if (!tokenList || tokenList.length === 0) {
    fail('token 列表为空，请检查 TOKEN_CSV 或注册/登录流程');
  }
  const token =
    TOKEN_STRATEGY === 'random'
      ? tokenList[Math.floor(Math.random() * tokenList.length)]
      : tokenList[exec.scenario.iterationInTest % tokenList.length];
  const res = postJson('/seckill', { product_id: data.productId }, { Authorization: `Bearer ${token}` });
  const body = safeJson(res);
  const ok = res.status === 200 && body?.code === 200;

  successRate.add(ok);
  businessFailRate.add(res.status === 200 && body?.code !== 200);
  httpErrorRate.add(res.status >= 400);

  // VU1 限量打印业务失败，帮助定位失败原因
  if (!ok && __VU === 1 && failLogRemain > 0) {
    const tokenHint = token ? `${token.slice(0, 12)}...${token.slice(-6)}` : 'empty';
    console.log(`业务失败: status=${res.status} body=${res.body} token=${tokenHint}`);
    failLogRemain -= 1;
  }

  check(res, {
    'http 200': r => r.status === 200,
  });
}
