/**
 * ApiError — structured error from the MetalDocs API.
 * Created as a sub-plan 1 primitive; consumed by rename-rollback (E9) and other consumer-facing handlers.
 */
export class ApiError extends Error {
  constructor(
    public readonly code: string,
    public readonly status: number,
    message: string,
  ) {
    super(message);
    this.name = 'ApiError';
  }
}
