import { describe, expect, test } from 'vitest';
import JSZip from 'jszip';
import { fanout } from '../fanout.js';

async function buildTemplateDocx(bodyInnerXml: string): Promise<Uint8Array> {
  const zip = new JSZip();
  const docXml = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body>${bodyInnerXml}<w:sectPr/></w:body>
</w:document>`;
  const contentTypes = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
</Types>`;
  const rels = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`;
  const wordRels = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"></Relationships>`;

  zip.file('[Content_Types].xml', contentTypes);
  zip.file('_rels/.rels', rels);
  zip.file('word/document.xml', docXml);
  zip.file('word/_rels/document.xml.rels', wordRels);

  const buf = await zip.generateAsync({ type: 'uint8array' });
  return buf;
}

async function extractDocumentXml(docx: Uint8Array): Promise<string> {
  const zip = await JSZip.loadAsync(docx);
  const file = zip.file('word/document.xml');
  if (!file) throw new Error('missing document.xml');
  return file.async('string');
}

/** Every entry's stored ZIP date, keyed by path — used to assert the
 * normalized timestamp directly rather than only its downstream hash. */
async function zipEntryDates(docx: Uint8Array): Promise<Record<string, string>> {
  const zip = await JSZip.loadAsync(docx);
  const out: Record<string, string> = {};
  zip.forEach((relativePath, entry) => {
    out[relativePath] = entry.date.toISOString();
  });
  return out;
}

/**
 * Freezes `new Date()` (zero-arg construction) and `Date.now()` to a fixed
 * instant until the returned restore function is called, without touching
 * setTimeout/setInterval
 * or any other timer. This targets exactly the wall-clock call sites this
 * suite needs to control — docxtemplater/PizZip's and jszip's own
 * `o.date = o.date || new Date()` fallback (see index.ts's
 * normalizeZipTimestamps doc comment for the full chain) — while leaving
 * `new Date(explicitArgs)` untouched.
 *
 * A vitest `vi.useFakeTimers()` was deliberately NOT used here: jszip's
 * internal stream scheduling calls `setImmediate` (utils.js), which fake
 * timers intercept too, and would hang loadAsync/generateAsync mid-test
 * unless manually advanced. Freezing only the Date global sidesteps that
 * entirely.
 */
function freezeDate(iso: string): () => void {
  const RealDate = globalThis.Date;
  const frozenMs = new RealDate(iso).getTime();
  const handler: ProxyHandler<DateConstructor> = {
    construct(target, args) {
      if (args.length === 0) {
        return new target(frozenMs);
      }
      return Reflect.construct(target, args);
    },
    get(target, prop, receiver) {
      if (prop === 'now') {
        return () => frozenMs;
      }
      return Reflect.get(target, prop, receiver);
    },
  };
  globalThis.Date = new Proxy(RealDate, handler) as DateConstructor;
  return () => {
    globalThis.Date = RealDate;
  };
}

describe('fanout', () => {
  test('returns buffer with stable sha256 contentHash', async () => {
    const body = `<w:p><w:r><w:t>Doc code: {doc_code}</w:t></w:r></w:p>`;
    const tpl = await buildTemplateDocx(body);

    const result = await fanout({
      bodyDocx: tpl,
      placeholderValues: { doc_code: 'ABC-001' },

      compositionConfig: {
        header_sub_blocks: [],
        footer_sub_blocks: [],
        sub_block_params: {},
      },
      resolvedValues: {},
    });

    expect(result.contentHash).toMatch(/^[0-9a-f]{64}$/);
    expect(result.buffer.byteLength).toBeGreaterThan(0);

    const xml = await extractDocumentXml(result.buffer);
    expect(xml).toContain('ABC-001');
    expect(result.replacedVars).toContain('doc_code');
    expect(Array.isArray(result.unreplacedVars)).toBe(true);
  });

  test('renders header sub-blocks via resolvedValues', async () => {
    const body = `<w:p><w:r><w:t>{doc_code}</w:t></w:r></w:p>`;
    const tpl = await buildTemplateDocx(body);

    const result = await fanout({
      bodyDocx: tpl,
      placeholderValues: { doc_code: 'ABC-001' },

      compositionConfig: {
        header_sub_blocks: ['doc_header_standard'],
        footer_sub_blocks: [],
        sub_block_params: {},
      },
      resolvedValues: {
        title: 'Test Doc',
        doc_code: 'ABC-001',
        effective_date: '2026-04-23',
        revision_number: '1',
      },
    });

    expect(result.contentHash).toMatch(/^[0-9a-f]{64}$/);
    expect(result.buffer.byteLength).toBeGreaterThan(0);
  });

  test('renders byte-identical ZIPs across clock instants with normalized entry dates', async () => {
    const body = `<w:p><w:r><w:t>{doc_code}</w:t></w:r></w:p>`;
    const tpl = await buildTemplateDocx(body);
    const input = {
      bodyDocx: tpl,
      placeholderValues: { doc_code: 'STABLE-1' },

      compositionConfig: {
        header_sub_blocks: [],
        footer_sub_blocks: [],
        sub_block_params: {},
      },
      resolvedValues: {},
    };

    const renderAt = async (iso: string) => {
      const restoreDate = freezeDate(iso);
      try {
        return await fanout(input);
      } finally {
        restoreDate();
      }
    };

    const first = await renderAt('2026-01-01T00:00:00.000Z');
    const second = await renderAt('2026-01-01T00:00:04.000Z');

    expect(Array.from(first.buffer)).toEqual(Array.from(second.buffer));
    expect(first.contentHash).toBe(second.contentHash);

    const normalizedDate = '1980-01-01T00:00:00.000Z';
    for (const dates of [await zipEntryDates(first.buffer), await zipEntryDates(second.buffer)]) {
      expect(Object.values(dates)).not.toHaveLength(0);
      expect(Object.values(dates).every((date) => date === normalizedDate)).toBe(true);
    }
  });
});
