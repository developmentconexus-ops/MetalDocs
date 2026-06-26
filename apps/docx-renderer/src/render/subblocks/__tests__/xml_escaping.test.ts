import { describe, expect, test } from "vitest";
import { escapeXml, xmlText } from "../xml";
import { DocHeaderStandard } from "../doc_header_standard";
import { ApprovalSignaturesBlock } from "../approval_signatures_block";
import { RevisionBox } from "../revision_box";
import { FooterControlledCopyNotice } from "../footer_controlled_copy_notice";

// Sub-block OOXML is raw-injected (docxtemplater `{@...}` token) with NO downstream
// escaping. A tenant value containing OOXML markup must therefore be neutralized at
// the `<w:t>` sink, or it forges live document structure in the frozen, hash-audited
// output (see xml.ts). These tests guard that escaping.

const INJECTION = "</w:t></w:r><w:r><w:t>FORGED";

describe("escapeXml", () => {
  test("escapes the five XML metacharacters", () => {
    expect(escapeXml(`& < > " '`)).toBe("&amp; &lt; &gt; &quot; &apos;");
  });

  test("escapes & first so entities are not double-encoded", () => {
    expect(escapeXml("a < b")).toBe("a &lt; b");
    expect(escapeXml("&lt;")).toBe("&amp;lt;");
  });

  test("xmlText stringifies null/undefined to empty", () => {
    expect(xmlText(null)).toBe("");
    expect(xmlText(undefined)).toBe("");
    expect(xmlText(3)).toBe("3");
  });
});

describe("sub-block OOXML injection is neutralized", () => {
  test("doc_header_standard escapes a malicious title", async () => {
    const ooxml = await DocHeaderStandard.render({
      params: {},
      values: { title: INJECTION, doc_code: "x", effective_date: "x", revision_number: 1 },
    });
    expect(ooxml).not.toContain(INJECTION);
    expect(ooxml).toContain("&lt;/w:t&gt;&lt;/w:r&gt;&lt;w:r&gt;&lt;w:t&gt;FORGED");
  });

  test("approval_signatures_block escapes a malicious display_name", async () => {
    const ooxml = await ApprovalSignaturesBlock.render({
      params: {},
      values: { approvers: [{ display_name: INJECTION, signed_at: "2026-01-01" }] },
    });
    expect(ooxml).not.toContain(INJECTION);
    expect(ooxml).toContain("&lt;/w:t&gt;");
  });

  test("revision_box escapes a malicious description", async () => {
    const ooxml = await RevisionBox.render({
      params: {},
      values: { revision_history: [{ rev: "1", date: "2026-01-01", description: INJECTION }] },
    });
    expect(ooxml).not.toContain(INJECTION);
  });

  test("footer_controlled_copy_notice escapes a malicious notice override", async () => {
    const ooxml = await FooterControlledCopyNotice.render({
      params: { notice_text: INJECTION },
      values: {},
    });
    expect(ooxml).not.toContain(INJECTION);
    expect(ooxml).toContain("&lt;/w:t&gt;");
  });
});
