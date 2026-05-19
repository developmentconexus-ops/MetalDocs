import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { EditorMetaSidebar } from './EditorMetaSidebar';

describe('EditorMetaSidebar', () => {
  it('renders real governed metadata and hides approval chain in draft', () => {
    render(
      <EditorMetaSidebar
        open
        onToggle={() => {}}
        code="POP-RH-001"
        profileLabel="Procedimento Operacional"
        areaLabel="Recursos Humanos"
        visibilityLabel="Restrito a area Recursos Humanos"
        history={[
          {
            documentId: 'doc-2',
            revisionCode: 'REV01',
            revisionTitle: 'Ajuste operacional',
            status: 'draft',
            createdAt: '2026-05-18T12:00:00Z',
            isCurrent: true,
          },
          {
            documentId: 'doc-1',
            revisionCode: 'REV00',
            revisionTitle: 'Primeira versao',
            status: 'published',
            createdAt: '2026-05-10T12:00:00Z',
            isCurrent: false,
          },
        ]}
        approvalChain={null}
        documentStatus="draft"
      />,
    );

    expect(screen.getByText('Procedimento Operacional')).toBeInTheDocument();
    expect(screen.getByText('Recursos Humanos')).toBeInTheDocument();
    expect(screen.getByText('Restrito a area Recursos Humanos')).toBeInTheDocument();
    expect(screen.getByText('REV01 · Draft')).toBeInTheDocument();
    expect(screen.getByText('Ajuste operacional')).toBeInTheDocument();
    expect(screen.queryByText('Proximos aprovadores')).not.toBeInTheDocument();
  });

  it('renders full approval chain only during under review', () => {
    render(
      <EditorMetaSidebar
        open
        onToggle={vi.fn()}
        history={[]}
        approvalChain={{
          stages: [
            {
              id: 'stage-1',
              label: 'Qualidade',
              status: 'active',
              actors: [
                { user_id: 'user-1', display_name: 'Maria Souza', status: 'approved', decision: 'approve' },
                { user_id: 'user-2', display_name: 'Joao Lima', status: 'active', decision: null },
              ],
              signoffs: [{ id: 'signoff-1', actor_user_id: 'Maria Souza', decision: 'approve' }],
            },
            {
              id: 'stage-2',
              label: 'Diretoria',
              status: 'pending',
              actors: [
                { user_id: 'user-3', display_name: 'Ana Costa', status: 'waiting', decision: null },
              ],
              signoffs: [],
            },
          ],
        }}
        documentStatus="under_review"
      />,
    );

    expect(screen.getByText('Proximos aprovadores')).toBeInTheDocument();
    expect(screen.getByText('Maria Souza')).toBeInTheDocument();
    expect(screen.getByText('Joao Lima')).toBeInTheDocument();
    expect(screen.getByText('Ana Costa')).toBeInTheDocument();
    expect(screen.getByText('aprovou')).toBeInTheDocument();
    expect(screen.getByText('proximo')).toBeInTheDocument();
    expect(screen.getByText('aguarda')).toBeInTheDocument();
  });
});
