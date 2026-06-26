import { describe, it, expect } from 'vitest';
import { renderErrorToHttp } from '../../routes/fanout';
import { RenderError } from '@metaldocs/eigenpal-adapter';

describe('renderErrorToHttp', () => {
  it('parse errors are a 400 client/template defect', () => {
    const { status, body } = renderErrorToHttp(new RenderError('template_parse', 'bad zip'));
    expect(status).toBe(400);
    expect(body).toEqual({ error: 'render_failed', kind: 'template_parse', message: 'bad zip' });
  });

  it('undefined_variable is 400 and carries the variable', () => {
    const { status, body } = renderErrorToHttp(new RenderError('undefined_variable', 'no var', { variable: 'foo' }));
    expect(status).toBe(400);
    expect(body.variable).toBe('foo');
  });

  it('template_render is a 422 (well-formed input, engine could not render)', () => {
    expect(renderErrorToHttp(new RenderError('template_render', 'x')).status).toBe(422);
  });

  it('unknown is a 500 (retryable by the worker)', () => {
    expect(renderErrorToHttp(new RenderError('unknown', 'x')).status).toBe(500);
  });
});
