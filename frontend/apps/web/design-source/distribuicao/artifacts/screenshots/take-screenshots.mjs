import { chromium } from 'playwright';
import { fileURLToPath } from 'url';
import path from 'path';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const URL = 'http://localhost:4174/documents/PR-EHS-014/distribution';

const viewports = [
  { width: 1440, height: 900,  name: '1440-impl' },
  { width: 1024, height: 768,  name: '1024-impl' },
  { width: 375,  height: 812,  name: '375-impl'  },
];

const browser = await chromium.launch();

for (const vp of viewports) {
  const page = await browser.newPage();
  await page.setViewportSize({ width: vp.width, height: vp.height });
  await page.goto(URL, { waitUntil: 'networkidle' });
  // Scroll to top
  await page.evaluate(() => window.scrollTo(0, 0));
  await page.waitForTimeout(500);
  const outPath = path.join(__dirname, `${vp.name}.png`);
  await page.screenshot({ path: outPath, fullPage: false });
  console.log(`Saved ${outPath}`);
  await page.close();
}

await browser.close();
console.log('Done.');
