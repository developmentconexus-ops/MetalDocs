import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { EditorMetaSidebar } from './EditorMetaSidebar';

describe('EditorMetaSidebar', () => {
  it('renders loading state instead of partial technical metadata during hydration', () => {
    render(
      <EditorMetaSidebar
        open
        onToggle={() => {}}
        loading
        code="POP-GENERAL-011"
      />,
    );

    expect(screen.getByText('Carregando metadados do documento')).toBeInTheDocument();
    expect(screen.queryByText('---')).not.toBeInTheDocument();
    expect(screen.queryByText('pop')).not.toBeInTheDocument();
    expect(screen.queryByText('general')).not.toBeInTheDocument();
  });

  it('renders real governed metadata and hides approval chain in draft', () => {
    render(
      <EditorMetaSidebar
        open
        onToggle={() => {}}
        code="POP-RH-001"
        profileLabel="Procedimento Operacional"
        areaLabel="Recursos Humanos"
        visibilityLabel="Restrito a area Recursos Humanos"
        pageCount={3}
        fileSizeBytes={1304}
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

    expect(screen.getByLabelText('Identificacao do documento')).toBeInTheDocument();
    expect(screen.getByText('Identificacao')).toBeInTheDocument();
    expect(screen.queryByText('Metadados')).not.toBeInTheDocument();
    expect(screen.getByText('Tipo')).toBeInTheDocument();
    expect(screen.getByText('Area responsavel')).toBeInTheDocument();
    expect(screen.getByText('Procedimento Operacional')).toBeInTheDocument();
    expect(screen.getByText('Recursos Humanos')).toBeInTheDocument();
    expect(screen.getByText('Restrito a area Recursos Humanos')).toBeInTheDocument();
    expect(screen.getByText('Paginas')).toBeInTheDocument();
    expect(screen.getByText('3 paginas \u00b7 1,3 KB')).toBeInTheDocument();
    expect(screen.getByText('REV01')).toBeInTheDocument();
    expect(screen.getByText('Ajuste operacional')).toBeInTheDocument();
    expect(screen.queryByText(/Draft|Em revis/i)).not.toBeInTheDocument();
    expect(screen.queryByText('Proximos aprovadores')).not.toBeInTheDocument();
  });

  it('renders revision code, title, and date without workflow status text', () => {
    render(
      <EditorMetaSidebar
        open
        onToggle={() => {}}
        history={[{
          documentId: 'doc-1',
          revisionCode: 'REV00',
          revisionTitle: 'Criacao do documento',
          status: 'draft',
          createdAt: '2026-05-19T10:00:00-03:00',
          isCurrent: true,
        }]}
      />,
    );

    expect(screen.getByText('REV00')).toBeInTheDocument();
    expect(screen.getByText('Criacao do documento')).toBeInTheDocument();
    expect(screen.getByText('19/05/2026')).toBeInTheDocument();
    expect(screen.queryByText(/Draft|Em revis/i)).not.toBeInTheDocument();
  });

  it('keeps a governed dossier fill area when sidebar content is short', () => {
    render(
      <EditorMetaSidebar
        open
        onToggle={() => {}}
        history={[{
          documentId: 'doc-1',
          revisionCode: 'REV00',
          revisionTitle: 'Criacao do documento',
          status: 'draft',
          createdAt: '2026-05-19T10:00:00-03:00',
          isCurrent: true,
        }]}
      />,
    );

    expect(screen.getByText('Dossie governado')).toBeInTheDocument();
  });

  it('collapses long governed histories and can expand them', async () => {
    const history = Array.from({ length: 5 }, (_, index) => ({
      documentId: `doc-${index}`,
      revisionCode: `REV0${index}`,
      revisionTitle: `Revision ${index}`,
      status: 'approved',
      createdAt: `2026-05-${String(10 + index).padStart(2, '0')}T10:00:00-03:00`,
      isCurrent: index === 4,
    }));

    render(<EditorMetaSidebar open onToggle={() => {}} history={history} />);

    expect(screen.getByText('REV04')).toBeInTheDocument();
    expect(screen.queryByText('REV00')).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: /ver todas as revis/i }));
    expect(screen.getByText('REV00')).toBeInTheDocument();
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
