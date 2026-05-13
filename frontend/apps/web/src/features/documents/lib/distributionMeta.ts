import type { TabBarItem } from '../../../components/ui/TabBar';

export type RecipientStatus = 'ack' | 'read' | 'pending' | 'overdue';

export const RECIPIENT_STATUS_TONE: Record<RecipientStatus, { label: string; bg: string; fg: string }> = {
  ack:     { label: 'Reconheceu',  bg: 'var(--success-bg)', fg: 'var(--success)' },
  read:    { label: 'Apenas leu',  bg: 'var(--brand-pale)', fg: 'var(--brand)' },
  pending: { label: 'Pendente',    bg: 'var(--surface-3)',  fg: 'var(--text-muted)' },
  overdue: { label: 'Em atraso',   bg: 'var(--danger-bg)',  fg: 'var(--danger)' },
};

export const RECIPIENT_TABS: TabBarItem[] = [
  { key: 'pending',  label: 'Pendentes',    count: 4 },
  { key: 'overdue',  label: 'Em atraso',    count: 3 },
  { key: 'read',     label: 'Apenas leram', count: 2 },
  { key: 'ack',      label: 'Reconheceram', count: 4 },
  { key: 'all',      label: 'Todos',        count: 11 },
];

export interface RecipientMock {
  who: string;
  role: string;
  area: string;
  status: RecipientStatus;
  overdue: boolean;
  last: string;
  when: string;
}

export interface AreaCoverageMock {
  area: string;
  total: number;
  read: number;
  ack: number;
}

export interface DailyReadMock {
  day: string;
  cumAck: number;
}

export interface DistributionMock {
  code: string;
  title: string;
  version: string;
  area: string;
  areaShort: string;
  typeShort: string;
  publishedAt: string;
  publishedAtFull: string;
  deadline: string;
  deadlineFull: string;
  daysLeft: number;
  policy: string;
  channel: string;
  remindersSent: number;
  remindersScheduled: string;
  totalTargets: number;
  read: number;
  acknowledged: number;
  pending: number;
  overdue: number;
  dailyReads: DailyReadMock[];
  byArea: AreaCoverageMock[];
  people: RecipientMock[];
}

// TODO(distribuicao:distribution): replace with useDistributionQuery(documentId)
// when GET /api/v1/documents/:id/distribution ships.
// Backlog: wiki/backlog/distribuicao.md
export const MOCK_DISTRIBUTION: DistributionMock = {
  code: 'PR-EHS-014',
  title: 'Procedimento de Bloqueio e Etiquetagem (LOTO)',
  version: 'v3.2',
  area: 'Saúde, Segurança & Meio Ambiente',
  areaShort: 'SSMA',
  typeShort: 'Procedimento',
  publishedAt: '12 mar 2026',
  publishedAtFull: '12 mar 2026 · 14:32',
  deadline: '22 mar 2026',
  deadlineFull: '22 mar 2026 · 23:59',
  daysLeft: 8,
  policy: 'Reconhecimento obrigatório (assinatura)',
  channel: 'E-mail + portal interno',
  remindersSent: 2,
  remindersScheduled: '20 mar · 08:00',
  totalTargets: 248,
  read: 182,
  acknowledged: 156,
  pending: 66,
  overdue: 12,
  dailyReads: [
    { day: '12/03', cumAck: 22 },
    { day: '13/03', cumAck: 58 },
    { day: '14/03', cumAck: 84 },
    { day: '15/03', cumAck: 102 },
    { day: '16/03', cumAck: 118 },
    { day: '17/03', cumAck: 132 },
    { day: '18/03', cumAck: 144 },
    { day: '19/03', cumAck: 156 },
  ],
  byArea: [
    { area: 'Manutenção',         total: 64, read: 58, ack: 52 },
    { area: 'Produção · Linha A', total: 42, read: 38, ack: 34 },
    { area: 'SSMA',               total: 18, read: 18, ack: 18 },
    { area: 'Engenharia',         total: 24, read: 19, ack: 16 },
    { area: 'Produção · Linha B', total: 38, read: 22, ack: 18 },
    { area: 'Logística',          total: 28, read: 14, ack: 10 },
    { area: 'Qualidade',          total: 16, read: 8,  ack: 5  },
    { area: 'Administrativo',     total: 18, read: 5,  ack: 3  },
  ],
  // TODO(distribuicao:recipients): replace with useRecipientsQuery(documentId, { tab, search, page })
  // when GET /api/v1/documents/:id/distribution/recipients ships.
  // Backlog: wiki/backlog/distribuicao.md
  people: [
    { who: 'Roberto Carvalho',   role: 'Op. Empilhadeira',    area: 'Logística',          status: 'pending', overdue: false, last: 'nunca abriu',          when: '—' },
    { who: 'Ana Paula Mendes',   role: 'Analista RH',         area: 'Administrativo',     status: 'pending', overdue: true,  last: 'nunca abriu',          when: '—' },
    { who: 'Thiago Rocha',       role: 'Insp. Qualidade',    area: 'Qualidade',          status: 'read',    overdue: false, last: 'apenas leu',           when: '08 mar 2026' },
    { who: 'Felipe Andrade',     role: 'Op. Multifuncional',  area: 'Produção · B',       status: 'pending', overdue: false, last: 'aberto há 6 dias',     when: '14 mar 2026' },
    { who: 'Beatriz Moreira',    role: 'Aux. Administrativo', area: 'Administrativo',     status: 'pending', overdue: true,  last: 'nunca abriu',          when: '—' },
    { who: 'Gabriel Costa',      role: 'Conferente',          area: 'Logística',          status: 'pending', overdue: true,  last: 'nunca abriu',          when: '—' },
    { who: 'Daniela Prado',      role: 'Téc. Mecânica',       area: 'Manutenção',         status: 'ack',     overdue: false, last: 'reconheceu há 4 min',  when: '19 mar · 14:08' },
    { who: 'Eduardo Castro',     role: 'Eng. Processo',       area: 'Engenharia',         status: 'ack',     overdue: false, last: 'reconheceu há 28 min', when: '19 mar · 13:44' },
    { who: 'Marina Albuquerque', role: 'Téc. Elétrica',       area: 'Manutenção',         status: 'read',    overdue: false, last: 'apenas leu há 1h',     when: '19 mar · 13:08' },
    { who: 'João Vitor Lima',    role: 'Líder de Turno',      area: 'Produção · Linha A', status: 'ack',     overdue: false, last: 'reconheceu ontem',     when: '18 mar · 17:22' },
    { who: 'Helena Tavares',     role: 'Téc. Segurança',      area: 'SSMA',               status: 'ack',     overdue: false, last: 'reconheceu há 2 dias', when: '17 mar · 09:14' },
  ],
};
