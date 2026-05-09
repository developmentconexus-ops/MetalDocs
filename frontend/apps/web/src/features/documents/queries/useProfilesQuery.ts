// Re-export from taxonomy — the query wraps taxonomy API and belongs there.
// Kept as a shim so existing mocks in test files continue to work unchanged.
export { useProfilesQuery } from '../../taxonomy/queries/useProfilesQuery';
