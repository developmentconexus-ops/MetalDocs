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

  it('renders every returned template as selectable, since the backend now only returns published templates (published=true)', () => {
    const onSelect = vi.fn();

    // (a) full published-ref case.
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

    // (b) a second published template to prove multiple cards all render enabled.
    const publishedTwo: TemplateDTO = {
      ...BASE_TEMPLATE,
      id: 'tmpl-pub-2',
      name: 'Template Publicado Dois',
      latest_version: {
        id: 'v-pub-2-latest',
        number: 2,
        revision_number: 2,
        status: 'published',
      },
      published_version: {
        id: '22222222-2222-4222-8222-222222222222',
        number: 2,
        revision_number: 2,
        status: 'published',
      },
    };

    render(
      <StepTemplate
        profileLabel="Perfil PRC"
        templates={[published, publishedTwo]}
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

    const publishedBadges = screen.getAllByText('publicada');
    expect(publishedBadges).toHaveLength(2);

    const publishedCard = screen.getByRole('radio', { name: /Template Publicado(?! Dois)/i });
    expect(publishedCard).toBeEnabled();
    fireEvent.click(publishedCard);
    expect(onSelect).toHaveBeenCalledWith('tmpl-pub', '11111111-1111-4111-8111-111111111111');

    const publishedTwoCard = screen.getByRole('radio', { name: /Template Publicado Dois/i });
    expect(publishedTwoCard).toBeEnabled();
    fireEvent.click(publishedTwoCard);
    expect(onSelect).toHaveBeenCalledWith('tmpl-pub-2', '22222222-2222-4222-8222-222222222222');

    // The blank/start-from-scratch option is always present alongside published templates.
    const blankCard = screen.getByRole('radio', { name: /Em branco/i });
    expect(blankCard).toBeEnabled();
  });
});
