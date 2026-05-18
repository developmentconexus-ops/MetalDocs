import { chromium } from '@playwright/test';
const browser = await chromium.launch({ headless: true });
const page = await browser.newPage();
await page.goto('http://localhost:4173/login', { waitUntil: 'load', timeout: 30000 });
const result = await page.evaluate(async () => {
  const ctl = new AbortController();
  const t = setTimeout(() => ctl.abort('timeout'), 5000);
  try {
    const res = await fetch('http://127.0.0.1:9000/', { method: 'GET', signal: ctl.signal });
    clearTimeout(t);
    return { ok: res.ok, status: res.status, text: await res.text() };
  } catch (error) {
    clearTimeout(t);
    return { error: String(error) };
  }
});
console.log(JSON.stringify(result));
await browser.close();
