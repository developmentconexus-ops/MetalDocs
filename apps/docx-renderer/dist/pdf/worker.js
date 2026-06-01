import { createHash } from 'node:crypto';
const MAX_ATTEMPTS = 3;
const BASE_BACKOFF_MS = 250;
const PDF_CONTENT_TYPE = 'application/pdf';
const DOCX_CONTENT_TYPE = 'application/vnd.openxmlformats-officedocument.wordprocessingml.document';
export async function runPdfJob(input, deps) {
    const docxBuffer = await deps.getObject(input.final_docx_s3_key);
    const pdfBuffer = await convertWithRetry(docxBuffer, deps);
    const pdfKey = `${input.final_docx_s3_key}.pdf`;
    await deps.putObject(pdfKey, pdfBuffer, PDF_CONTENT_TYPE);
    const pdfHash = createHash('sha256').update(pdfBuffer).digest('hex');
    const generatedAt = (deps.now ? deps.now() : new Date()).toISOString();
    return {
        final_pdf_s3_key: pdfKey,
        pdf_hash: pdfHash,
        pdf_generated_at: generatedAt,
    };
}
async function convertWithRetry(docx, deps) {
    const sleep = deps.sleep ?? defaultSleep;
    const url = `${deps.gotenbergUrl.replace(/\/+$/, '')}/forms/libreoffice/convert`;
    let lastErr = null;
    for (let attempt = 1; attempt <= MAX_ATTEMPTS; attempt++) {
        const form = new FormData();
        form.append('files', new Blob([docx], { type: DOCX_CONTENT_TYPE }), 'document.docx');
        const res = await fetch(url, { method: 'POST', body: form });
        if (res.ok) {
            return Buffer.from(await res.arrayBuffer());
        }
        lastErr = new Error(`gotenberg status ${res.status}`);
        if (res.status < 500)
            break;
        if (attempt < MAX_ATTEMPTS) {
            await sleep(BASE_BACKOFF_MS * 2 ** (attempt - 1));
        }
    }
    throw lastErr ?? new Error('gotenberg: unknown failure');
}
async function defaultSleep(ms) {
    await new Promise((resolve) => setTimeout(resolve, ms));
}
