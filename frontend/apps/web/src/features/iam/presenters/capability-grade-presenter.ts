// Capability classification for the Roles & Capabilities matrix.
// - Domain: pt-BR grouping shown as matrix columns.
// - Grade: view / manage / approval — drives cell dot rendering.
//
// Source of truth: internal/modules/iam/domain/model.go (Capability constants)
// and internal/modules/iam/domain/catalog.go (descriptions / categories).

export type CapabilityGrade = "view" | "manage" | "approval";

export type CapabilityDomain =
  | "document"
  | "approval"
  | "taxonomy"
  | "user"
  | "membership"
  | "audit"
  | "metrics";

interface DomainDescriptor {
  key: CapabilityDomain;
  label: string;
  helper: string;
  prefixes: ReadonlyArray<string>;
}

export const CAPABILITY_DOMAINS: ReadonlyArray<DomainDescriptor> = [
  {
    key: "document",
    label: "Documentos",
    helper: "Documentos, templates e ciclo controlado",
    prefixes: ["document", "doc", "controlled_documents", "template"],
  },
  {
    key: "approval",
    label: "Aprovação",
    helper: "Workflow e rotas de aprovação",
    prefixes: ["workflow", "route"],
  },
  {
    key: "taxonomy",
    label: "Taxonomia",
    helper: "Famílias, perfis e áreas",
    prefixes: ["taxonomy"],
  },
  {
    key: "user",
    label: "Usuários",
    helper: "Conta de usuário e sessão",
    prefixes: ["user", "session"],
  },
  {
    key: "membership",
    label: "Memberships",
    helper: "Vínculos de usuário a áreas",
    prefixes: ["membership"],
  },
  { key: "audit", label: "Auditoria", helper: "Trilha de auditoria", prefixes: ["audit"] },
  { key: "metrics", label: "Métricas", helper: "KPIs e telemetria", prefixes: ["metrics"] },
];

function prefixOf(code: string): string {
  const i = code.indexOf(".");
  return i < 0 ? code : code.slice(0, i);
}

function suffixOf(code: string): string {
  const i = code.lastIndexOf(".");
  return i < 0 ? code : code.slice(i + 1);
}

export function getCapabilityDomain(code: string): CapabilityDomain | null {
  const prefix = prefixOf(code);
  for (const d of CAPABILITY_DOMAINS) {
    if (d.prefixes.includes(prefix)) return d.key;
  }
  return null;
}

const APPROVAL_SUFFIXES = new Set(["submit", "review", "approve", "signoff"]);
const VIEW_SUFFIXES = new Set(["view", "read"]);

export function getCapabilityGrade(code: string): CapabilityGrade {
  const suffix = suffixOf(code);
  if (VIEW_SUFFIXES.has(suffix)) return "view";
  if (APPROVAL_SUFFIXES.has(suffix)) return "approval";
  return "manage";
}

export const GRADE_LABEL: Readonly<Record<CapabilityGrade, string>> = {
  view: "Leitura",
  manage: "Gestão",
  approval: "Aprovação",
};

export const GRADE_ORDER: ReadonlyArray<CapabilityGrade> = ["view", "manage", "approval"];
