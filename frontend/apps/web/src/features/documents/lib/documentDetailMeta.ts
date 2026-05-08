// Resolves profile code snapshot → human label using profiles list
export function resolveProfileLabel(code: string, profiles: Array<{ code: string; name: string }>): string {
  return profiles.find((p) => p.code === code)?.name ?? code;
}

// Resolves area code snapshot → human label using areas list
export function resolveAreaLabel(code: string, areas: Array<{ code: string; name: string }>): string {
  return areas.find((a) => a.code === code)?.name ?? code;
}

// Signoff actor status → display config
export type SignoffStatus = 'pending' | 'approved' | 'rejected' | 'abstained';

export const SIGNOFF_STATUS_META: Record<SignoffStatus, { label: string; className: string }> = {
  pending:   { label: 'Aguardando', className: 'pending'  },
  approved:  { label: 'Aprovado',   className: 'approved' },
  rejected:  { label: 'Rejeitado',  className: 'rejected' },
  abstained: { label: 'Abstido',    className: 'abstained'},
};
