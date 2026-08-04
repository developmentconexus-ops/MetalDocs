import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import React from 'react';

import { TemplateWizardPage } from './TemplateWizardPage';
import { ApiError } from '../../../lib/api';

// ── Mock react-router-dom ─────────────────────────────────────────────────────

const mockNavigate = vi.fn();

// The wizard seeds its step from ?step=N; 4 lands directly on StepConfirmation,
// which is the only step that can surface a create failure.
vi.mock('react-router-dom', () => ({
  useNavigate: () => mockNavigate,
  useSearchParams: () => [new URLSearchParams('step=4'), vi.fn()],
}));

// ── Mock server-state queries ─────────────────────────────────────────────────

vi.mock('../../taxonomy/queries/useProfilesQuery', () => ({
  useProfilesQuery: () => ({
    data: [
      {
        code: 'POP',
        name: 'Procedimento Operacional',
        familyCode: 'FAM',
        hasActiveTemplateRoute: true,
      },
    ],
    isLoading: false,
    isError: false,
    isSuccess: true,
    error: null,
    refetch: vi.fn(),
  }),
}));

// ── Seed the reducer so selectMaxReachableStep() reaches step 4 ───────────────

vi.mock('../state/templateWizard.reducer', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../state/templateWizard.reducer')>();
  return {
    ...actual,
    initialTemplateWizardState: {
      ...actual.initialTemplateWizardState,
      profileCode: 'POP',
      name: 'Meu Template',
      description: 'Descrição do template',
      startingPoint: 'blank' as const,
    },
  };
});

// ── Mock shell + the steps the confirmation path never renders ────────────────

vi.mock('../../shared/components/wizard/WizardShell', () => ({
  WizardShell: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}));

vi.mock('../components/wizard/steps/StepScope', () => ({
  StepScope: () => <div data-testid="step-scope" />,
}));

vi.mock('../components/wizard/steps/StepIdentity', () => ({
  StepIdentity: () => <div data-testid="step-identity" />,
}));

vi.mock('../components/wizard/steps/StepStructure', () => ({
  StepStructure: () => <div data-testid="step-structure" />,
}));

// Minimal confirm UI so a click drives handleSubmit() and submitError is visible.
vi.mock('../components/wizard/steps/StepConfirmation', () => ({
  StepConfirmation: ({
    onSubmit,
    submitError,
  }: {
    onSubmit: () => void;
    submitError?: string | null;
  }) => (
    <div data-testid="step-confirmation">
      <button type="button" onClick={onSubmit}>
        Criar e abrir editor
      </button>
      {submitError ? <div role="alert">{submitError}</div> : null}
    </div>
  ),
}));

// ── Mock the templates API ────────────────────────────────────────────────────

vi.mock('../api/templates', () => ({
  createTemplate: vi.fn(),
  importTemplateDocx: vi.fn(),
}));

import * as templatesApi from '../api/templates';

function renderWizardAtConfirmation() {
  return render(<TemplateWizardPage />);
}

async function submitAndReadAlert(): Promise<string> {
  renderWizardAtConfirmation();
  fireEvent.click(screen.getByRole('button', { name: 'Criar e abrir editor' }));
  const alert = await screen.findByRole('alert');
  return alert.textContent ?? '';
}

beforeEach(() => {
  vi.clearAllMocks();
});

describe('TemplateWizardPage — create-error copy is keyed on the Problem code', () => {
  it('surfaces the duplicate-key copy for conflict.already_exists', async () => {
    vi.mocked(templatesApi.createTemplate).mockRejectedValue(
      new ApiError('conflict.already_exists', 409, 'template key conflict'),
    );

    const text = await submitAndReadAlert();

    expect(text).toBe(
      'Já existe um template com este identificador técnico. Escolha outro nome.',
    );
    expect(mockNavigate).not.toHaveBeenCalled();
  });

  it('surfaces the route-configuration copy for state.approval_route_missing (ADR 0086)', async () => {
    vi.mocked(templatesApi.createTemplate).mockRejectedValue(
      new ApiError('state.approval_route_missing', 409, 'no active template route'),
    );

    const text = await submitAndReadAlert();

    expect(text).toBe(
      'Este perfil não tem rota de aprovação de template ativa. Configure a rota em Rotas de Aprovação antes de criar o template.',
    );
    expect(mockNavigate).not.toHaveBeenCalled();
  });

  it('maps on the code, not on the backend message text', async () => {
    // Same human-readable detail as the wizard-specific cases; only the code
    // differs, so a message-based map would mis-route this one.
    vi.mocked(templatesApi.createTemplate).mockRejectedValue(
      new ApiError('validation.failed', 422, 'template key conflict'),
    );

    const text = await submitAndReadAlert();

    expect(text).toBe('A validação dos dados informados falhou.');
  });

  it('falls through to resolveErrorMessage for a code the wizard does not special-case', async () => {
    vi.mocked(templatesApi.createTemplate).mockRejectedValue(
      new ApiError('TEMPLATE_SOME_UNMAPPED_CODE', 500, 'boom'),
    );

    const text = await submitAndReadAlert();

    expect(text).toBe('Não foi possível concluir a ação. Código: TEMPLATE_SOME_UNMAPPED_CODE');
  });

  it('falls back to the Error message for a non-ApiError failure', async () => {
    vi.mocked(templatesApi.createTemplate).mockRejectedValue(new Error('network down'));

    const text = await submitAndReadAlert();

    expect(text).toBe('network down');
  });

  it('navigates to the editor when create succeeds', async () => {
    vi.mocked(templatesApi.createTemplate).mockResolvedValue({
      template: { id: 'tpl-1' },
      version: { version_number: 1 },
    } as never);

    renderWizardAtConfirmation();
    fireEvent.click(screen.getByRole('button', { name: 'Criar e abrir editor' }));

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/templates/tpl-1/edit');
    });
    expect(screen.queryByRole('alert')).toBeNull();
  });
});
