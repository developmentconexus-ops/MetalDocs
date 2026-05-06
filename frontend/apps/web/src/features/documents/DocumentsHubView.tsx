// DocumentsHubView is being replaced by the Library screen.
// See: wiki/implementation/screen-redesign-tracker.md
const Stub = (_props?: Record<string, unknown>) => (
  <div style={{ padding: 40, color: 'var(--text-muted)', fontFamily: 'var(--font-sans)' }}>
    Library — em desenvolvimento
  </div>
);

// Named export preserved so existing imports don't break during transition.
export { Stub as DocumentsHubView };
export default Stub;
