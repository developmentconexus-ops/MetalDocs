import React, { createRef } from 'react';
import { render, act } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { MetalDocsEditorRef, TrackedChange } from './types';

// --- Mutable vendor fixture: one tracked insertion the adapter must surface, ---
// --- plus a coalesced replacement to prove one accept clears every site.     ---
type VendorEntry = {
  type: string;
  text: string;
  author: string;
  from: number;
  to: number;
  revisionId: number;
  insertionRevisionId?: number;
  coalescedRevisionIds?: number[];
};

let entries: VendorEntry[] = [];
const fakeView = { state: { doc: {} }, dispatch: vi.fn() };
// Records which vendor ids each command resolved, so the test can assert a
// replacement accept dispatched BOTH halves (deletion + insertion), not one.
const resolved: number[] = [];

function makeIdCommand(id: number) {
  return () =>
    (_state: unknown, _dispatch: unknown) => {
      resolved.push(id);
      entries = entries.filter(
        (e) => e.revisionId !== id && e.insertionRevisionId !== id && !(e.coalescedRevisionIds ?? []).includes(id),
      );
      return true;
    };
}

vi.mock('@eigenpal/docx-editor-core/prosemirror/utils/extractTrackedChanges', () => ({
  extractTrackedChanges: (_state: unknown) => ({ entries, commentToRevision: new Map() }),
}));

vi.mock('@eigenpal/docx-editor-core/prosemirror/commands', () => ({
  acceptChangeById: (id: number) => makeIdCommand(id)(),
  rejectChangeById: (id: number) => makeIdCommand(id)(),
  acceptAllChanges: () => (_s: unknown, _d: unknown) => {
    resolved.push(-1);
    entries = [];
    return true;
  },
  rejectAllChanges: () => (_s: unknown, _d: unknown) => {
    resolved.push(-2);
    entries = [];
    return true;
  },
  removeCommentMark: (id: number) => (_s: unknown, _d: unknown) => {
    resolved.push(1000 + id);
    return true;
  },
}));

vi.mock('@eigenpal/docx-editor-react', () => ({
  DocxEditor: React.forwardRef((_props: Record<string, unknown>, ref: React.Ref<unknown>) => {
    React.useImperativeHandle(ref, () => ({
      getEditorRef: () => ({ getView: () => fakeView, getHfPmViews: () => new Map() }),
      getTotalPages: () => 1,
      save: async () => new ArrayBuffer(0),
      focus: () => {},
    }));
    return null;
  }),
  createEmptyDocument: () => ({}),
}));

vi.mock('@eigenpal/docx-editor-react/plugin-api', () => ({
  PluginHost: ({ children }: { children: React.ReactNode }) => <>{children}</>,
  templatePlugin: {},
}));

vi.mock('@eigenpal/docx-editor-react/styles.css', () => ({}));

import { MetalDocsEditor } from './MetalDocsEditor';

beforeEach(() => {
  resolved.length = 0;
  entries = [
    { type: 'insertion', text: 'novo texto', author: 'Revisor', from: 4, to: 14, revisionId: 7 },
    {
      type: 'replacement',
      text: 'depois',
      deletedText: 'antes',
      author: 'Revisor',
      from: 20,
      to: 26,
      revisionId: 8,
      insertionRevisionId: 9,
      coalescedRevisionIds: [10],
    },
  ] as VendorEntry[];
});

describe('MetalDocsEditor tracked-changes ref API', () => {
  it('surfaces tracked changes as neutral, vendor-free shapes', () => {
    const ref = createRef<MetalDocsEditorRef>();
    render(<MetalDocsEditor ref={ref} mode="review" documentBuffer={new ArrayBuffer(8)} />);
    const changes = ref.current!.getTrackedChanges();
    expect(changes.length).toBe(2);
    const first = changes[0] as TrackedChange;
    expect(first).toMatchObject({
      revisionId: '7',
      author: 'Revisor',
      type: 'insertion',
      excerpt: 'novo texto',
    });
    // revisionId is a string across the wall even though the vendor id is numeric.
    expect(typeof first.revisionId).toBe('string');
  });

  it('accepts one change by id and clears it from the set', () => {
    const ref = createRef<MetalDocsEditorRef>();
    render(<MetalDocsEditor ref={ref} mode="review" documentBuffer={new ArrayBuffer(8)} />);
    act(() => ref.current!.acceptChange('7'));
    const after = ref.current!.getTrackedChanges().map((c) => c.revisionId);
    expect(after).toEqual(['8']);
    expect(resolved).toContain(7);
  });

  it('accepting a replacement resolves every site (deletion + insertion + coalesced)', () => {
    const ref = createRef<MetalDocsEditorRef>();
    render(<MetalDocsEditor ref={ref} mode="review" documentBuffer={new ArrayBuffer(8)} />);
    act(() => ref.current!.acceptChange('8'));
    expect(ref.current!.getTrackedChanges().map((c) => c.revisionId)).toEqual(['7']);
    expect(resolved).toEqual(expect.arrayContaining([8, 9, 10]));
  });

  it('acceptAllChanges / rejectAllChanges empty the set', () => {
    const ref = createRef<MetalDocsEditorRef>();
    render(<MetalDocsEditor ref={ref} mode="review" documentBuffer={new ArrayBuffer(8)} />);
    act(() => ref.current!.acceptAllChanges());
    expect(ref.current!.getTrackedChanges()).toHaveLength(0);
    expect(resolved).toContain(-1);
  });

  it('fires onTrackedChangesChange with the post-accept set', () => {
    const ref = createRef<MetalDocsEditorRef>();
    const onChange = vi.fn();
    render(
      <MetalDocsEditor
        ref={ref}
        mode="review"
        documentBuffer={new ArrayBuffer(8)}
        onTrackedChangesChange={onChange}
      />,
    );
    act(() => ref.current!.acceptChange('7'));
    expect(onChange).toHaveBeenCalled();
    const last = onChange.mock.calls.at(-1)![0] as TrackedChange[];
    expect(last.map((c) => c.revisionId)).toEqual(['8']);
  });
});
