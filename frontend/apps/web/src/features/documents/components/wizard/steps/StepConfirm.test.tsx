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
  governanceClass: 'controlado',
  hasActiveRoute: true,
  hasActiveTemplateRoute: true,
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

const baseProps = {
  profile,
  area,
  title: 'Documento de teste',
  visibility: 'company' as const,
  visibilityAreaCodes: [],
  inviteeCount: 0,
  template: null,
  isBlankTemplateSelected: true,
  blankTemplateName: 'Em branco',
  authorDisplayName: 'Admin',
  createdAt: new Date('2026-05-19T10:00:00-03:00'),
  consent: false,
  submitting: false,
  error: null,
  onConsent: vi.fn(),
  onSubmit: vi.fn(),
  onBack: vi.fn(),
  onCancel: vi.fn(),
  submitDisabled: true,
};

describe('StepConfirm', () => {
  it('shows governed REV00 semantics instead of raw v1 during document creation', () => {
    render(
      <StepConfirm
        {...baseProps}
        previewCode="POP-GENERAL-011"
        previewCodeLoading={false}
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
        {...baseProps}
        previewCode="POP-GENERAL-011"
        previewCodeLoading={false}
        onConsent={onConsent}
      />,
    );

    fireEvent.click(screen.getByRole('checkbox'));
    expect(onConsent).toHaveBeenCalledWith(true);
  });

  it('shows loading affordance (POP-GENERAL-…) when query is in flight and profile+area are set', () => {
    render(
      <StepConfirm
        {...baseProps}
        previewCode={null}
        previewCodeLoading={true}
      />,
    );

    // Should show loading ellipsis form, not a bare ???
    expect(screen.getAllByText('POP-GENERAL-…').length).toBeGreaterThan(0);
    expect(screen.queryByText(/^\?\?\?/)).not.toBeInTheDocument();
  });

  it('shows graceful fallback (POP-GENERAL-???) when query finished with null and profile+area are set', () => {
    render(
      <StepConfirm
        {...baseProps}
        previewCode={null}
        previewCodeLoading={false}
      />,
    );

    // Should show profile-area-??? fallback, not a bare ???-???-???
    expect(screen.getAllByText('POP-GENERAL-???').length).toBeGreaterThan(0);
    expect(screen.queryByText('???-???-???')).not.toBeInTheDocument();
  });

  it('shows ???-???-??? when profile is null (selections incomplete)', () => {
    render(
      <StepConfirm
        {...baseProps}
        profile={null}
        previewCode={null}
        previewCodeLoading={false}
      />,
    );

    // All code slots should show the full not-ready placeholder
    const chips = screen.getAllByText('???-???-???');
    expect(chips.length).toBeGreaterThan(0);
  });
});
