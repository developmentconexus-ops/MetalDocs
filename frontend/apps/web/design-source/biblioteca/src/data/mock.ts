import type { Area, Profile, Document, Stat, StatusConfig, DocStatus } from "../types";

export const AREAS: Area[] = [
  { code: "ven", label: "Vendas / Atendimento", color: "#3A45A0", count: 18 },
  { code: "est", label: "Estoque / Logística",  color: "#2A6B35", count: 15 },
  { code: "fin", label: "Financeiro / Adm.",    color: "#7A5010", count: 15 },
  { code: "rh",  label: "RH / Pessoas",         color: "#6B2A7A", count: 15 },
  { code: "com", label: "Compras",              color: "#7A2020", count: 15 },
];

export const PROFILES: Profile[] = [
  { code: "po", label: "Procedimentos",          family: "procedure",        bg: "#EEEFFE", color: "#3238A8", count: 30 },
  { code: "it", label: "Instruções de trabalho", family: "work_instruction", bg: "#EDFAF2", color: "#1A6B3A", count: 18 },
  { code: "rg", label: "Registros",              family: "record",           bg: "#FEF7E8", color: "#8A5800", count: 15 },
  { code: "fm", label: "Formulários",            family: "form",             bg: "#F8EEFB", color: "#6B2A7A", count: 15 },
];

export const STATUS_CONFIG: Record<DocStatus, StatusConfig> = {
  vigente: { label: "Vigente",    bg: "#E6F5EC", color: "#1A6B35" },
  draft:   { label: "Rascunho",   bg: "#FDF5E0", color: "#7A5010" },
  review:  { label: "Em revisão", bg: "#E8EEF8", color: "#1A3A7A" },
};

export const DOCUMENTS: Document[] = [
  { id: 1,  code: "PO-VEN-001", title: "Atendimento e venda — balcão / WhatsApp",  profile: "po", area: "ven", status: "vigente", owner: "Leandro M.", version: "v1.2", nextReview: "Mar 2027", urgent: false },
  { id: 2,  code: "PO-VEN-002", title: "Orçamento e proposta comercial",            profile: "po", area: "ven", status: "vigente", owner: "Leandro M.", version: "v1.0", nextReview: "Mar 2027", urgent: false },
  { id: 3,  code: "PO-VEN-003", title: "Pedido, faturamento e emissão de NF",       profile: "po", area: "ven", status: "vigente", owner: "Ana P.",     version: "v1.1", nextReview: "Jun 2027", urgent: false },
  { id: 4,  code: "PO-VEN-004", title: "Venda externa — novo canal",                profile: "po", area: "ven", status: "review",  owner: "Leandro M.", version: "v0.2", nextReview: "Abr 2026", urgent: true  },
  { id: 5,  code: "PO-VEN-005", title: "Pós-venda e tratamento de reclamação",      profile: "po", area: "ven", status: "vigente", owner: "Ana P.",     version: "v1.0", nextReview: "Mar 2027", urgent: false },
  { id: 6,  code: "IT-VEN-001", title: "Script de atendimento ao cliente",           profile: "it", area: "ven", status: "vigente", owner: "Ana P.",     version: "v1.1", nextReview: "Jan 2027", urgent: false },
  { id: 7,  code: "IT-VEN-006", title: "Roteiro do vendedor externo",                profile: "it", area: "ven", status: "draft",   owner: "Leandro M.", version: "v0.1", nextReview: "—",        urgent: false },
  { id: 8,  code: "RG-VEN-002", title: "Reclamações de clientes",                   profile: "rg", area: "ven", status: "vigente", owner: "Ana P.",     version: "v1.0", nextReview: "Jun 2026", urgent: false },
  { id: 9,  code: "FM-VEN-001", title: "Formulário de orçamento",                   profile: "fm", area: "ven", status: "vigente", owner: "Leandro M.", version: "v1.0", nextReview: "Mar 2027", urgent: false },
  { id: 10, code: "FM-VEN-002", title: "Cadastro de cliente PJ",                    profile: "fm", area: "ven", status: "vigente", owner: "Leandro M.", version: "v1.0", nextReview: "Mar 2027", urgent: false },
  { id: 11, code: "PO-EST-001", title: "Recebimento de mercadoria",                 profile: "po", area: "est", status: "vigente", owner: "Carlos R.",  version: "v2.0", nextReview: "Ago 2026", urgent: false },
  { id: 12, code: "PO-EST-002", title: "Movimentação e armazenagem",                profile: "po", area: "est", status: "vigente", owner: "Carlos R.",  version: "v1.0", nextReview: "Ago 2026", urgent: false },
  { id: 13, code: "IT-EST-001", title: "Conferência de NF no recebimento",          profile: "it", area: "est", status: "vigente", owner: "Carlos R.",  version: "v1.3", nextReview: "Mai 2026", urgent: true  },
  { id: 14, code: "RG-EST-001", title: "Recebimento de mercadoria",                 profile: "rg", area: "est", status: "vigente", owner: "Carlos R.",  version: "v1.0", nextReview: "Ago 2026", urgent: false },
  { id: 15, code: "FM-EST-001", title: "Formulário de recebimento",                 profile: "fm", area: "est", status: "draft",   owner: "Carlos R.",  version: "v0.2", nextReview: "—",        urgent: false },
  { id: 16, code: "PO-FIN-001", title: "Contas a pagar",                            profile: "po", area: "fin", status: "vigente", owner: "Marta S.",   version: "v1.0", nextReview: "Nov 2026", urgent: false },
  { id: 17, code: "PO-FIN-002", title: "Contas a receber e cobrança",               profile: "po", area: "fin", status: "vigente", owner: "Marta S.",   version: "v1.0", nextReview: "Nov 2026", urgent: false },
  { id: 18, code: "IT-FIN-002", title: "Régua de cobrança",                         profile: "it", area: "fin", status: "review",  owner: "Marta S.",   version: "v0.3", nextReview: "Mai 2026", urgent: true  },
  { id: 19, code: "PO-RH-001",  title: "Recrutamento e seleção",                    profile: "po", area: "rh",  status: "vigente", owner: "Julia T.",   version: "v1.1", nextReview: "Dez 2026", urgent: false },
  { id: 20, code: "PO-RH-002",  title: "Integração (onboarding)",                   profile: "po", area: "rh",  status: "draft",   owner: "Julia T.",   version: "v0.1", nextReview: "—",        urgent: false },
  { id: 21, code: "PO-COM-001", title: "Solicitação e aprovação de compra",         profile: "po", area: "com", status: "vigente", owner: "Leandro M.", version: "v1.0", nextReview: "Out 2026", urgent: false },
  { id: 22, code: "FM-COM-002", title: "Mapa de cotação comparativa",               profile: "fm", area: "com", status: "vigente", owner: "Leandro M.", version: "v1.0", nextReview: "Out 2026", urgent: false },
];

export const STATS: Stat[] = [
  { label: "Total no acervo",    value: 78, sub: "5 áreas de processo",  color: "#6B1F2A", pct: 100 },
  { label: "Vigentes",           value: 61, sub: "78% do acervo",        color: "#1A6B35", pct: 78  },
  { label: "Em andamento",       value: 12, sub: "Rascunho ou revisão",  color: "#C87020", pct: 15  },
  { label: "Revisões pendentes", value: 5,  sub: "Próximos 30 dias",     color: "#C8364A", pct: 6   },
];
