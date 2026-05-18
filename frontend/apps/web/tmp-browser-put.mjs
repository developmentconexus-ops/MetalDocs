import { chromium } from '@playwright/test';
const url = process.argv[2];
const browser = await chromium.launch({ headless: true });
const page = await browser.newPage();
await page.goto('http://localhost:4173/login', { waitUntil: 'load', timeout: 30000 });
const result = await page.evaluate(async (uploadUrl) => {
  try {
    const res = await fetch(uploadUrl, {
      method: 'PUT',
      headers: { 'content-type': 'application/vnd.openxmlformats-officedocument.wordprocessingml.document' },
      body: new Uint8Array([104,101,108,108,111,45,98,114,111,119,115,101,114]),
    });
    return { ok: res.ok, status: res.status, statusText: res.statusText, text: await res.text() };
  } catch (error) {
    return { error: String(error) };
  }
}, url);
console.log(JSON.stringify(result));
await browser.close();
