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
    distribution: {
      summary: (id: string) => ['documents', 'distribution', 'summary', id] as const,
      recipients: (id: string, params: Record<string, unknown> = {}) =>
        ['documents', 'distribution', 'recipients', id, params] as const,
      coverage: (id: string) => ['documents', 'distribution', 'coverage', id] as const,
    },
  },
  inbox: (params: InboxParams = {}) =>
    ['approval', 'inbox', params] as const,
  audit: {
    // GET /audit/events
    recent: (limit = 10) => ['audit', 'recent', limit] as const,
  },
  controlledDocuments: {
    all: ['controlled-documents'] as const,
    list: (filter: ControlledDocFilter = {}) =>
      ['controlled-documents', 'list', filter] as const,
    detail: (id: string) => ['controlled-documents', 'detail', id] as const,
    activeDocument: (id: string) => ['controlled-documents', 'active-document', id] as const,
    preview: (profileCode: string, areaCode: string) =>
      ['controlled-documents', 'preview', profileCode, areaCode] as const,
  },
  taxonomy: {
    all: ['taxonomy'] as const,
    profiles: (params: { includeArchived?: boolean } = {}) =>
      ['taxonomy', 'profiles', params] as const,
    areas: (params: { includeArchived?: boolean } = {}) =>
      ['taxonomy', 'areas', params] as const,
    families: (params: { includeInactive?: boolean } = {}) =>
      ['taxonomy', 'families', params] as const,
  },
  templates: {
    list: () => ['templates', 'list'] as const,
    blank: () => ['templates', 'blank'] as const,
    byProfile: (profileCode: string) =>
      ['templates', 'by-profile', profileCode] as const,
    // Wizard-only: published-only subset for a profile (Step 3 template
    // picker). Distinct key from byProfile/list so it can never collide with
    // the admin management page's all-status cache for the same profile.
    byProfilePublished: (profileCode: string) =>
      ['templates', 'by-profile', profileCode, 'published'] as const,
    placeholderCatalog: () => ['templates', 'placeholder-catalog'] as const,
    detail: (id: string) => ['templates', 'detail', id] as const,
    // SLICE 8b (unit 3.2): submit-time route preview for a template version.
    versionPreview: (templateId: string, versionNum: number) =>
      ['templates', 'version-preview', templateId, versionNum] as const,
  },
  tokens: {
    all: ['tokens'] as const,
    list: () => ['tokens', 'list'] as const,
  },
  approval: {
    // Root for whole-subtree invalidation (covers instance/routes below and
    // the top-level inbox() key, which all share the 'approval' prefix).
    all: ['approval'] as const,
    instance: (documentId: string) =>
      ['approval', 'instance', documentId] as const,
    routes: {
      list: () => ['approval', 'routes', 'list'] as const,
      detail: (id: string) => ['approval', 'routes', 'detail', id] as const,
    },
    // SLICE 8b (unit 3.2): submit-time route preview, read-only. Keyed apart
    // from instance()/routes.* — this is the pre-submit resolution, not the
    // post-submit instance or the admin route catalogue.
    documentPreview: (documentId: string) =>
      ['approval', 'document-preview', documentId] as const,
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
    memberships: {
      all: ['iam', 'admin', 'memberships'] as const,
      list: (params: Record<string, unknown> = {}) =>
        ['iam', 'admin', 'memberships', 'list', params] as const,
      byArea: (areaCode: string) =>
        ['iam', 'admin', 'memberships', 'by-area', areaCode] as const,
      byUser: (userId: string) =>
        ['iam', 'admin', 'memberships', 'by-user', userId] as const,
    },
    audit: (params?: Record<string, unknown>) => ['iam', 'admin', 'audit', params ?? {}] as const,
    sessions: (params?: Record<string, unknown>) => ['iam', 'admin', 'sessions', params ?? {}] as const,
    sessionsAll: () => ['iam', 'admin', 'sessions'] as const,
    mfaCoverage: () => ['iam', 'admin', 'security', 'mfa-coverage'] as const,
    lockouts: () => ['iam', 'admin', 'security', 'lockouts'] as const,
    securitySignals: (window?: string) => ['iam', 'admin', 'security', 'signals', window ?? '7d'] as const,
  },
  notifications: {
    unreadCount: () => ['notifications', 'unread-count'] as const,
    list: (params: { status?: string; limit?: number } = {}) =>
      ['notifications', 'list', params] as const,
  },
} as const;
