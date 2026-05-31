import { errorMessages } from './errorMessages';

export type FieldError = {
  field: string;
  code: string;
  message: string;
};

export type Problem = {
  type?: string;
  title: string;
  status: number;
  detail?: string;
  instance?: string;
  code: string;
  errors?: FieldError[];
};

const mergedMessages: Record<string, string> = errorMessages;

function isFieldError(value: unknown): value is FieldError {
  if (!value || typeof value !== 'object') return false;

  const candidate = value as Partial<FieldError>;
  return (
    typeof candidate.field === 'string' &&
    typeof candidate.code === 'string' &&
    typeof candidate.message === 'string'
  );
}

export function isProblem(value: unknown): value is Problem {
  if (!value || typeof value !== 'object') return false;

  const candidate = value as Partial<Problem>;

  if (
    typeof candidate.code !== 'string' ||
    typeof candidate.title !== 'string' ||
    typeof candidate.status !== 'number'
  ) {
    return false;
  }

  if (candidate.type !== undefined && typeof candidate.type !== 'string') return false;
  if (candidate.detail !== undefined && typeof candidate.detail !== 'string') return false;
  if (candidate.instance !== undefined && typeof candidate.instance !== 'string') return false;

  if (candidate.errors !== undefined) {
    if (!Array.isArray(candidate.errors)) return false;
    if (!candidate.errors.every((item) => isFieldError(item))) return false;
  }

  return true;
}

export async function parseProblem(res: Response): Promise<Problem | null> {
  const contentType = res.headers.get('content-type');
  if (!contentType) return null;

  const mediaType = contentType.split(';')[0]?.trim().toLowerCase();
  if (mediaType !== 'application/problem+json') return null;

  try {
    const body = await res.json();
    return isProblem(body) ? body : null;
  } catch {
    return null;
  }
}

export class ApiError extends Error {
  readonly status: number;
  readonly code: string;
  readonly title: string;
  readonly detail?: string;
  readonly details?: unknown;
  readonly instance?: string;
  readonly errors: ReadonlyArray<FieldError>;

  constructor(problem: Problem);
  constructor(code: string, status: number, message: string, details?: unknown);
  constructor(
    problemOrCode: Problem | string,
    status?: number,
    message?: string,
    detailsArg?: unknown,
  ) {
    const problem = typeof problemOrCode === 'string'
      ? {
          code: problemOrCode,
          status: status ?? 0,
          title: message ?? 'Erro inesperado.',
          detail: message ?? 'Erro inesperado.',
        }
      : problemOrCode;

    super(problem.detail || problem.title);
    this.name = 'ApiError';
    this.status = problem.status;
    this.code = problem.code;
    this.title = problem.title;
    this.detail = problem.detail;
    this.details = detailsArg;
    this.instance = problem.instance;
    this.errors = Object.freeze([...(problem.errors ?? [])]);
  }

  static fromLegacy(code: string, status: number, message: string, _details?: unknown): ApiError {
    return new ApiError({
      code,
      status,
      title: message,
      detail: message,
      errors: [],
    });
  }

  getFieldErrors(): Record<string, FieldError> {
    const mapped: Record<string, FieldError> = {};
    for (const fieldError of this.errors) {
      mapped[fieldError.field] = fieldError;
    }
    return mapped;
  }

  hasFieldError(field: string): boolean {
    return this.errors.some((item) => item.field === field);
  }
}

function unmappedFallback(code: string): string {
  return `Não foi possível concluir a ação. Código: ${code}`;
}

function reportUnmapped(code: string, context: string): void {
  // Breadcrumb so unmapped codes surface in observability. The PT-BR
  // fallback above keeps the UI usable, but the code must reach the team.
  console.warn('[api] unmapped error code', { code, context });
}

export function resolveErrorMessage(err: unknown): string;
export function resolveErrorMessage(code: string | undefined, backendMessage?: string): string;
export function resolveErrorMessage(errOrCode: unknown, backendMessage?: string): string {
  if (errOrCode instanceof ApiError) {
    const mapped = mergedMessages[errOrCode.code];
    if (mapped) return mapped;
    reportUnmapped(errOrCode.code, errOrCode.instance ?? 'ApiError');
    return unmappedFallback(errOrCode.code);
  }

  if (typeof errOrCode === 'string' || errOrCode === undefined) {
    if (typeof errOrCode === 'string') {
      const mapped = mergedMessages[errOrCode];
      if (mapped) return mapped;
      reportUnmapped(errOrCode, 'resolveErrorMessage(code)');
      return backendMessage ?? unmappedFallback(errOrCode);
    }

    if (backendMessage) return backendMessage;
    return 'Erro inesperado.';
  }

  if (errOrCode instanceof Error) {
    return errOrCode.message;
  }

  return 'Erro inesperado.';
}
