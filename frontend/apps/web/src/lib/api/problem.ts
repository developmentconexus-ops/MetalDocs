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

const extraMessages: Record<string, string> = {
  VALIDATION_ERROR: 'Há campos inválidos. Revise os dados informados.',
  UNKNOWN_FIELD: 'A requisição contém um campo desconhecido.',
  UNKNOWN_FILTER: 'Foi informado um filtro não suportado para esta consulta.',
  INVALID_SORT_FIELD: 'O campo de ordenação informado não é válido.',
  INVALID_CURSOR: 'O cursor informado é inválido ou expirou.',
  INCLUDE_NOT_SUPPORTED: 'O include informado não é suportado para este endpoint.',
  UNAUTHENTICATED: 'Você precisa se autenticar para continuar.',
  FORBIDDEN_CAPABILITY: 'Você não tem permissão para executar esta ação.',
  FORBIDDEN_AREA: 'Você não tem permissão para atuar nesta área.',
  NOT_FOUND: 'O recurso solicitado não foi encontrado.',
  ALREADY_EXISTS: 'O recurso informado já existe.',
  STATE_TRANSITION_INVALID: 'A transição de estado solicitada não é permitida.',
  CONCURRENT_MODIFICATION: 'O recurso foi alterado por outro usuário. Atualize e tente novamente.',
  IDEMPOTENCY_KEY_REUSED: 'A chave de idempotência já foi usada com outro conteúdo.',
  IDEMPOTENCY_REPLAY: 'A requisição foi processada anteriormente e a resposta foi reaproveitada.',
  RATE_LIMITED: 'Muitas requisições em sequência. Tente novamente em instantes.',
  INTERNAL_ERROR: 'Ocorreu um erro interno. Tente novamente.',
};

const mergedMessages: Record<string, string> = {
  ...errorMessages,
  ...extraMessages,
};

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

export function resolveErrorMessage(err: unknown): string;
export function resolveErrorMessage(code: string | undefined, backendMessage?: string): string;
export function resolveErrorMessage(errOrCode: unknown, backendMessage?: string): string {
  if (errOrCode instanceof ApiError) {
    const mapped = mergedMessages[errOrCode.code];
    if (mapped) return mapped;
    return errOrCode.detail || errOrCode.title || 'Erro inesperado.';
  }

  if (typeof errOrCode === 'string' || errOrCode === undefined) {
    if (typeof errOrCode === 'string') {
      const mapped = mergedMessages[errOrCode];
      if (mapped) return mapped;
    }

    if (backendMessage) return backendMessage;
    return 'Erro inesperado.';
  }

  if (errOrCode instanceof Error) {
    return errOrCode.message;
  }

  return 'Erro inesperado.';
}
