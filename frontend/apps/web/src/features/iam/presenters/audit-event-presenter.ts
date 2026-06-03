type AuditEventSeverity = "info" | "warn" | "error" | "success";

const LABEL_MAP: Readonly<Record<string, string>> = {
  "auth.login.success": "Login realizado",
  "auth.login.failed": "Falha no login",
};

const SEVERITY_MAP: Readonly<Record<string, AuditEventSeverity>> = {
  "auth.login.success": "success",
  "auth.login.failed": "warn",
};

export function getAuditEventLabel(action: string): string {
  return LABEL_MAP[action] ?? "Ação registrada";
}

export function getAuditEventSeverity(action: string): AuditEventSeverity {
  return SEVERITY_MAP[action] ?? "info";
}
