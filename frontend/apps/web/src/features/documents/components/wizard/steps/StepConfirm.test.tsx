import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { StepConfirm } from './StepConfirm';
import type { DocumentProfile, ProcessArea } from '../../../../taxonomy/types';

const profile: DocumentProfile = {
  code: 'POP',
  tenantId: 'tenant-1',
  familyCode: 'SOP',
  name: 'Procedimento Operacional',
  description: 'Perfil de procedimento operacional',
  reviewIntervalDays: 365,
  defaultTemplateVersionId: null,
  ownerUserId: null,
  editableByRole: 'admin',
  archivedAt: null,
  createdAt: '2026-05-19T10:00:00-03:00',
};

const area: ProcessArea = {
  code: 'GENERAL',
  tenantId: 'tenant-1',
  name: 'Geral',
  description: 'Area geral',
  parentCode: null,
  ownerUserId: null,
  defaultApproverRole: null,
  archivedAt: null,
  createdAt: '2026-05-19T10:00:00-03:00',
};

describe('StepConfirm', () => {
  it('shows governed REV00 semantics instead of raw v1 during document creation', () => {
    render(
      <StepConfirm
        profile={profile}
        area={area}
        title="Documento de teste"
        visibility="company"
        visibilityAreaCodes={[]}
        inviteeCount={0}
        template={null}
        isBlankTemplateSelected
        blankTemplateName="Em branco"
        previewCode="POP-GENERAL-011"
        authorDisplayName="Admin"
        createdAt={new Date('2026-05-19T10:00:00-03:00')}
        consent={false}
        submitting={false}
        error={null}
        onConsent={vi.fn()}
        onSubmit={vi.fn()}
        onBack={vi.fn()}
        onCancel={vi.fn()}
        submitDisabled
      />,
    );

    expect(screen.getByText('REV00')).toBeInTheDocument();
    expect(screen.queryByText(/^v1$/i)).not.toBeInTheDocument();
    expect(screen.getByText('Documento de teste')).toBeInTheDocument();
    expect(screen.getByText('POP-GENERAL-011 REV00')).toBeInTheDocument();
  });

  it('calls onConsent when the user toggles the confirmation checkbox', () => {
    const onConsent = vi.fn();

    render(
      <StepConfirm
        profile={profile}
        area={area}
        title="Documento de teste"
        visibility="company"
        visibilityAreaCodes={[]}
        inviteeCount={0}
        template={null}
        isBlankTemplateSelected
        blankTemplateName="Em branco"
        previewCode="POP-GENERAL-011"
        authorDisplayName="Admin"
        createdAt={new Date('2026-05-19T10:00:00-03:00')}
        consent={false}
        submitting={false}
        error={null}
        onConsent={onConsent}
        onSubmit={vi.fn()}
        onBack={vi.fn()}
        onCancel={vi.fn()}
        submitDisabled
      />,
    );

    fireEvent.click(screen.getByRole('checkbox'));
    expect(onConsent).toHaveBeenCalledWith(true);
  });
});
