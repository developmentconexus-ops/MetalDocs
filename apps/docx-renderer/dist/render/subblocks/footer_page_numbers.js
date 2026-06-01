export const FooterPageNumbers = {
    key: "footer_page_numbers",
    async render() {
        return (`<w:p><w:r><w:t xml:space="preserve">Page </w:t></w:r>` +
            `<w:fldSimple w:instr="PAGE"><w:r><w:t>1</w:t></w:r></w:fldSimple>` +
            `<w:r><w:t xml:space="preserve"> of </w:t></w:r>` +
            `<w:fldSimple w:instr="NUMPAGES"><w:r><w:t>1</w:t></w:r></w:fldSimple></w:p>`);
    },
};
