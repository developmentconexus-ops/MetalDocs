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
}

// Mock extras — positional, wraps around if more items than mock entries
// TODO [BACKLOG: caixa-aprovacao.md]: Remove when API provides these fields
const MOCK_EXTRAS: Omit<RichInboxItem, keyof InboxItem>[] = [
  { code: 'POP-QUA-0148', kind: 'POP', deadline: '3h 28min', urgent: true, summary: 'Revisão da inspeção de soldas em juntas críticas. Adiciona §3.2 sobre inspeção visual ampliada conforme NBR ISO 5817 (qualidade B).', changes: 12, version: 'v2.3 → v2.4' },
  { code: 'IT-PROD-0072', kind: 'IT', deadline: '9h', urgent: true, summary: 'Procedimento de setup revisado após incidente de 12/abr. Inclui dupla checagem de pressão (ponto crítico §4.1).', changes: 8, version: 'v1.6 → v1.7' },
  { code: 'POL-RH-0011', kind: 'POL', deadline: '1 dia', urgent: false, summary: 'Atualização para incluir modalidade híbrida pós-licença médica.', changes: 4, version: 'v3.0 → v3.1' },
  { code: 'DC-TI-0203', kind: 'DC', deadline: '2 dias', urgent: false, summary: 'Reflete a segregação para ICS/SCADA aprovada no Q1.', changes: 6, version: 'v0.9 → v1.0' },
  { code: 'POP-QUA-0149', kind: 'POP', deadline: '4 dias', urgent: false, summary: 'Revisão de procedimento padrão.', changes: 3, version: 'v1.1 → v1.2' },
];

export function enrichInboxItem(item: InboxItem, idx: number): RichInboxItem {
  // TODO [BACKLOG: caixa-aprovacao.md]: Replace mock spread with real API fields
  const mock = MOCK_EXTRAS[idx % MOCK_EXTRAS.length];
  return { ...item, ...mock };
}
