import { render } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

const navigateMock = vi.fn();
const documentEditorPageMock = vi.fn();

vi.mock('react-router-dom', () => ({
  useNavigate: () => navigateMock,
  useParams: () => ({ documentId: 'doc-modern-1' }),
}));

vi.mock('./DocumentEditorPage', () => ({
  DocumentEditorPage: (props: { documentID: string; onDone: () => void }) => {
    documentEditorPageMock(props);
    return null;
  },
}));

import { Component as DocumentEditorRoutePage } from './DocumentEditorRoutePage';

describe('DocumentEditorRoutePage', () => {
  it('keeps onDone navigation inside the modern document editor route', () => {
    render(<DocumentEditorRoutePage />);

    expect(documentEditorPageMock).toHaveBeenCalledWith(
      expect.objectContaining({
        documentID: 'doc-modern-1',
        onDone: expect.any(Function),
      }),
    );

    const props = documentEditorPageMock.mock.calls[0][0] as {
      onDone: () => void;
    };

    props.onDone();

    expect(navigateMock).toHaveBeenCalledWith('/documents/doc-modern-1/edit', { replace: true });
    expect(navigateMock).not.toHaveBeenCalledWith('/controlled-documents');
  });
});
