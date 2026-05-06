// TODO(novo-documento): backend has no `visibility` field on `controlled_documents`
// today. UI captures the choice but does NOT submit it.
// See wiki/backlog/novo-documento.md#visibility for the backend prereq.

import type { IconName } from '../../../components/ui/Icon';

export const VISIBILITY_KEYS = ['area', 'people', 'company', 'external'] as const;

export type VisibilityKey = (typeof VISIBILITY_KEYS)[number];

// Icon names use the existing `IconName` union from `components/ui/Icon.tsx`.
// Mapping picked 2026-05-06 (option B in Phase 3b open question §3b.2):
// avoids extending the icon set; reuses the design's JSX reference set.
//   area     → taxonomy   (area / scope)
//   people   → users      (group of people)
//   company  → home       (company-wide)
//   external → link       (external sharing link)
export type VisibilityMetaEntry = {
  label: string;
  icon: IconName;
  description: string;
};

export const VISIBILITY_META: Record<VisibilityKey, VisibilityMetaEntry> = {
  area: {
    label: 'Apenas minha área',
    icon: 'taxonomy',
    description: 'Visível somente para colaboradores da área deste documento.',
  },
  people: {
    label: 'Pessoas específicas',
    icon: 'users',
    description: 'Convide pessoas individualmente. Cada convidado vê o documento.',
  },
  company: {
    label: 'Toda empresa',
    icon: 'home',
    description: 'Padrão para documentos do SGQ. Visível para todos os colaboradores.',
  },
  external: {
    label: 'Compartilhamento externo',
    icon: 'link',
    description: 'Gera link com senha, marca d’água e expiração configuráveis.',
  },
};
