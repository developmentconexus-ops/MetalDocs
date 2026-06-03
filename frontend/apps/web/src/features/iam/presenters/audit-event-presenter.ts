// Canonical map keyed off backend audit action enums. Update whenever a new
// `recordAudit(...)` call site or `audit.Record(..., Action: ...)` is added in
// internal/modules/{iam,audit,auth,...}. Never sniff substrings — match exact
// keys so an unknown action surfaces as the explicit `unknown` fallback.

export type AuditEventSeverity = "info" | "warn" | "error" | "success";

export type AuditEventCategory =
  | "auth"
  | "iam"
  | "document"
  | "template"
  | "workflow"
  | "audit"
  | "other";

interface AuditEventDescriptor {
  label: string;
  severity: AuditEventSeverity;
  category: AuditEventCategory;
}

const REGISTRY: Readonly<Record<string, AuditEventDescriptor>> = {
  // ── auth ───────────────────────────────────────────────────────────────
  "auth.login": { label: "Login realizado", severity: "success", category: "auth" },
  "auth.login.success": { label: "Login realizado", severity: "success", category: "auth" },
  "auth.login.failed": { label: "Falha no login", severity: "warn", category: "auth" },
  "auth.logout": { label: "Logout", severity: "info", category: "auth" },
  "auth.password.changed": { label: "Senha alterada", severity: "info", category: "auth" },
  "auth.user.password_reset": { label: "Senha redefinida pelo admin", severity: "warn", category: "auth" },
  "auth.user.unlocked": { label: "Usuário desbloqueado", severity: "info", category: "auth" },
  "auth.session.revoked": { label: "Sessão revogada", severity: "warn", category: "auth" },

  // ── iam ────────────────────────────────────────────────────────────────
  "iam.user.invited": { label: "Usuário convidado", severity: "info", category: "iam" },
  "iam.user.updated": { label: "Usuário atualizado", severity: "info", category: "iam" },
  "iam.user.role.upserted": { label: "Função aplicada", severity: "info", category: "iam" },
  "iam.user.roles.replaced": { label: "Funções substituídas", severity: "info", category: "iam" },
  "iam.user.bulk.deactivate": { label: "Desativação em lote", severity: "warn", category: "iam" },
  "iam.user.bulk.force-logout": { label: "Logout forçado em lote", severity: "warn", category: "iam" },
  "iam.user.bulk.unlock": { label: "Desbloqueio em lote", severity: "info", category: "iam" },

  // ── documents ──────────────────────────────────────────────────────────
  "session.acquired": { label: "Edição iniciada", severity: "info", category: "document" },
  "session.released": { label: "Edição liberada", severity: "info", category: "document" },
  "doc.create": { label: "Documento criado", severity: "info", category: "document" },
  "doc.submit": { label: "Documento submetido", severity: "info", category: "document" },
  "doc.publish": { label: "Documento publicado", severity: "success", category: "document" },
  "doc.signoff": { label: "Documento assinado", severity: "success", category: "document" },
  "doc.supersede": { label: "Documento substituído", severity: "warn", category: "document" },
  "document.create": { label: "Documento criado", severity: "info", category: "document" },
  "document.edit": { label: "Documento editado", severity: "info", category: "document" },
  "document.rename": { label: "Documento renomeado", severity: "info", category: "document" },
  "document.submit": { label: "Documento submetido", severity: "info", category: "document" },
  "document.signoff": { label: "Documento assinado", severity: "success", category: "document" },
  "document.view": { label: "Documento visualizado", severity: "info", category: "document" },
  "controlled_documents.create": { label: "Documento controlado criado", severity: "info", category: "document" },
  "controlled_documents.supersede": { label: "Documento controlado substituído", severity: "warn", category: "document" },
  "controlled_documents.obsolete": { label: "Documento controlado obsoletado", severity: "warn", category: "document" },

  // ── templates ──────────────────────────────────────────────────────────
  "template.create": { label: "Template criado", severity: "info", category: "template" },
  "template.edit": { label: "Template editado", severity: "info", category: "template" },
  "template.submit": { label: "Template submetido", severity: "info", category: "template" },
  "template.review": { label: "Template revisado", severity: "info", category: "template" },
  "template.approve": { label: "Template aprovado", severity: "success", category: "template" },
  "template.publish": { label: "Template publicado", severity: "success", category: "template" },
  "template.view": { label: "Template visualizado", severity: "info", category: "template" },

  // ── workflow ───────────────────────────────────────────────────────────
  "workflow.review": { label: "Workflow revisado", severity: "info", category: "workflow" },
  "workflow.approve": { label: "Workflow aprovado", severity: "success", category: "workflow" },

  // ── audit subsystem ────────────────────────────────────────────────────
  "audit.read": { label: "Auditoria consultada", severity: "info", category: "audit" },
  "audit.export.requested": { label: "Exportação de auditoria solicitada", severity: "info", category: "audit" },
  "audit.test": { label: "Evento de teste", severity: "info", category: "audit" },
};

const UNKNOWN: AuditEventDescriptor = {
  label: "Ação registrada",
  severity: "info",
  category: "other",
};

function describe(action: string): AuditEventDescriptor {
  return REGISTRY[action] ?? UNKNOWN;
}

export function getAuditEventLabel(action: string): string {
  return describe(action).label;
}

export function getAuditEventSeverity(action: string): AuditEventSeverity {
  return describe(action).severity;
}

export function getAuditEventCategory(action: string): AuditEventCategory {
  return describe(action).category;
}

// Action enum keys for facet filters. Exposed as readonly so consumers can
// build <option> lists keyed off the canonical map.
export const KNOWN_AUDIT_ACTIONS: ReadonlyArray<string> = Object.keys(REGISTRY);

export const AUDIT_CATEGORY_LABELS: Readonly<Record<AuditEventCategory, string>> = {
  auth: "Autenticação",
  iam: "Usuários e funções",
  document: "Documentos",
  template: "Templates",
  workflow: "Workflow",
  audit: "Auditoria",
  other: "Outros",
};
