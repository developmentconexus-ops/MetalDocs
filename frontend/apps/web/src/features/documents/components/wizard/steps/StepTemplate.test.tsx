import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { StepTemplate } from './StepTemplate';

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
});
