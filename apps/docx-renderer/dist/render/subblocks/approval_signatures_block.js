function str(v) {
    if (v === null || v === undefined)
        return "";
    return String(v);
}
function cell(text) {
    return `<w:tc><w:p><w:r><w:t xml:space="preserve">${text}</w:t></w:r></w:p></w:tc>`;
}
function row(a, b) {
    return `<w:tr>${cell(a)}${cell(b)}</w:tr>`;
}
function isApprover(v) {
    return typeof v === "object" && v !== null;
}
export const ApprovalSignaturesBlock = {
    key: "approval_signatures_block",
    async render(ctx) {
        const raw = ctx.values.approvers;
        const approvers = Array.isArray(raw) ? raw.filter(isApprover) : [];
        const header = row("Name", "Signed At");
        const body = approvers
            .map((a) => row(str(a.display_name), str(a.signed_at)))
            .join("");
        return `<w:tbl>${header}${body}</w:tbl>`;
    },
};
