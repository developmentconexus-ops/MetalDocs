import { describe, expect, test } from 'vitest';
import JSZip from 'jszip';
import { fanout } from '../fanout.js';
import { detectTokens, isValidIdent } from '@metaldocs/shared-tokens';

// Probe set: snake_case-valid + charset-invalid shapes the vendor may still act on.
const VALID = ['a_b', 'ABC', 'x1', '_y'];
const INVALID = ['a.b', 'a-b', '1n', 'a b'];
const PROBES = [...VALID, ...INVALID];
const PROBE_TEXT = PROBES.map((p) => `{${p}}`).join(' ');

async function buildProbeDocx(): Promise<Uint8Array> {
  const zip = new JSZip();
  const docXml = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body><w:p><w:r><w:t xml:space="preserve">${PROBE_TEXT}</w:t></w:r></w:p><w:sectPr/></w:body>
</w:document>`;
  zip.file('[Content_Types].xml', `<?xml version="1.0"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/></Types>`);
  zip.file('_rels/.rels', `<?xml version="1.0"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/></Relationships>`);
  zip.file('word/document.xml', docXml);
  zip.file('word/_rels/document.xml.rels', `<?xml version="1.0"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"></Relationships>`);
  return zip.generateAsync({ type: 'uint8array' });
}

describe('token grammar parity: detector ⇄ real vendor freeze', () => {
  test('freeze does not throw on charset-invalid tags', async () => {
    const tpl = await buildProbeDocx();
    const values = Object.fromEntries(PROBES.map((p) => [p, `V_${p}`]));
    await expect(
      fanout({
        bodyDocx: tpl,
        placeholderValues: values,
        compositionConfig: { header_sub_blocks: [], footer_sub_blocks: [], sub_block_params: {} },
        resolvedValues: {},
      }),
    ).resolves.toBeTruthy();
  });

  test('HARD GATE: detector surfaces every token freeze acts on (no silent loss)', async () => {
    const tpl = await buildProbeDocx();
    const values = Object.fromEntries(PROBES.map((p) => [p, `V_${p}`]));
    const result = await fanout({
      bodyDocx: tpl,
      placeholderValues: values,
      compositionConfig: { header_sub_blocks: [], footer_sub_blocks: [], sub_block_params: {} },
      resolvedValues: {},
    });

    const detected = new Set(detectTokens(PROBE_TEXT).map((d) => d.name));

    // Exclude synthetic composition keys injected by fanout (not user-visible tokens).
    // These are internal plumbing keys that fanout always adds to the variables map;
    // they are not part of the token grammar seen by document authors.
    const SYNTHETIC_KEYS = new Set(['__header_composition__', '__footer_composition__']);
    const touched = new Set(
      [...result.replacedVars, ...result.unreplacedVars].filter((k) => !SYNTHETIC_KEYS.has(k)),
    );

    // (1) Anything freeze touched must be detected. If this fails, the vendor names
    //     a tag differently than its literal text (e.g. nested-path) — STOP and report
    //     it as an architecture finding (spec §4.3); do NOT weaken this assertion.
    for (const name of touched) {
      expect(detected.has(name), `freeze touched "${name}" but detector missed it`).toBe(true);
    }

    // (2) All snake_case-valid probes substitute.
    for (const v of VALID) {
      expect(result.replacedVars, `valid probe "${v}" should substitute`).toContain(v);
    }

    // (3) Every touched token that is charset-invalid is flagged invalid by the detector.
    for (const name of touched) {
      if (!isValidIdent(name)) {
        const d = detectTokens(`{${name}}`)[0];
        expect(d?.valid, `invalid token "${name}" must be flagged valid:false`).toBe(false);
      }
    }
  });
});
