import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { StepTemplate } from './StepTemplate';
import type { TemplateDTO } from '../../../../templates';

const BASE_TEMPLATE = {
  id: 'tmpl-base',
  name: 'Template Sem Chave',
  description: null,
  doc_type_code: null,
  created_by: 'user-1',
  created_at: '2026-01-01T00:00:00Z',
  updated_at: '2026-01-01T00:00:00Z',
  latest_version: {
    id: 'v-base',
    number: 1,
    revision_number: 0,
    status: 'draft',
  },
  published_version: null,
  archived_at: null,
} as unknown as TemplateDTO;

describe('StepTemplate', () => {
  it('shows and selects blank template when profile has no published templates', () => {
    const onSelect = vi.fn();

    render(
      <StepTemplate
        profileLabel="Perfil PRC"
        templates={[]}
        isLoading={false}
        isError={false}
        error={null}
        onRetry={() => {}}
        selectedTemplateID={null}
        selectedVersionID={null}
        blankTemplateID="blank-template"
        blankTemplateVersionID="blank-tv-1"
        blankTemplateName="Em branco"
        onSelect={onSelect}
        onAdvance={() => {}}
        onBack={() => {}}
        onCancel={() => {}}
        advanceDisabled={true}
      />,
    );

    expect(screen.getByText('Nenhum template publicado para este perfil.')).toBeTruthy();
    const blankCard = screen.getByRole('radio', { name: /Em branco/i });
    expect(blankCard).toBeEnabled();

    fireEvent.click(blankCard);
    expect(onSelect).toHaveBeenCalledWith('blank-template', 'blank-tv-1');
  });

  it('treats templates without a published_version as unselectable, including wire-drift payloads missing the key entirely (bug 2026-07-03)', () => {
    const onSelect = vi.fn();

    // (d) simulates a legacy/drifted wire payload where the backend omits the
    // key altogether — the generated TemplateDTO type has published_version as
    // required-and-nullable, so this cast is deliberate: it forces the shape
    // past the compiler to prove the component still guards against
    // `undefined` at runtime (documents the "never absent" contract).
    const missingKey = { ...BASE_TEMPLATE } as unknown as Record<string, unknown>;
    delete missingKey.published_version;
    const withoutKey = missingKey as unknown as TemplateDTO;

    // (a) published_version: null case.
    const withNull: TemplateDTO = {
      ...BASE_TEMPLATE,
      id: 'tmpl-null',
      name: 'Template Null',
      published_version: null,
    };

    // (b) full published-ref case.
    const published: TemplateDTO = {
      ...BASE_TEMPLATE,
      id: 'tmpl-pub',
      name: 'Template Publicado',
      latest_version: {
        id: 'v-pub-latest',
        number: 1,
        revision_number: 1,
        status: 'published',
      },
      published_version: {
        id: '11111111-1111-4111-8111-111111111111',
        number: 1,
        revision_number: 1,
        status: 'published',
      },
    };

    // (c) latest_version.status: 'under_review' badge case.
    const underReview: TemplateDTO = {
      ...BASE_TEMPLATE,
      id: 'tmpl-review',
      name: 'Template Em Revisao',
      latest_version: {
        id: 'v-review',
        number: 2,
        revision_number: 0,
        status: 'under_review',
      },
      published_version: null,
    };

    render(
      <StepTemplate
        profileLabel="Perfil PRC"
        templates={[withoutKey, withNull, published, underReview]}
        isLoading={false}
        isError={false}
        error={null}
        onRetry={() => {}}
        selectedTemplateID={null}
        selectedVersionID={null}
        blankTemplateID="blank-template"
        blankTemplateVersionID="blank-tv-1"
        blankTemplateName="Em branco"
        onSelect={onSelect}
        onAdvance={() => {}}
        onBack={() => {}}
        onCancel={() => {}}
        advanceDisabled={true}
      />,
    );

    const disabledBadges = screen.getAllByText('sem versão publicada');
    expect(disabledBadges).toHaveLength(2);

    const missingKeyCard = screen.getByRole('radio', { name: /Template Sem Chave/i });
    expect(missingKeyCard).toBeDisabled();
    fireEvent.click(missingKeyCard);
    expect(onSelect).not.toHaveBeenCalled();

    const nullCard = screen.getByRole('radio', { name: /Template Null/i });
    expect(nullCard).toBeDisabled();
    fireEvent.click(nullCard);
    expect(onSelect).not.toHaveBeenCalled();

    const reviewCard = screen.getByRole('radio', { name: /Template Em Revisao/i });
    expect(reviewCard).toBeDisabled();
    expect(screen.getByText('em revisão')).toBeTruthy();

    const publishedBadge = screen.getByText('publicada');
    expect(publishedBadge).toBeTruthy();

    const publishedCard = screen.getByRole('radio', { name: /Template Publicado/i });
    expect(publishedCard).toBeEnabled();
    fireEvent.click(publishedCard);
    expect(onSelect).toHaveBeenCalledWith('tmpl-pub', '11111111-1111-4111-8111-111111111111');
  });
});
