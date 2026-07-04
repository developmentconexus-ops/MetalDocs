import { describe, expect, it } from 'vitest';

import { getDocumentStatusPresentation } from './documentStatusPresentation';

// Guards the governed per-status copy for the document view model. This copy is a
// UX contract lifted byte-for-byte from DocumentPublishedPage; a regression in any
// badge/subtitle/ownerMeta string must fail here rather than hide in the adapter.

const PUB = '08 de maio de 2026';
const EM_DASH = '—';

describe('getDocumentStatusPresentation', () => {
  it('approved: subtitle and ownerMeta carry the publish date', () => {
    expect(getDocumentStatusPresentation('approved', PUB)).toEqual({
      badgeLabel: 'aprovado',
      subtitle: `aprovado em ${PUB}`,
      ownerMeta: `aprovado em ${PUB}`,
    });
  });

  it('approved without publish date falls back to bare labels', () => {
    expect(getDocumentStatusPresentation('approved', EM_DASH)).toEqual({
      badgeLabel: 'aprovado',
      subtitle: 'Aprovado',
      ownerMeta: 'aprovado',
    });
  });

  it('scheduled: ownerMeta is DISTINCT from subtitle (governed requirement)', () => {
    const r = getDocumentStatusPresentation('scheduled', PUB);
    expect(r.badgeLabel).toBe('agendado');
    expect(r.subtitle).toBe(`aprovação concluída em ${PUB}`);
    expect(r.ownerMeta).toBe(`publicação agendada · aprovado em ${PUB}`);
    expect(r.ownerMeta).not.toBe(r.subtitle);
  });

  it('published: subtitle and ownerMeta carry the publish date', () => {
    expect(getDocumentStatusPresentation('published', PUB)).toEqual({
      badgeLabel: 'publicado',
      subtitle: `publicado em ${PUB}`,
      ownerMeta: `publicado em ${PUB}`,
    });
  });

  it('superseded: ownerMeta is DISTINCT from subtitle', () => {
    const r = getDocumentStatusPresentation('superseded', PUB);
    expect(r.badgeLabel).toBe('substituído');
    expect(r.subtitle).toBe(`publicado em ${PUB}`);
    expect(r.ownerMeta).toBe(`substituído · publicado em ${PUB}`);
    expect(r.ownerMeta).not.toBe(r.subtitle);
  });

  it('obsolete carries the publish date when present', () => {
    const r = getDocumentStatusPresentation('obsolete', PUB);
    expect(r.badgeLabel).toBe('obsoleto');
    expect(r.subtitle).toBe(`obsoleto após publicação em ${PUB}`);
    expect(r.ownerMeta).toBe(`obsoleto · publicado em ${PUB}`);
  });

  it.each([
    ['draft', 'rascunho', 'Rascunho — ainda não publicado', 'rascunho'],
    ['under_review', 'em revisão', 'Aguardando decisão de aprovação', 'em revisão'],
  ])('%s renders fixed pre-publication copy', (status, badgeLabel, subtitle, ownerMeta) => {
    expect(getDocumentStatusPresentation(status, EM_DASH)).toEqual({ badgeLabel, subtitle, ownerMeta });
  });

  it('rejected (M4 F4.1: removed document-status) falls through to the unknown-status branch', () => {
    // documents.status 'rejected' is dead (no producer). It is no longer a
    // recognized document status and must echo the raw value like any other
    // unknown status, rather than keep bespoke governed copy. Unrelated to the
    // LIVE approval-decision 'rejected' under features/approval/*.
    expect(getDocumentStatusPresentation('rejected', EM_DASH)).toEqual({
      badgeLabel: 'rejected',
      subtitle: null,
      ownerMeta: 'rejected',
    });
  });

  it('unknown status with publish date echoes a publish subtitle', () => {
    expect(getDocumentStatusPresentation('mystery', PUB)).toEqual({
      badgeLabel: 'mystery',
      subtitle: `publicado em ${PUB}`,
      ownerMeta: `publicado em ${PUB}`,
    });
  });

  it('unknown status without publish date echoes the raw status', () => {
    expect(getDocumentStatusPresentation('mystery', EM_DASH)).toEqual({
      badgeLabel: 'mystery',
      subtitle: null,
      ownerMeta: 'mystery',
    });
  });

  it('empty status yields the sem-status fallback', () => {
    expect(getDocumentStatusPresentation('', EM_DASH)).toEqual({
      badgeLabel: 'sem status',
      subtitle: null,
      ownerMeta: EM_DASH,
    });
  });
});
