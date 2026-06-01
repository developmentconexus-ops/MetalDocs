function str(v) {
    if (v === null || v === undefined)
        return "";
    return String(v);
}
function cell(text) {
    return `<w:tc><w:p><w:r><w:t xml:space="preserve">${text}</w:t></w:r></w:p></w:tc>`;
}
function row(a, b, c) {
    return `<w:tr>${cell(a)}${cell(b)}${cell(c)}</w:tr>`;
}
function isRevisionEntry(v) {
    return typeof v === "object" && v !== null;
}
export const RevisionBox = {
    key: "revision_box",
    async render(ctx) {
        const raw = ctx.values.revision_history;
        const entries = Array.isArray(raw) ? raw.filter(isRevisionEntry) : [];
        const header = row("Rev", "Date", "Description");
        const body = entries.length === 0
            ? row("—", "—", "—")
            : entries.map((e) => row(str(e.rev), str(e.date), str(e.description))).join("");
        return `<w:tbl>${header}${body}</w:tbl>`;
    },
};
