import { describe, expect, test } from 'vitest';
import JSZip from 'jszip';
import { fanout } from '../fanout.js';

async function buildDocxWithHeaderFooter(opts: {
  body: string;
  header: string;
  footer: string;
}): Promise<Uint8Array> {
  const zip = new JSZip();
  const documentXml = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <w:body>
    ${opts.body}
    <w:sectPr>
      <w:headerReference w:type="default" r:id="rId2"/>
      <w:footerReference w:type="default" r:id="rId3"/>
    </w:sectPr>
  </w:body>
</w:document>`;
  const headerXml = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:hdr xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">${opts.header}</w:hdr>`;
  const footerXml = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:ftr xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">${opts.footer}</w:ftr>`;
  const contentTypes = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
  <Override PartName="/word/header1.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.header+xml"/>
  <Override PartName="/word/footer1.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.footer+xml"/>
</Types>`;
  const rels = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`;
  const wordRels = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/header" Target="header1.xml"/>
  <Relationship Id="rId3" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/footer" Target="footer1.xml"/>
</Relationships>`;

  zip.file('[Content_Types].xml', contentTypes);
  zip.file('_rels/.rels', rels);
  zip.file('word/document.xml', documentXml);
  zip.file('word/_rels/document.xml.rels', wordRels);
  zip.file('word/header1.xml', headerXml);
  zip.file('word/footer1.xml', footerXml);

  return zip.generateAsync({ type: 'uint8array' });
}

async function extractPart(docx: Uint8Array, part: string): Promise<string> {
  const zip = await JSZip.loadAsync(docx);
  const file = zip.file(part);
  if (!file) throw new Error(`missing ${part}`);
  return file.async('string');
}

describe('fanout — header/footer substitution', () => {
  test('substitutes {tokens} in native header and footer parts at freeze', async () => {
    const tpl = await buildDocxWithHeaderFooter({
      body: `<w:p><w:r><w:t>Body {doc_code}</w:t></w:r></w:p>`,
      header: `<w:p><w:r><w:t>Codigo: {doc_code}</w:t></w:r></w:p>`,
      footer: `<w:p><w:r><w:t>Elaborado por: {author}</w:t></w:r></w:p>`,
    });

    const result = await fanout({
      bodyDocx: tpl,
      placeholderValues: { doc_code: 'PR-OP-001', author: 'Ana Lima' },
      compositionConfig: { header_sub_blocks: [], footer_sub_blocks: [], sub_block_params: {} },
      resolvedValues: {},
    });

    const header = await extractPart(result.buffer, 'word/header1.xml');
    const footer = await extractPart(result.buffer, 'word/footer1.xml');

    expect(header).toContain('PR-OP-001');
    expect(header).not.toContain('{doc_code}');
    expect(footer).toContain('Ana Lima');
    expect(footer).not.toContain('{author}');
    expect(result.unreplacedVars).not.toContain('doc_code');
    expect(result.unreplacedVars).not.toContain('author');
  });
});
