// Centralized TanStack Query key constants.
// All useQuery / invalidateQueries calls must import from here - never inline string arrays.

type InboxParams = {
  page?: number;
  areaFilter?: string;
  onlyOverdue?: boolean;
  limit?: number;
};

type ControlledDocFilter = {
  profileCode?: string;
  processAreaCode?: string;
  status?: string;
};

type DocumentLibraryListParams = {
  page?: number;
  pageSize?: number;
  status?: string;
  areaCode?: string;
  profileCode?: string;
  q?: string;
  includeArchived?: boolean;
};

export const QK = {
  documents: {
    all: ['documents'] as const,
    list: (params: DocumentLibraryListParams = {}) => ['documents', 'list', params] as const,
    stats: () => ['documents', 'stats'] as const,
    detail: (id: string) => ['documents', 'detail', id] as const,
    revisionHistory: (id: string) => ['documents', 'revision-history', id] as const,
    comments: (id: string) => ['documents', 'comments', id] as const,
  },
  inbox: (params: InboxParams = {}) =>
    ['approval', 'inbox', params] as const,
  audit: {
    // GET /audit/events
    recent: (limit = 10) => ['audit', 'recent', limit] as const,
  },
  controlledDocuments: {
    list: (filter: ControlledDocFilter = {}) =>
      ['controlled-documents', 'list', filter] as const,
    detail: (id: string) => ['controlled-documents', 'detail', id] as const,
    activeDocument: (id: string) => ['controlled-documents', 'active-document', id] as const,
    preview: (profileCode: string, areaCode: string) =>
      ['controlled-documents', 'preview', profileCode, areaCode] as const,
  },
  taxonomy: {
    profiles: () => ['taxonomy', 'profiles'] as const,
    areas: () => ['taxonomy', 'areas'] as const,
  },
  templates: {
    list: () => ['templates', 'list'] as const,
    blank: () => ['templates', 'blank'] as const,
    byProfile: (profileCode: string) =>
      ['templates', 'by-profile', profileCode] as const,
  },
  approval: {
    instance: (documentId: string) =>
      ['approval', 'instance', documentId] as const,
    routes: {
      list: () => ['approval', 'routes', 'list'] as const,
      detail: (id: string) => ['approval', 'routes', 'detail', id] as const,
    },
  },
  iam: {
    roles: () => ['iam', 'roles'] as const,
    capabilities: () => ['iam', 'capabilities'] as const,
    roleCapabilities: () => ['iam', 'role-capabilities'] as const,
    adminOverview: () => ['iam', 'admin', 'overview'] as const,
    kpi: () => ['iam', 'admin', 'kpi'] as const,
    usage: () => ['iam', 'admin', 'usage'] as const,
    presenceSnapshot: () => ['iam', 'admin', 'presence', 'snapshot'] as const,
    users: (params?: Record<string, unknown>) => ['iam', 'admin', 'users', params ?? {}] as const,
    userMemberships: (userId: string) => ['iam', 'admin', 'users', userId, 'memberships'] as const,
    audit: (params?: Record<string, unknown>) => ['iam', 'admin', 'audit', params ?? {}] as const,
    sessions: (params?: Record<string, unknown>) => ['iam', 'admin', 'sessions', params ?? {}] as const,
    sessionsAll: () => ['iam', 'admin', 'sessions'] as const,
    mfaCoverage: () => ['iam', 'admin', 'security', 'mfa-coverage'] as const,
    lockouts: () => ['iam', 'admin', 'security', 'lockouts'] as const,
    securitySignals: (window?: string) => ['iam', 'admin', 'security', 'signals', window ?? '7d'] as const,
  },
  notifications: {
    unreadCount: () => ['notifications', 'unread-count'] as const,
  },
} as const;
