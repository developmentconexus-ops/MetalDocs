import { test, expect, type Page } from '@playwright/test';

// ---------------------------------------------------------------------------
// Shared fixtures
// ---------------------------------------------------------------------------

const TPL_ID = 'tpl-e2e-001';

const draftVersion = {
  id: 'ver-e2e-001',
  template_id: TPL_ID,
  version_number: 1,
  status: 'draft',
  docx_storage_key: null,
  content_hash: null,
  metadata_schema: null,
  placeholder_schema: null,
  author_id: 'user-1',
  reviewer_id: null,
  approver_id: null,
  submitted_at: null,
  reviewed_at: null,
  approved_at: null,
  published_at: null,
  obsoleted_at: null,
  created_at: '2026-04-20T00:00:00Z',
};

const inReviewVersion = { ...draftVersion, status: 'under_review', submitted_at: '2026-04-20T01:00:00Z' };
const approvedVersion = { ...inReviewVersion, status: 'approved', reviewer_id: 'user-1', reviewed_at: '2026-04-20T02:00:00Z' };
const publishedVersion = { ...approvedVersion, status: 'published', approver_id: 'user-1', published_at: '2026-04-20T03:00:00Z' };

const template = {
  id: TPL_ID,
  tenant_id: 'tenant-1',
  doc_type_code: null,
  key: 'po-e2e',
  name: 'E2E Purchase Order',
  description: null,
  areas: [],
  visibility: 'public',
  specific_areas: [],
  latest_version: 1,
  published_version_id: null,
  created_by: 'user-1',
  created_at: '2026-04-20T00:00:00Z',
  archived_at: null,
};

async function mockBaseAPIs(page: Page, versionOverride: Record<string, unknown> = draftVersion) {
  // Feature flags
  await page.route('**/api/v1/feature-flags', (r) =>
    r.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ flags: {} }) }),
  );
  // Auth session
  await page.route('**/api/v1/auth/session', (r) =>
    r.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        user_id: 'user-1',
        email: 'test@example.com',
        roles: ['admin', 'quality'],
        // template.review named a capability from the retired role-branched
        // review/approve split; the approval-kernel gates are template.submit
        // (draft submit), template.approve (under_review signoff),
        // template.publish (approved publish). See canActOnVersion.ts.
        capabilities: ['template.submit', 'template.approve', 'template.publish'],
        tenant_id: 'tenant-1',
      }),
    }),
  );
  // Templates list — match with or without query string
  await page.route(/\/api\/v1\/templates(\?.*)?$/, (r) => {
    if (r.request().method() === 'GET') {
      r.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          data: { templates: [template] },
          meta: { limit: 50, offset: 0 },
        }),
      });
    } else {
      r.continue();
    }
  });
  // Single template
  await page.route(`**/api/v1/templates/${TPL_ID}`, (r) => {
    if (r.request().method() === 'GET') {
      r.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          data: { template, latest_version: versionOverride },
        }),
      });
    } else {
      r.continue();
    }
  });
  // Version by number
  await page.route(`**/api/v1/templates/${TPL_ID}/versions/${versionOverride.version_number}`, (r) =>
    r.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ data: { version: versionOverride } }),
    }),
  );
  // DOCX download URL — key is null, return 404 so useTemplateDraft skips fetch
  await page.route(`**/api/v1/templates/${TPL_ID}/versions/*/docx-url`, (r) =>
    r.fulfill({
      status: 404,
      contentType: 'application/json',
      body: JSON.stringify({ error: { message: 'no upload yet' } }),
    }),
  );
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

test('golden path: create → author → submit → approve review → publish', async ({ page }) => {
  await mockBaseAPIs(page);

  // Create template
  await page.route(/\/api\/v2\/templates(\?.*)?$/, async (r) => {
    if (r.request().method() === 'POST') {
      r.fulfill({
        status: 201,
        contentType: 'application/json',
        body: JSON.stringify({ data: { template, version: draftVersion } }),
      });
    } else {
      r.continue();
    }
  });

  // NOTE (unit 3.1a S3 sweep): route paths below were renamed to match the
  // approval-kernel endpoints (submit-for-approval / signoff / publish); the
  // response envelope shapes were left as the pre-existing {data:{version}}
  // fixtures. This test was already broken before the kernel rebuild — it
  // asserts English button text ("submit for review", "approve review",
  // "publish") and a bespoke "reviewer actions"/"approver actions" panel that
  // don't exist in the current pt-BR ArtifactApprovalScreen-based UI, and its
  // create call targets a nonexistent /api/v2/templates. Full repair is out of
  // scope for this FE-only slice; left broken pending a dedicated e2e pass.

  // Submit → under_review
  await page.route(`**/api/v1/templates/${TPL_ID}/versions/1/submit-for-approval`, (r) =>
    r.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ data: { version: inReviewVersion } }),
    }),
  );

  // Reviewer signs off → approved
  await page.route(`**/api/v1/templates/${TPL_ID}/versions/1/signoff`, (r) =>
    r.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ data: { version: approvedVersion } }),
    }),
  );

  // Publish → published
  await page.route(`**/api/v1/templates/${TPL_ID}/versions/1/publish`, (r) =>
    r.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ data: { version: publishedVersion } }),
    }),
  );

  await page.goto('/templates');
  await expect(page.getByText('E2E Purchase Order')).toBeVisible({ timeout: 10_000 });

  // Open create dialog and submit
  await page.getByRole('button', { name: /new template/i }).click();
  await page.getByLabel(/key/i).fill('po-e2e');
  await page.getByLabel(/name/i).fill('E2E Purchase Order');
  const approverField = page.getByLabel(/approver role/i);
  if (await approverField.isVisible()) {
    await approverField.fill('quality');
  }
  await page.getByRole('button', { name: /^create$/i }).click();

  // Author page
  await expect(page.getByRole('heading', { name: /E2E Purchase Order/i })).toBeVisible({ timeout: 10_000 });
  await expect(page.getByText(/draft/i)).toBeVisible();

  // Submit for review
  const submitBtn = page.getByRole('button', { name: /submit for review/i });
  await expect(submitBtn).toBeEnabled();
  await submitBtn.click();

  // under_review — reviewer panel visible, submit button gone
  await expect(page.getByText(/reviewer actions/i)).toBeVisible({ timeout: 5_000 });
  await expect(submitBtn).not.toBeVisible();

  // Approve review
  await page.getByRole('button', { name: /approve review/i }).click();
  await expect(page.getByText(/review approved/i)).toBeVisible({ timeout: 5_000 });

  // approved — approver panel visible
  await expect(page.getByText(/approver actions/i)).toBeVisible({ timeout: 5_000 });

  // Publish
  await page.getByRole('button', { name: /^publish$/i }).click();
  await expect(page.getByText(/this version is published/i)).toBeVisible({ timeout: 5_000 });
});

test('legacy /templates redirect lands on templates list', async ({ page }) => {
  await mockBaseAPIs(page);

  await page.goto('/templates');
  await expect(page.getByText('E2E Purchase Order')).toBeVisible({ timeout: 10_000 });
});

test('P10.1a strict-lock: under_review version — no submit button, reviewer actions visible', async ({ page }) => {
  await mockBaseAPIs(page, inReviewVersion);

  await page.goto('/templates');
  await expect(page.getByText('E2E Purchase Order')).toBeVisible({ timeout: 10_000 });

  // Open the template from the list
  await page.getByRole('button', { name: /^open$/i }).click();

  // Author page loads in under_review state
  await expect(page.getByRole('heading', { name: /E2E Purchase Order/i })).toBeVisible({ timeout: 10_000 });
  await expect(page.getByText(/em\s*revis/i)).toBeVisible();

  // Submit button must NOT appear for non-draft versions
  await expect(page.getByRole('button', { name: /submit for review/i })).not.toBeVisible();

  // Reviewer actions panel must be visible
  await expect(page.getByText(/reviewer actions/i)).toBeVisible();
});

test('P10.1b strict-lock: published version — published banner, no action buttons', async ({ page }) => {
  await mockBaseAPIs(page, publishedVersion);

  await page.goto('/templates');
  await expect(page.getByText('E2E Purchase Order')).toBeVisible({ timeout: 10_000 });

  await page.getByRole('button', { name: /^open$/i }).click();

  await expect(page.getByRole('heading', { name: /E2E Purchase Order/i })).toBeVisible({ timeout: 10_000 });
  await expect(page.getByText(/this version is published/i)).toBeVisible();
  await expect(page.getByRole('button', { name: /submit for review/i })).not.toBeVisible();
  await expect(page.getByRole('button', { name: /approve review/i })).not.toBeVisible();
  await expect(page.getByRole('button', { name: /^publish$/i })).not.toBeVisible();
});
