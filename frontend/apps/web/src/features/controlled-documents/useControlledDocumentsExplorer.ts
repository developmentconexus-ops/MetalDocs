// Controlled-documents explorer will be rebuilt using TanStack Query.
// This stub exists only so that imports in WorkspaceRoot don't break the build.
export function useControlledDocumentsExplorer(_onRefresh?: () => Promise<void> | void) {
  return {
    documentProfiles: [] as unknown[],
    processAreas: [] as unknown[],
    documentDepartments: [] as unknown[],
    subjects: [] as unknown[],
    selectedProfileSchema: null,
    selectedProfileSchemas: [] as unknown[],
    selectedProfileGovernance: null,
    loadDocumentProfiles: async () => {},
    loadProcessAreas: async () => {},
    loadDocumentDepartments: async () => {},
    handleProfileSelect: async (_code: string) => {},
  };
}
