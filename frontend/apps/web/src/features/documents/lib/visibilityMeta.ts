import type { IconName } from '../../../components/ui/Icon';

export const VISIBILITY_KEYS = ['area', 'people', 'company', 'external'] as const;

export type VisibilityKey = (typeof VISIBILITY_KEYS)[number];

export type VisibilityMetaEntry = {
  label: string;
  icon: IconName;
  description: string;
};

export const VISIBILITY_META: Record<VisibilityKey, VisibilityMetaEntry> = {
  area: {
    label: 'Area do documento',
    icon: 'taxonomy',
    description: 'Restrito a area selecionada neste documento.',
  },
  people: {
    label: 'Pessoas especificas',
    icon: 'users',
    description: 'Restrito por pessoas e/ou areas especificas.',
  },
  company: {
    label: 'Toda empresa',
    icon: 'home',
    description: 'Visivel para todos os colaboradores da empresa.',
  },
  external: {
    label: 'Compartilhamento externo (em breve)',
    icon: 'link',
    description: 'Nao disponivel no fluxo atual de criacao.',
  },
};
