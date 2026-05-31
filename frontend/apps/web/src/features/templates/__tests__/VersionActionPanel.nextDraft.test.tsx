import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest';
import { VersionActionPanel } from '../VersionActionPanel';
import type { VersionDTO } from '../api/templates';
import { useAuthStore } from '../../../store/auth.store';

vi.mock('../api/templates', async () => {
  const actual = await vi.importActual<typeof import('../api/templates')>('../api/templates');
  return {
    ...actual,
    approveVersion: vi.fn(),
    reviewVersion: vi.fn(),
  };
});

import { approveVersion } from '../api/templates';

const approveMock = approveVersion as unknown as ReturnType<typeof vi.fn>;

function seedActor(roles: string[], capabilities: string[]): void {
  useAuthStore.setState({
    user: {
      userId: 'u-1',
      tenantId: 't-1',
      tenantName: 'Test Tenant',
      username: 'tester',
      displayName: 'Tester',
      mustChangePassword: false,
      roles,
      capabilities,
    },
  } as never);
}

function makeVersion(overrides: Partial<VersionDTO> = {}): VersionDTO {
  return {
    id: 'ver-1',
    template_id: 'tpl-1',
    version_number: 1,
    revision_number: 0,
    status: 'in_review',
    docx_storage_key: null,
    content_hash: null,
    metadata_schema: null,
    placeholder_schema: null,
    author_id: 'author-1',
    pending_reviewer_role: null,
    pending_approver_role: 'approver',
    reviewer_id: null,
    approver_id: null,
    submitted_at: null,
    reviewed_at: null,
    approved_at: null,
    published_at: null,
    obsoleted_at: null,
    created_at: new Date().toISOString(),
    ...overrides,
  };
}

beforeEach(() => {
  approveMock.mockReset();
  seedActor(['approver'], ['template.approve', 'template.publish']);
});

afterEach(() => {
  vi.clearAllMocks();
});

describe('VersionActionPanel — approve→publish next_draft handoff', () => {
  it('forwards nextDraft to onVersionUpdate on no-reviewer approve and shows publish toast', async () => {
    const v = makeVersion();
    const published = { ...v, status: 'published' as const };
    approveMock.mockResolvedValueOnce({
      version: published,
      nextDraft: { id: 'ver-2', versionNumber: 2 },
    });
    const onVersionUpdate = vi.fn();

    render(<VersionActionPanel version={v} onVersionUpdate={onVersionUpdate} />);
    fireEvent.click(screen.getByRole('button', { name: 'Publish' }));

    await waitFor(() => expect(onVersionUpdate).toHaveBeenCalledTimes(1));
    expect(approveMock).toHaveBeenCalledWith('tpl-1', 1, true, '');
    expect(onVersionUpdate).toHaveBeenCalledWith(published, { id: 'ver-2', versionNumber: 2 });
    expect(await screen.findByText('Publicado. Nova versão de rascunho criada.')).toBeInTheDocument();
  });

  it('passes nextDraft=null on reject and keeps generic message', async () => {
    const v = makeVersion();
    const rejected = { ...v, status: 'draft' as const };
    approveMock.mockResolvedValueOnce({ version: rejected, nextDraft: null });
    const onVersionUpdate = vi.fn();

    render(<VersionActionPanel version={v} onVersionUpdate={onVersionUpdate} />);
    fireEvent.click(screen.getByRole('button', { name: 'Reject' }));

    await waitFor(() => expect(onVersionUpdate).toHaveBeenCalledTimes(1));
    expect(onVersionUpdate).toHaveBeenCalledWith(rejected, null);
    expect(screen.queryByText('Publicado. Nova versão de rascunho criada.')).not.toBeInTheDocument();
  });
});
