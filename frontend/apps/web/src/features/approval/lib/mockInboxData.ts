// TODO [BACKLOG: caixa-aprovacao.md]: ALL fields in RichInboxItem beyond InboxItem base are MOCK.
// Replace with real API fields when backend adds them.
// Tracked in wiki/backlog/caixa-aprovacao.md.

import type { InboxItem } from '../api/approvalTypes';

export interface RichInboxItem extends InboxItem {
  /** TODO [BACKLOG]: needs InboxItem.code from API */
  code: string;
  /** TODO [BACKLOG]: needs InboxItem.kind from API */
  kind: string;
  /** TODO [BACKLOG]: needs InboxItem.deadline from API (InboxItem.deadline_at) */
  deadline: string;
  /** TODO [BACKLOG]: needs InboxItem.urgent from API */
  urgent: boolean;
  /** TODO [BACKLOG]: needs InboxItem.summary from API */
  summary: string;
  /** TODO [BACKLOG]: needs InboxItem.changes from API */
  changes: number;
  /** TODO [BACKLOG]: needs InboxItem.version from API */
  version: string;
  /** TODO [BACKLOG]: needs real deadline_at ISO timestamp from API */
  deadline_at: string;
}

// Mock extras — positional, wraps around if more items than mock entries
// TODO [BACKLOG: caixa-aprovacao.md]: Remove when API provides these fields
const MOCK_EXTRAS: Omit<RichInboxItem, keyof InboxItem>[] = [
  { code: 'POP-QUA-0148', kind: 'POP', deadline: '3h 28min', urgent: true, summary: 'Revisão da inspeção de soldas em juntas críticas. Adiciona §3.2 sobre inspeção visual ampliada conforme NBR ISO 5817 (qualidade B).', changes: 12, version: 'v2.3 → v2.4', deadline_at: new Date(Date.now() + 3.5 * 3600 * 1000).toISOString() },
  { code: 'IT-PROD-0072', kind: 'IT', deadline: '9h', urgent: true, summary: 'Procedimento de setup revisado após incidente de 12/abr. Inclui dupla checagem de pressão (ponto crítico §4.1).', changes: 8, version: 'v1.6 → v1.7', deadline_at: new Date(Date.now() + 9 * 3600 * 1000).toISOString() },
  { code: 'POL-RH-0011', kind: 'POL', deadline: '1 dia', urgent: false, summary: 'Atualização para incluir modalidade híbrida pós-licença médica.', changes: 4, version: 'v3.0 → v3.1', deadline_at: new Date(Date.now() + 28 * 3600 * 1000).toISOString() },
  { code: 'DC-TI-0203', kind: 'DC', deadline: '2 dias', urgent: false, summary: 'Reflete a segregação para ICS/SCADA aprovada no Q1.', changes: 6, version: 'v0.9 → v1.0', deadline_at: new Date(Date.now() + 52 * 3600 * 1000).toISOString() },
  { code: 'POP-QUA-0149', kind: 'POP', deadline: '4 dias', urgent: false, summary: 'Revisão de procedimento padrão.', changes: 3, version: 'v1.1 → v1.2', deadline_at: new Date(Date.now() + 100 * 3600 * 1000).toISOString() },
];

// Canonical mock InboxItem base objects matching the 4 hardcoded Phase 3a items.
// TODO [BACKLOG: caixa-aprovacao.md]: Replace with real API data when endpoint is stable.
export const MOCK_INBOX_ITEMS: RichInboxItem[] = [
  {
    instance_id: 'mock-inst-001',
    document_id: 'mock-doc-001',
    controlled_document_id: 'POP-QUA-0148',
    document_title: 'Inspeção de recebimento — solda MIG/MAG',
    area_code: 'QUA',
    submitted_by: 'Carla Pinheiro',
    submitted_at: new Date(Date.now() - 2 * 24 * 3600 * 1000).toISOString(),
    stage_label: 'L2 · Você',
    quorum_progress: '0/1',
    ...MOCK_EXTRAS[0],
  },
  {
    instance_id: 'mock-inst-002',
    document_id: 'mock-doc-002',
    controlled_document_id: 'IT-PROD-0072',
    document_title: 'Setup da prensa hidráulica P-440',
    area_code: 'PROD',
    submitted_by: 'João Aguiar',
    submitted_at: new Date(Date.now() - 1 * 24 * 3600 * 1000).toISOString(),
    stage_label: 'L2 · Você',
    quorum_progress: '0/1',
    ...MOCK_EXTRAS[1],
  },
  {
    instance_id: 'mock-inst-003',
    document_id: 'mock-doc-003',
    controlled_document_id: 'POL-RH-0011',
    document_title: 'Política de afastamento parcial',
    area_code: 'RH',
    submitted_by: 'Fábio Ribas',
    submitted_at: new Date(Date.now() - 3 * 24 * 3600 * 1000).toISOString(),
    stage_label: 'L2 · Você',
    quorum_progress: '0/1',
    ...MOCK_EXTRAS[2],
  },
  {
    instance_id: 'mock-inst-004',
    document_id: 'mock-doc-004',
    controlled_document_id: 'DC-TI-0203',
    document_title: 'Diagrama de rede — DMZ planta 2',
    area_code: 'TI',
    submitted_by: 'Ana Iuri',
    submitted_at: new Date(Date.now() - 4 * 24 * 3600 * 1000).toISOString(),
    stage_label: 'L2 · Você',
    quorum_progress: '0/1',
    ...MOCK_EXTRAS[3],
  },
];

export function enrichInboxItem(item: InboxItem, idx: number): RichInboxItem {
  // TODO [BACKLOG: caixa-aprovacao.md]: Replace mock spread with real API fields
  const mock = MOCK_EXTRAS[idx % MOCK_EXTRAS.length];
  return { ...item, ...mock };
}
