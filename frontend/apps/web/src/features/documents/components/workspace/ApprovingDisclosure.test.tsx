import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { ApprovingDisclosure } from './ApprovingDisclosure';

describe('ApprovingDisclosure', () => {
  it('renders the collapsed integrity disclosure summary and no delegation badge by default', () => {
    render(
      <ApprovingDisclosure
        documentId="doc-1"
        contentHash="hash-abc"
        frozenContentHash={null}
        delegatedFrom={null}
      />,
    );
    expect(screen.getByText('Conteúdo verificado ✓ · detalhes')).toBeInTheDocument();
    expect(screen.queryByText(/hash-abc/)).not.toBeInTheDocument();
    expect(screen.queryByTestId('delegation-badge')).not.toBeInTheDocument();
  });

  it('reveals the content hash when the disclosure is expanded', async () => {
    render(
      <ApprovingDisclosure
        documentId="doc-1"
        contentHash="hash-abc"
        frozenContentHash={null}
        delegatedFrom={null}
      />,
    );
    fireEvent.click(screen.getByText('Conteúdo verificado ✓ · detalhes'));
    await waitFor(() => expect(screen.getByText('hash-abc')).toBeInTheDocument());
  });

  it('shows the delegation badge with the delegating actor when signing on delegation', () => {
    render(
      <ApprovingDisclosure
        documentId="doc-1"
        contentHash="hash-abc"
        frozenContentHash={null}
        delegatedFrom="Bruno Said"
      />,
    );
    expect(screen.getByTestId('delegation-badge')).toHaveTextContent('Assinando por delegação de Bruno Said');
  });
});
