import { ApiError } from './errors';
import { resolveErrorMessage } from './errors';

/**
 * Translate an `unknown` error (typically from TanStack Query) into a
 * user-facing string. Mirrors the inline triad used at every error callsite:
 *   ApiError → resolveErrorMessage(code, message)
 *   Error    → message
 *   *        → fallback
 */
export function resolveQueryError(err: unknown, fallback: string): string {
  if (err instanceof ApiError) {
    return resolveErrorMessage(err.code, err.message);
  }
  if (err instanceof Error) {
    return err.message;
  }
  return fallback;
}
