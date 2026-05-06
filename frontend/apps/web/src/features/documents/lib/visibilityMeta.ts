// TODO(novo-documento): backend has no `visibility` field on `controlled_documents`
// today. UI captures the choice but does NOT submit it.
// See wiki/backlog/novo-documento.md#visibility for the backend prereq.

export const VISIBILITY_KEYS = ['area', 'people', 'company', 'external'] as const;

export type VisibilityKey = (typeof VISIBILITY_KEYS)[number];

// Icon names referenced from `components/ui/Icon.tsx`. Phase 3 of the
// novo-documento rollout will add `user-plus`, `building`, and
// `external-link` to the Icon set; until then, these strings are stored
// as data and any consumer that doesn't have them yet should fall back.
export type VisibilityIconName = 'users' | 'user-plus' | 'building' | 'external-link';

export type VisibilityMetaEntry = {
  label: string;
  icon: VisibilityIconName;
  description: string;
};

export const VISIBILITY_META: Record<VisibilityKey, VisibilityMetaEntry> = {
  area: {
    label: 'Apenas minha área',
    icon: 'users',
    description: 'Visível somente para colaboradores da área deste documento.',
  },
  people: {
    label: 'Pessoas específicas',
    icon: 'user-plus',
    description: 'Convide pessoas individualmente. Cada convidado vê o documento.',
  },
  company: {
    label: 'Toda empresa',
    icon: 'building',
    description: 'Padrão para documentos do SGQ. Visível para todos os colaboradores.',
  },
  external: {
    label: 'Compartilhamento externo',
    icon: 'external-link',
    description: 'Gera link com senha, marca d’água e expiração configuráveis.',
  },
};
