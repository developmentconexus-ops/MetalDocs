import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import { VersionActionPanel } from '../VersionActionPanel';
import { useAuthStore } from '../../../store/auth.store';
import type { VersionDTO } from '../api/templates';
import type { CurrentUser, UserRole } from '../../../lib/types';

function setUser(roles: UserRole[], capabilities: string[]) {
  const user: CurrentUser = {
    userId: 'u1',
    tenantId: 't1',
    tenantName: 'Acme',
    username: 'alice',
    displayName: 'Alice',
    mustChangePassword: false,
    roles,
    capabilities,
  };
  useAuthStore.setState({ user });
}

function inReviewWithReviewer(): VersionDTO {
  return {
    id: 'v1',
    template_id: 't1',
    version_number: 1,
    revision_number: 0,
    status: 'in_review',
    docx_storage_key: null,
    content_hash: null,
    metadata_schema: null,
    placeholder_schema: null,
    author_id: 'author-1',
    pending_reviewer_role: 'reviewer',
    pending_approver_role: 'approver',
    reviewer_id: null,
    approver_id: null,
    submitted_at: null,
    reviewed_at: null,
    approved_at: null,
    published_at: null,
    obsoleted_at: null,
    created_at: '2026-01-01T00:00:00Z',
  };
}

describe('VersionActionPanel gating', () => {
  beforeEach(() => {
    useAuthStore.setState({ user: null });
  });

  it('disables Approve Review with tooltip when actor lacks template.review', () => {
    setUser(['reviewer'], []);
    render(<VersionActionPanel version={inReviewWithReviewer()} onVersionUpdate={() => {}} />);
    const acceptBtn = screen.getByRole('button', { name: /approve review/i });
    expect(acceptBtn).toBeDisabled();
    expect(acceptBtn.getAttribute('title')).toContain('template.review');
  });

  it('disables Approve Review with tooltip when actor lacks pending_reviewer_role', () => {
    setUser(['approver'], ['template.review']);
    render(<VersionActionPanel version={inReviewWithReviewer()} onVersionUpdate={() => {}} />);
    const acceptBtn = screen.getByRole('button', { name: /approve review/i });
    expect(acceptBtn).toBeDisabled();
    expect(acceptBtn.getAttribute('title')).toContain('reviewer');
  });

  it('enables Approve Review when actor has cap + role', () => {
    setUser(['reviewer'], ['template.review']);
    render(<VersionActionPanel version={inReviewWithReviewer()} onVersionUpdate={() => {}} />);
    const acceptBtn = screen.getByRole('button', { name: /approve review/i });
    expect(acceptBtn).not.toBeDisabled();
    expect(acceptBtn.getAttribute('title')).toBeFalsy();
  });
});
