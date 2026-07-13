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
    capabilities: ['template.submit', 'template.approve', 'template.publish'],
    ...overrides,
  };
}

function makeHandlers(): TemplateApprovalHandlers & {
  runSignoff: ReturnType<typeof vi.fn>;
  runPublish: ReturnType<typeof vi.fn>;
} {
  return {
    runSignoff: vi.fn(),
    runPublish: vi.fn(),
  };
}

describe('buildTemplateApprovalActions', () => {
  describe('draft status', () => {
    // R1: the template editor is the sole author submit surface — the cockpit
    // yields no actions for draft versions.
    it('emits no actions', () => {
      const actions = buildTemplateApprovalActions(makeVersion(), makeActor(), makeHandlers());
      expect(actions).toEqual([]);
    });
  });

  describe('under_review status', () => {
    it('emits accept ("Assinar aprovação") and reject ("Rejeitar")', () => {
      const version = makeVersion({ status: 'under_review' });
      const actions = buildTemplateApprovalActions(version, makeActor(), makeHandlers());

      expect(actions).toHaveLength(2);
      expect(actions[0].key).toBe('accept');
      expect(actions[0].label).toBe('Assinar aprovação');
      expect(actions[0].variant).toBe('primary');
      expect(actions[1].key).toBe('reject');
      expect(actions[1].label).toBe('Rejeitar');
      expect(actions[1].variant).toBe('danger');
    });

    it('both actions available when actor has template.approve', () => {
      const version = makeVersion({ status: 'under_review' });
      const actor = makeActor({ capabilities: ['template.approve'] });
      const actions = buildTemplateApprovalActions(version, actor, makeHandlers());

      expect(actions[0].available).toBe(true);
      expect(actions[1].available).toBe(true);
    });

    it('both actions denied when actor lacks template.approve', () => {
      const version = makeVersion({ status: 'under_review' });
      const actor = makeActor({ capabilities: [] });
      const actions = buildTemplateApprovalActions(version, actor, makeHandlers());

      expect(actions[0].available).toBe(false);
      expect(actions[0].reason).toBeTruthy();
      expect(actions[1].available).toBe(false);
    });

    it('clicking accept calls runSignoff(true)', () => {
      const version = makeVersion({ status: 'under_review' });
      const handlers = makeHandlers();
      const actions = buildTemplateApprovalActions(version, makeActor(), handlers);

      actions[0].run();
      expect(handlers.runSignoff).toHaveBeenCalledWith(true);
      expect(handlers.runPublish).not.toHaveBeenCalled();
    });

    it('clicking reject calls runSignoff(false)', () => {
      const version = makeVersion({ status: 'under_review' });
      const handlers = makeHandlers();
      const actions = buildTemplateApprovalActions(version, makeActor(), handlers);

      actions[1].run();
      expect(handlers.runSignoff).toHaveBeenCalledWith(false);
    });
  });

  describe('approved status', () => {
    it('emits a single "publish" action labeled "Publicar"', () => {
      const version = makeVersion({ status: 'approved' });
      const actions = buildTemplateApprovalActions(version, makeActor(), makeHandlers());

      expect(actions).toHaveLength(1);
      expect(actions[0].key).toBe('publish');
      expect(actions[0].label).toBe('Publicar');
      expect(actions[0].variant).toBe('primary');
      expect(actions[0].available).toBe(true);
    });

    it('denied when actor lacks template.publish (template.approve alone is not enough)', () => {
      const version = makeVersion({ status: 'approved' });
      const actor = makeActor({ capabilities: ['template.approve'] });
      const actions = buildTemplateApprovalActions(version, actor, makeHandlers());

      expect(actions[0].available).toBe(false);
    });

    it('clicking publish calls runPublish()', () => {
      const version = makeVersion({ status: 'approved' });
      const handlers = makeHandlers();
      const actions = buildTemplateApprovalActions(version, makeActor(), handlers);

      actions[0].run();
      expect(handlers.runPublish).toHaveBeenCalledTimes(1);
      expect(handlers.runSignoff).not.toHaveBeenCalled();
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
