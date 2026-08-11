import { describe, it, expect } from 'vitest';
import JSZip from 'jszip';
import { makeTemplateProcessor, RenderError, type EigenpalEngine } from '../index';

// The adapter now normalizes the engine's returned buffer as a ZIP (that is
// the fix under test elsewhere), so the fake engine here must return a
// buffer JSZip can actually load — a real zip, not arbitrary bytes.
async function miniZipBuffer(): Promise<ArrayBuffer> {
  const zip = new JSZip();
  zip.file('word/document.xml', '<w:document/>');
  return zip.generateAsync({ type: 'arraybuffer' });
}

async function okResult() {
  return {
    buffer: await miniZipBuffer(),
    replacedVariables: ['name'],
    unreplacedVariables: ['missing'],
    warnings: [{ message: 'soft', variable: 'x', type: 'render' as const }],
  };
}

function engineReturning(result: unknown): EigenpalEngine {
  return { processTemplateDetailed: () => result as never };
}
function engineThrowing(err: unknown): EigenpalEngine {
  return { processTemplateDetailed: () => { throw err; } };
}

describe('TemplateProcessor', () => {
  it('maps a successful result to RenderResult (buffer + replaced/unreplaced + warnings)', async () => {
    const tp = makeTemplateProcessor(engineReturning(await okResult()));
    const r = await tp.processTemplate(new ArrayBuffer(0), { name: 'A' });
    const roundTripped = await JSZip.loadAsync(r.buffer);
    expect(await roundTripped.file('word/document.xml')?.async('string')).toBe('<w:document/>');
    expect(r.replacedVariables).toEqual(['name']);
    expect(r.unreplacedVariables).toEqual(['missing']);
    expect(r.warnings).toEqual([{ kind: 'template_render', message: 'soft', variable: 'x' }]);
  });

  it('translates a thrown parse error to RenderError(template_parse)', async () => {
    const tp = makeTemplateProcessor(engineThrowing({ type: 'parse', message: 'bad zip' }));
    await expect(tp.processTemplate(new ArrayBuffer(0), {})).rejects.toMatchObject({
      kind: 'template_parse',
      message: expect.stringContaining('bad zip'),
    });
  });

  it('translates type undefined → undefined_variable and carries the variable', async () => {
    const tp = makeTemplateProcessor(engineThrowing({ type: 'undefined', message: 'no var', variable: 'foo' }));
    const e: RenderError = await tp.processTemplate(new ArrayBuffer(0), {}).catch((x) => x);
    expect(e.kind).toBe('undefined_variable');
    expect(e.variable).toBe('foo');
  });

  it('translates an unrecognized throw (no .type) to RenderError(unknown) and keeps cause', async () => {
    const original = new Error('boom');
    const tp = makeTemplateProcessor(engineThrowing(original));
    const e: RenderError = await tp.processTemplate(new ArrayBuffer(0), {}).catch((x) => x);
    expect(e.kind).toBe('unknown');
    expect(e.cause).toBe(original);
  });

  it('translates type render → template_render', async () => {
    const tp = makeTemplateProcessor(engineThrowing({ type: 'render', message: 'render fail' }));
    const e: RenderError = await tp.processTemplate(new ArrayBuffer(0), {}).catch((x) => x);
    expect(e.kind).toBe('template_render');
  });

  it('preserves the archive comment and per-entry comments through normalization', async () => {
    // Normalization re-serializes the archive from scratch, so anything it does
    // not explicitly copy across is silently dropped. Comments are metadata of
    // the frozen artifact; losing them is data loss even though determinism
    // would hold either way (both renders would lose them identically).
    const zip = new JSZip();
    zip.file('word/document.xml', '<w:document/>', { comment: 'entry-comment' });
    (zip as unknown as { comment: string }).comment = 'archive-comment';
    const buffer = await zip.generateAsync({ type: 'arraybuffer' });

    const tp = makeTemplateProcessor(
      engineReturning({ buffer, replacedVariables: [], unreplacedVariables: [], warnings: [] }),
    );
    const r = await tp.processTemplate(new ArrayBuffer(0), {});

    const out = await JSZip.loadAsync(r.buffer);
    expect((out as unknown as { comment?: string }).comment).toBe('archive-comment');
    expect(out.file('word/document.xml')?.comment).toBe('entry-comment');
  });

  it('translates a malformed (non-ZIP) engine buffer to RenderError(unknown) rather than throwing raw', async () => {
    const tp = makeTemplateProcessor(
      engineReturning({ buffer: new Uint8Array([1, 2, 3]).buffer, replacedVariables: [], unreplacedVariables: [], warnings: [] }),
    );
    const e: RenderError = await tp.processTemplate(new ArrayBuffer(0), {}).catch((x) => x);
    expect(e).toBeInstanceOf(RenderError);
    expect(e.kind).toBe('unknown');
  });
});
