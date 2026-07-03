import { useMutation, useQueryClient } from '@tanstack/react-query';

import {
  archiveArea,
  archiveProfile,
  createArea,
  createFamily,
  createProfile,
  deactivateFamily,
  setDefaultTemplate,
  updateArea,
  updateFamily,
  updateProfile,
} from '../api/taxonomy';
import type {
  CreateAreaRequest,
  CreateFamilyRequest,
  CreateProfileRequest,
  SetDefaultTemplateRequest,
  UpdateAreaRequest,
  UpdateFamilyRequest,
  UpdateProfileRequest,
} from '../types';
import { QK } from '../../../lib/queryKeys';

// Taxonomy admin mutations. Every mutation invalidates the whole 'taxonomy'
// cache subtree on success — the admin page, and every cross-feature picker
// that reads active profiles/areas (documents, templates, approval), must
// see the change immediately rather than serving a stale cached list.
function invalidateTaxonomy(queryClient: ReturnType<typeof useQueryClient>) {
  void queryClient.invalidateQueries({ queryKey: QK.taxonomy.all });
}

export function useCreateProfileMutation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (req: CreateProfileRequest) => createProfile(req),
    onSuccess: () => invalidateTaxonomy(queryClient),
  });
}

export function useUpdateProfileMutation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ code, req }: { code: string; req: UpdateProfileRequest }) => updateProfile(code, req),
    onSuccess: () => invalidateTaxonomy(queryClient),
  });
}

export function useSetDefaultTemplateMutation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ code, req }: { code: string; req: SetDefaultTemplateRequest }) => setDefaultTemplate(code, req),
    onSuccess: () => invalidateTaxonomy(queryClient),
  });
}

export function useArchiveProfileMutation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (code: string) => archiveProfile(code),
    onSuccess: () => invalidateTaxonomy(queryClient),
  });
}

export function useCreateAreaMutation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (req: CreateAreaRequest) => createArea(req),
    onSuccess: () => invalidateTaxonomy(queryClient),
  });
}

export function useUpdateAreaMutation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ code, req }: { code: string; req: UpdateAreaRequest }) => updateArea(code, req),
    onSuccess: () => invalidateTaxonomy(queryClient),
  });
}

export function useArchiveAreaMutation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (code: string) => archiveArea(code),
    onSuccess: () => invalidateTaxonomy(queryClient),
  });
}

export function useCreateFamilyMutation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (req: CreateFamilyRequest) => createFamily(req),
    onSuccess: () => invalidateTaxonomy(queryClient),
  });
}

export function useUpdateFamilyMutation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ code, req }: { code: string; req: UpdateFamilyRequest }) => updateFamily(code, req),
    onSuccess: () => invalidateTaxonomy(queryClient),
  });
}

export function useDeactivateFamilyMutation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (code: string) => deactivateFamily(code),
    onSuccess: () => invalidateTaxonomy(queryClient),
  });
}
