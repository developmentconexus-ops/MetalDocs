import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { InboxStack } from './InboxStack';
import type { InboxItem } from '../api/approvalTypes';

const items: InboxItem[] = [
  {
    instance_id: 'inst-1',
    document_id: 'doc-1',
    controlled_document_id: 'CD-001',
    document_title: 'Documento 1',
    area_code: 'QA',
    submitted_by: 'user-1',
    submitted_at: new Date().toISOString(),
    stage_label: 'Revisão',
    quorum_progress: '0/1',
  },
];

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
});
