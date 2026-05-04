import { errorMessages } from "./errorMessages";

export class ApiError extends Error {
  constructor(
    public readonly code: string,
    public readonly status: number,
    message: string,
    public readonly details?: unknown,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

export function resolveErrorMessage(code: string | undefined, backendMessage?: string): string {
  if (code && errorMessages[code]) {
    return errorMessages[code];
  }

  if (backendMessage) {
    return backendMessage;
  }

  return "Ocorreu um erro. Tente novamente.";
}
