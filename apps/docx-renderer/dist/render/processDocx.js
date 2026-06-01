import { createHash } from 'node:crypto';
import { processTemplateDetailed } from '@eigenpal/docx-js-editor/headless';
export async function processDocx(templateBuffer, formData) {
    // processTemplateDetailed expects Record<string, string> — coerce values to string
    const variables = {};
    for (const [k, v] of Object.entries(formData)) {
        variables[k] = v == null ? '' : String(v);
    }
    const result = processTemplateDetailed(templateBuffer.buffer.slice(templateBuffer.byteOffset, templateBuffer.byteOffset + templateBuffer.byteLength), variables, { nullGetter: 'empty' });
    const buf = new Uint8Array(result.buffer);
    const contentHash = createHash('sha256').update(buf).digest('hex');
    return {
        buffer: buf,
        contentHash,
        unreplacedVars: result.unreplacedVariables ?? [],
    };
}
