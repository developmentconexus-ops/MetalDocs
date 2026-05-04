import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { RegistryCreateDialog } from './RegistryCreateDialog';

vi.mock('../../store/auth.store', () => ({
  useAuthStore: vi.fn(),
}));
vi.mock('../taxonomy/api', () => ({
  fetchProfiles: vi.fn().mockResolvedValue([]),
  fetchAreas: vi.fn().mockResolvedValue([]),
}));
vi.mock('./api', () => ({
  createControlledDocument: vi.fn(),
}));

import { useAuthStore } from '../../store/auth.store';

beforeEach(() => {
  vi.clearAllMocks();
});

describe('RegistryCreateDialog E12 submit gate', () => {
  it('disables submit and shows placeholder while currentUser is null', () => {
    (useAuthStore as unknown as ReturnType<typeof vi.fn>).mockImplementation(
      (sel: (s: { user: null }) => unknown) => sel({ user: null })
    );
    render(<RegistryCreateDialog onClose={() => {}} onCreated={() => {}} />);

    const btn = screen.getByRole('button', { name: /Aguardando/i });
    expect(btn).toBeTruthy();
    expect((btn as HTMLButtonElement).disabled).toBe(true);

    const input = screen.getByDisplayValue('Aguardando autenticação...');
    expect(input).toBeTruthy();
  });

  it('enables submit when currentUser is ready', () => {
    (useAuthStore as unknown as ReturnType<typeof vi.fn>).mockImplementation(
      (sel: (s: { user: { userId: string; displayName: string } }) => unknown) =>
        sel({ user: { userId: 'u-1', displayName: 'Alice' } })
    );
    render(<RegistryCreateDialog onClose={() => {}} onCreated={() => {}} />);

    const btn = screen.getByRole('button', { name: /^Criar$/i });
    expect(btn).toBeTruthy();
    expect((btn as HTMLButtonElement).disabled).toBe(false);
    expect(screen.getByDisplayValue('Alice')).toBeTruthy();
  });

  it('falls back to userId when displayName is null', () => {
    (useAuthStore as unknown as ReturnType<typeof vi.fn>).mockImplementation(
      (sel: (s: { user: { userId: string; displayName: null } }) => unknown) =>
        sel({ user: { userId: 'u-2', displayName: null } })
    );
    render(<RegistryCreateDialog onClose={() => {}} onCreated={() => {}} />);

    expect(screen.getByDisplayValue('u-2')).toBeTruthy();
    const btn = screen.getByRole('button', { name: /^Criar$/i });
    expect((btn as HTMLButtonElement).disabled).toBe(false);
  });
});
