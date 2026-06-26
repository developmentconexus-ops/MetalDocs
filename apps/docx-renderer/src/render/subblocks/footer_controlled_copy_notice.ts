import { SubBlockRenderer, SubBlockContext } from "./registry";
import { xmlText as str } from "./xml";

const DEFAULT_NOTICE = "CONTROLLED COPY — WHEN PRINTED";

export const FooterControlledCopyNotice: SubBlockRenderer = {
  key: "footer_controlled_copy_notice",
  async render(ctx: SubBlockContext): Promise<string> {
    const override = ctx.params.notice_text;
    const text = override === undefined || override === null || override === ""
      ? DEFAULT_NOTICE
      : str(override);
    return `<w:p><w:r><w:t xml:space="preserve">${text}</w:t></w:r></w:p>`;
  },
};
