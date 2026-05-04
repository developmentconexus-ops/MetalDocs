/**
 * resolveErrorMessage — maps an API error code to a localised user-facing string.
 * Falls back to `fallback` when no mapping exists.
 * Created as a sub-plan 1 primitive; consumed by rename-rollback (E9) and other consumer-facing handlers.
 */
const ERROR_MESSAGES: Record<string, string> = {
  not_found: 'Recurso não encontrado.',
  conflict: 'Conflito ao processar a solicitação.',
  forbidden: 'Você não tem permissão para realizar esta ação.',
  validation_error: 'Dados inválidos. Verifique os campos e tente novamente.',
  server_error: 'Erro interno do servidor. Tente novamente mais tarde.',
};

export function resolveErrorMessage(
  code: string | undefined,
  fallback: string,
): string {
  if (code && ERROR_MESSAGES[code]) return ERROR_MESSAGES[code];
  return fallback;
}
