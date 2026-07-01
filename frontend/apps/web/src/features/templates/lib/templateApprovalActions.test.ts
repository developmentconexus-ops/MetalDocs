import { describe, it, expect, vi } from 'vitest';
import { buildTemplateApprovalActions, type TemplateApprovalHandlers } from './templateApprovalActions';
import type { ActorContext } from './canActOnVersion';
import type { VersionDTO } from '../api/templates';

function makeVersion(overrides: Partial<VersionDTO> = {}): VersionDTO {
  return {
    id: 'v1',
    template_id: 't1',
    version_number: 1,
    revision_number: 0,
    status: 'draft',
    docx_storage_key: null,
    content_hash: null,
    metadata_schema: null,
    placeholder_schema: null,
    author_id: 'author-1',
    pending_reviewer_role: null,
    pending_approver_role: null,
    reviewer_id: null,
    approver_id: null,
    submitted_at: null,
    reviewed_at: null,
    approved_at: null,
    published_at: null,
    obsoleted_at: null,
    created_at: '2026-01-01T00:00:00Z',
    ...overrides,
  };
}

function makeActor(overrides: Partial<ActorContext> = {}): ActorContext {
  return {
    roles: ['reviewer', 'approver'],
    capabilities: ['template.submit', 'template.review', 'template.approve'],
    ...overrides,
  };
}

function makeHandlers(): TemplateApprovalHandlers & {
  runSubmit: ReturnType<typeof vi.fn>;
  runReview: ReturnType<typeof vi.fn>;
  runApprove: ReturnType<typeof vi.fn>;
} {
  return {
    runSubmit: vi.fn(),
    runReview: vi.fn(),
    runApprove: vi.fn(),
  };
}

describe('buildTemplateApprovalActions', () => {
  describe('draft status', () => {
    it('emits exactly one action with key "submit" and correct label', () => {
      const actions = buildTemplateApprovalActions(makeVersion(), makeActor(), makeHandlers());
      expect(actions).toHaveLength(1);
      expect(actions[0].key).toBe('submit');
      expect(actions[0].label).toBe('Submeter para revisão');
      expect(actions[0].variant).toBe('primary');
    });

    it('submit action is available when actor has template.submit', () => {
      const actions = buildTemplateApprovalActions(makeVersion(), makeActor(), makeHandlers());
      expect(actions[0].available).toBe(true);
      expect(actions[0].reason).toBeUndefined();
    });

    it('submit action is unavailable when actor lacks template.submit, with a non-empty reason', () => {
      const actor = makeActor({ capabilities: ['template.review', 'template.approve'] });
      const actions = buildTemplateApprovalActions(makeVersion(), actor, makeHandlers());
      expect(actions[0].available).toBe(false);
      expect(actions[0].reason).toBeTruthy();
    });
  });

  describe('under_review WITH pending_reviewer_role', () => {
    it('emits two actions: accept (label "Aprovar revisão") and reject', () => {
      const version = makeVersion({
        status: 'under_review',
        pending_reviewer_role: 'reviewer',
        pending_approver_role: 'approver',
      });
      const actor = makeActor({ roles: ['reviewer'], capabilities: ['template.review'] });
      const actions = buildTemplateApprovalActions(version, actor, makeHandlers());

      expect(actions).toHaveLength(2);
      expect(actions[0].key).toBe('accept');
      expect(actions[0].label).toBe('Aprovar revisão');
      expect(actions[0].variant).toBe('primary');
      expect(actions[1].key).toBe('reject');
      expect(actions[1].label).toBe('Rejeitar');
      expect(actions[1].variant).toBe('danger');
    });

    it('both actions are available when actor has template.review + matching role', () => {
      const version = makeVersion({
        status: 'under_review',
        pending_reviewer_role: 'reviewer',
      });
      const actor = makeActor({ roles: ['reviewer'], capabilities: ['template.review'] });
      const actions = buildTemplateApprovalActions(version, actor, makeHandlers());

      expect(actions[0].available).toBe(true);
      expect(actions[1].available).toBe(true);
    });

    it('clicking accept calls runReview with true', () => {
      const version = makeVersion({
        status: 'under_review',
        pending_reviewer_role: 'reviewer',
      });
      const actor = makeActor({ roles: ['reviewer'], capabilities: ['template.review'] });
      const handlers = makeHandlers();
      const actions = buildTemplateApprovalActions(version, actor, handlers);

      actions[0].run();
      expect(handlers.runReview).toHaveBeenCalledWith(true);
      expect(handlers.runApprove).not.toHaveBeenCalled();
    });

    it('clicking reject calls runReview with false', () => {
      const version = makeVersion({
        status: 'under_review',
        pending_reviewer_role: 'reviewer',
      });
      const actor = makeActor({ roles: ['reviewer'], capabilities: ['template.review'] });
      const handlers = makeHandlers();
      const actions = buildTemplateApprovalActions(version, actor, handlers);

      actions[1].run();
      expect(handlers.runReview).toHaveBeenCalledWith(false);
    });
  });

  describe('under_review WITHOUT pending_reviewer_role (no-reviewer flow)', () => {
    it('emits accept with label "Publicar" and reject', () => {
      const version = makeVersion({
        status: 'under_review',
        pending_reviewer_role: null,
        pending_approver_role: 'approver',
      });
      const actor = makeActor({ roles: ['approver'], capabilities: ['template.approve'] });
      const actions = buildTemplateApprovalActions(version, actor, makeHandlers());

      expect(actions).toHaveLength(2);
      expect(actions[0].key).toBe('accept');
      expect(actions[0].label).toBe('Publicar');
      expect(actions[0].available).toBe(true);
      expect(actions[1].key).toBe('reject');
    });

    it('clicking accept calls runApprove with true', () => {
      const version = makeVersion({
        status: 'under_review',
        pending_reviewer_role: null,
        pending_approver_role: 'approver',
      });
      const actor = makeActor({ roles: ['approver'], capabilities: ['template.approve'] });
      const handlers = makeHandlers();
      const actions = buildTemplateApprovalActions(version, actor, handlers);

      actions[0].run();
      expect(handlers.runApprove).toHaveBeenCalledWith(true);
      expect(handlers.runReview).not.toHaveBeenCalled();
    });
  });

  describe('approved status', () => {
    it('emits accept with label "Publicar" and reject, both available', () => {
      const version = makeVersion({
        status: 'approved',
        pending_approver_role: 'approver',
      });
      const actor = makeActor({ roles: ['approver'], capabilities: ['template.approve'] });
      const actions = buildTemplateApprovalActions(version, actor, makeHandlers());

      expect(actions).toHaveLength(2);
      expect(actions[0].key).toBe('accept');
      expect(actions[0].label).toBe('Publicar');
      expect(actions[0].available).toBe(true);
      expect(actions[1].key).toBe('reject');
      expect(actions[1].available).toBe(true);
    });

    it('clicking accept calls runApprove with true', () => {
      const version = makeVersion({
        status: 'approved',
        pending_approver_role: 'approver',
      });
      const handlers = makeHandlers();
      const actions = buildTemplateApprovalActions(version, makeActor(), handlers);

      actions[0].run();
      expect(handlers.runApprove).toHaveBeenCalledWith(true);
    });

    it('clicking reject calls runApprove with false', () => {
      const version = makeVersion({
        status: 'approved',
        pending_approver_role: 'approver',
      });
      const handlers = makeHandlers();
      const actions = buildTemplateApprovalActions(version, makeActor(), handlers);

      actions[1].run();
      expect(handlers.runApprove).toHaveBeenCalledWith(false);
    });
  });

  describe('terminal statuses', () => {
    it('published → empty actions array', () => {
      const actions = buildTemplateApprovalActions(
        makeVersion({ status: 'published' }),
        makeActor(),
        makeHandlers(),
      );
      expect(actions).toEqual([]);
    });

    it('obsolete → empty actions array', () => {
      const actions = buildTemplateApprovalActions(
        makeVersion({ status: 'obsolete' }),
        makeActor(),
        makeHandlers(),
      );
      expect(actions).toEqual([]);
    });
  });
});
