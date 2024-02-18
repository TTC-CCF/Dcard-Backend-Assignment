import http from 'k6/http';
import { randomItem, randomIntBetween } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';
import { check, sleep } from 'k6';

const gender = ['M', 'F'];
const country = ['TW', 'US', 'JP', 'KR', 'CN', 'HK', 'CA', 'UK', 'FR', 'DE', 'IT'];
const platform = ['ios', 'android', 'web'];

export const options = {
  stages: [
    { duration: '10s', target: 100 },
    { duration: '1m', target: 100 },
    { duration: '10s', target: 0 },
  ],
};

const host = 'http://localhost:3001/api/v1/ad';

export default function() {
  const randomAge = randomIntBetween(1, 100);
  const randomCountry = randomItem(country);
  const randomGender = randomItem(gender);
  const randomPlatform = randomItem(platform);

  const urls = [
    host,
    `${host}?age=${randomAge}`,
    `${host}?country=${randomCountry}`,
    `${host}?gender=${randomGender}`,
    `${host}?platform=${randomPlatform}`,
    `${host}?age=${randomAge}&country=${randomCountry}`,
    `${host}?age=${randomAge}&gender=${randomGender}`,
    `${host}?age=${randomAge}&platform=${randomPlatform}`,
    `${host}?country=${randomCountry}&gender=${randomGender}`,
    `${host}?country=${randomCountry}&platform=${randomPlatform}`,
    `${host}?gender=${randomGender}&platform=${randomPlatform}`,
    `${host}?age=${randomAge}&country=${randomCountry}&gender=${randomGender}`,
    `${host}?age=${randomAge}&country=${randomCountry}&platform=${randomPlatform}`,
    `${host}?age=${randomAge}&gender=${randomGender}&platform=${randomPlatform}`,
    `${host}?country=${randomCountry}&gender=${randomGender}&platform=${randomPlatform}`,
    `${host}?age=${randomAge}&country=${randomCountry}&gender=${randomGender}&platform=${randomPlatform}`,
  ]

  const randomUrl = randomItem(urls);
  const res = http.get(randomUrl);

  check(res, {
    'Post status is 200': (r) => res.status === 200,
  });
}
