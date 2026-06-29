import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import * as tokensApi from '../api/tokens';
import { Component as TokensRoutePage } from './TokensRoutePage';

vi.mock('../../templates', () => ({
  usePlaceholderCatalogQuery: () => ({ data: [{ key: 'author', label: 'Autor', description: '' }] }),
}));
const hasCapMock = vi.fn();
vi.mock('../../iam/hooks/useHasCapability', () => ({ useHasCapability: (c: string) => hasCapMock(c) }));

afterEach(() => vi.restoreAllMocks());

function wrapper({ children }: { children: ReactNode }) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return <QueryClientProvider client={qc}>{children}</QueryClientProvider>;
}

describe('TokensRoutePage', () => {
  it('lists entries and shows the New button for managers', async () => {
    hasCapMock.mockReturnValue(true);
    vi.spyOn(tokensApi, 'listTokens').mockResolvedValue([
      { id: '1', name: 'company_slogan', value: 'v', label: 'Slogan', description: '', created_at: 'x', updated_at: 'y' },
    ]);
    render(<TokensRoutePage />, { wrapper });
    await waitFor(() => expect(screen.getByText('company_slogan')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'Novo token' })).toBeInTheDocument();
  });

  it('hides the New button without the manage capability', async () => {
    hasCapMock.mockReturnValue(false);
    vi.spyOn(tokensApi, 'listTokens').mockResolvedValue([]);
    render(<TokensRoutePage />, { wrapper });
    await waitFor(() => expect(screen.getByText('Nenhum token cadastrado.')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'Novo token' })).not.toBeInTheDocument();
  });
});
