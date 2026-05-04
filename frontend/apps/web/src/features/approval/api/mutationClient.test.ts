import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { AUTH_EXPIRED_EVENT } from '../../../lib/api/authBus';
import { ApprovalError, mutate } from './mutationClient';

describe('mutationClient', () => {
  afterEach(() => vi.restoreAllMocks());

  describe('ApprovalError', () => {
    it('is instance of ApprovalError and has correct fields', () => {
      const err = new ApprovalError('authn.expired', 401, 'Sessão expirada');
      expect(err).toBeInstanceOf(ApprovalError);
      expect(err.code).toBe('authn.expired');
      expect(err.status).toBe(401);
      expect(err.message).toBe('Sessão expirada');
      expect(err.name).toBe('ApprovalError');
    });
  });

  // E4: 401 dispatches auth:expired event — no toast, throws ApprovalError
  describe('401 handling', () => {
    it('E4 — dispatches auth:expired event and throws ApprovalError', async () => {
      vi.spyOn(global, 'fetch').mockResolvedValue(
        new Response(JSON.stringify({}), { status: 401 }),
      );

      const listener = vi.fn();
      window.addEventListener(AUTH_EXPIRED_EVENT, listener);

      await expect(mutate('POST', '/api/test', {})).rejects.toBeInstanceOf(ApprovalError);
      expect(listener).toHaveBeenCalledTimes(1);

      window.removeEventListener(AUTH_EXPIRED_EVENT, listener);
    });

    it('E4 — thrown error has status 401', async () => {
      vi.spyOn(global, 'fetch').mockResolvedValue(
        new Response(JSON.stringify({}), { status: 401 }),
      );

      let caught: ApprovalError | undefined;
      try {
        await mutate('POST', '/api/test', {});
      } catch (e) {
        caught = e as ApprovalError;
      }
      expect(caught?.status).toBe(401);
    });
  });

  // E2: 403 throws ApprovalError with backend code — no generic toast
  describe('403 handling', () => {
    it('E2 — 403 sod.submitter_cannot_sign throws ApprovalError with correct code', async () => {
      vi.spyOn(global, 'fetch').mockImplementation(() =>
        Promise.resolve(
          new Response(
            JSON.stringify({ error: { code: 'sod.submitter_cannot_sign', message: 'forbidden' } }),
            { status: 403 },
          ),
        ),
      );

      let caught: ApprovalError | undefined;
      try {
        await mutate('POST', '/api/test', {});
      } catch (e) {
        caught = e as ApprovalError;
      }
      expect(caught).toBeInstanceOf(ApprovalError);
      expect(caught?.code).toBe('sod.submitter_cannot_sign');
      expect(caught?.status).toBe(403);
    });

    it('E2 — 403 sod.cross_stage_duplicate throws ApprovalError with correct code', async () => {
      vi.spyOn(global, 'fetch').mockImplementation(() =>
        Promise.resolve(
          new Response(
            JSON.stringify({ error: { code: 'sod.cross_stage_duplicate', message: 'forbidden' } }),
            { status: 403 },
          ),
        ),
      );

      let caught: ApprovalError | undefined;
      try {
        await mutate('POST', '/api/test', {});
      } catch (e) {
        caught = e as ApprovalError;
      }
      expect(caught).toBeInstanceOf(ApprovalError);
      expect(caught?.code).toBe('sod.cross_stage_duplicate');
    });
  });

  describe('412 handling', () => {
    it('throws ApprovalError with conflict.stale and status 412', async () => {
      vi.spyOn(global, 'fetch').mockImplementation(() =>
        Promise.resolve(
          new Response(
            JSON.stringify({ error: { code: 'conflict.stale' } }),
            { status: 412 },
          ),
        ),
      );

      let caught: ApprovalError | undefined;
      try {
        await mutate('POST', '/api/test', {});
      } catch (e) {
        caught = e as ApprovalError;
      }
      expect(caught).toBeInstanceOf(ApprovalError);
      expect(caught?.status).toBe(412);
      expect(caught?.code).toBe('conflict.stale');
    });
  });

  describe('429 handling', () => {
    it('throws ApprovalError with authn.rate_limited', async () => {
      vi.spyOn(global, 'fetch').mockResolvedValue(
        new Response(JSON.stringify({}), { status: 429 }),
      );

      let caught: ApprovalError | undefined;
      try {
        await mutate('POST', '/api/test', {});
      } catch (e) {
        caught = e as ApprovalError;
      }
      expect(caught).toBeInstanceOf(ApprovalError);
      expect(caught?.code).toBe('authn.rate_limited');
      expect(caught?.status).toBe(429);
    });
  });
});
