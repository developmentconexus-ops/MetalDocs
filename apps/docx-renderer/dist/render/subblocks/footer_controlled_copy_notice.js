const DEFAULT_NOTICE = "CONTROLLED COPY — WHEN PRINTED";
function str(v) {
    if (v === null || v === undefined)
        return "";
    return String(v);
}
export const FooterControlledCopyNotice = {
    key: "footer_controlled_copy_notice",
    async render(ctx) {
        const override = ctx.params.notice_text;
        const text = override === undefined || override === null || override === ""
            ? DEFAULT_NOTICE
            : str(override);
        return `<w:p><w:r><w:t xml:space="preserve">${text}</w:t></w:r></w:p>`;
    },
};
