import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { InboxStack } from './InboxStack';
import type { InboxItem } from '../api/approvalTypes';

const items: InboxItem[] = [
  {
    instance_id: 'inst-1',
    subject_kind: 'document',
    subject_key: 'doc-1',
    subject_title: 'Documento 1',
    subject_ref: 'doc-1',
    controlled_document_id: '11111111-1111-1111-1111-111111111111',
    controlled_document_code: 'POP-QUA-001',
    area_code: 'QA',
    submitted_by: 'user-1',
    submitted_at: new Date().toISOString(),
    stage_label: 'Revisão',
    quorum_progress: '0/1',
    stage_kind: 'review',
    due_at: '2026-07-01T00:00:00Z',
  },
];

const templateItem: InboxItem = {
  instance_id: 'inst-tpl-1',
  subject_kind: 'template',
  subject_key: 'tpl-version-1',
  subject_title: 'Modelo POP',
  subject_ref: 'tpl-1',
  controlled_document_id: null,
  controlled_document_code: null,
  area_code: 'QA',
  submitted_by: 'user-2',
  submitted_at: new Date().toISOString(),
  stage_label: 'Revisão',
  quorum_progress: '0/1',
  stage_kind: 'review',
  due_at: '2026-07-01T00:00:00Z',
};

// FE-19 N8: A/D keyboard shortcuts were removed — they were empty TODO
// handlers wired to a live keydown listener with no effect, while the hint
// row advertised working "Aprovar"/"Devolver" shortcuts. Only ←/→ navigation
// is real; approve/reject stay reachable via the card buttons (onApprove/onReject).
describe('InboxStack', () => {
  it('does not advertise A/D shortcuts in the keyboard hint row', () => {
    const { container } = render(
      <InboxStack
        items={items}
        selectedIdx={0}
        onSelect={() => {}}
        onNext={() => {}}
        onPrev={() => {}}
      />,
    );

    // Scope to the hint row itself — InboxApprovalCard renders its own real
    // "Aprovar e assinar" button, which must stay untouched.
    const hints = container.querySelector('kbd')?.parentElement;
    expect(hints?.textContent).not.toMatch(/Aprovar/);
    expect(hints?.textContent).not.toMatch(/Devolver/);
    expect(hints?.textContent).toMatch(/Navegar/);
  });

  it('does not call onApprove/onReject when A or D is pressed (no live shortcut)', () => {
    const onApprove = vi.fn();
    const onReject = vi.fn();
    render(
      <InboxStack
        items={items}
        selectedIdx={0}
        onSelect={() => {}}
        onNext={() => {}}
        onPrev={() => {}}
        onApprove={onApprove}
        onReject={onReject}
      />,
    );

    fireEvent.keyDown(window, { key: 'a' });
    fireEvent.keyDown(window, { key: 'd' });

    expect(onApprove).not.toHaveBeenCalled();
    expect(onReject).not.toHaveBeenCalled();
  });

  it('still navigates with ArrowLeft/ArrowRight', () => {
    const onNext = vi.fn();
    const onPrev = vi.fn();
    render(
      <InboxStack
        items={items}
        selectedIdx={0}
        onSelect={() => {}}
        onNext={onNext}
        onPrev={onPrev}
      />,
    );

    fireEvent.keyDown(window, { key: 'ArrowRight' });
    fireEvent.keyDown(window, { key: 'ArrowLeft' });

    expect(onNext).toHaveBeenCalledOnce();
    expect(onPrev).toHaveBeenCalledOnce();
  });

  // Unit 4.2 slice 2: a template row must render subject_title and a
  // subject-aware queue code without crashing on the null
  // controlled_document_id.
  it('renders a template row using subject_title and subject_ref, no crash on null controlled_document_id', () => {
    render(
      <InboxStack
        items={[templateItem]}
        selectedIdx={0}
        onSelect={() => {}}
        onNext={() => {}}
        onPrev={() => {}}
      />,
    );

    expect(screen.getAllByText('Modelo POP').length).toBeGreaterThan(0);
    expect(screen.getAllByText('tpl-1').length).toBeGreaterThan(0);
  });

  // F-QA4-8: a document row identifies itself by the canonical human code
  // (controlled_documents.code), never by the controlled-document uuid.
  it('renders the controlled-document code, not the uuid, for a document row', () => {
    render(
      <InboxStack
        items={items}
        selectedIdx={0}
        onSelect={() => {}}
        onNext={() => {}}
        onPrev={() => {}}
      />,
    );

    expect(screen.getAllByText('POP-QUA-001').length).toBeGreaterThan(0);
    expect(screen.queryByText('11111111-1111-1111-1111-111111111111')).toBeNull();
  });
});
