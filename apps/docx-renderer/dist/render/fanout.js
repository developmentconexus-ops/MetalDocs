import { createHash } from 'node:crypto';
import { processTemplateDetailed } from '@eigenpal/docx-js-editor/headless';
import { SubBlockRegistry } from './subblocks/registry.js';
import { registerV1Builtins } from './subblocks/builtins.js';
export async function fanout(input) {
    const subReg = new SubBlockRegistry();
    registerV1Builtins(subReg);
    const headerOoxml = (await Promise.all(input.compositionConfig.header_sub_blocks.map((k) => subReg.render(k, {
        params: input.compositionConfig.sub_block_params[k] ?? {},
        values: input.resolvedValues,
    })))).join('');
    const footerOoxml = (await Promise.all(input.compositionConfig.footer_sub_blocks.map((k) => subReg.render(k, {
        params: input.compositionConfig.sub_block_params[k] ?? {},
        values: input.resolvedValues,
    })))).join('');
    const variables = {
        ...input.placeholderValues,
        __header_composition__: headerOoxml,
        __footer_composition__: footerOoxml,
    };
    const result = processTemplateDetailed(input.bodyDocx.buffer.slice(input.bodyDocx.byteOffset, input.bodyDocx.byteOffset + input.bodyDocx.byteLength), variables, { nullGetter: 'empty' });
    const buf = new Uint8Array(result.buffer);
    const contentHash = createHash('sha256').update(buf).digest('hex');
    return {
        buffer: buf,
        contentHash,
        unreplacedVars: result.unreplacedVariables ?? [],
    };
}
