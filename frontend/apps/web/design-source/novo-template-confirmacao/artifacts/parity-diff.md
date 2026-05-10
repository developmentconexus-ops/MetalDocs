# Parity Diff · novo-template-confirmacao · 1440px

| region | field | ref | impl | delta |
|---|---|---:|---:|---:|
| previewCard | grid columns | 120px 1fr | 120px 1fr | 0 |
| previewCard | gap | 18px | 18px | 0 |
| previewCard | padding | 18px | 18px | 0 |
| thumb | size | 120px × 152px | 120px × 152px | 0 |
| thumb | radius | 2px | 2px | 0 |
| thumb | padding | 10px/9px (design) | 8px (sp-2) | top/bottom +2px, lr −1px |
| previewBody | header gap | 8px | 8px | 0 |
| previewBody | header margin-bottom | 8px | 8px | 0 |
| previewBody | template name font-size | 16px | 16px | 0 |
| previewBody | template name margin-bottom | 14px | 14px | 0 |
| metaGrid | columns | 1fr 1fr | 1fr 1fr | 0 |
| metaGrid | gap | 8px | 8px | 0 |
| metaGrid | font-size | 12px | 12px | 0 |
| confirmBlock | padding | 14px | 14px | 0 |
| confirmBlock | margin-bottom | 22px | 22px | 0 |
| confirmBlock | radius | r-2 | r-2 | 0 |
| checkLabel | layout | flex, flex-start | flex, flex-start | 0 |
| checkLabel | gap | 8px (sp-2) | 8px (sp-2) | 0 |
| checkLabel | input margin-top | 2px | 2px | 0 |

**Non-zero deltas:** thumb padding (cosmetic; documented above). All other deltas: 0.

**Note (post-review fix):** `.intro` `margin-bottom` was 20px (cascade override by WizardShell) — fixed with `!important` to restore design 24px (sp-6). Thumb highlight pattern corrected to `Set([1,4,7,10])`. `thumbLine` color changed from `--border-strong` to `--paper-line`. Tier reclassified to Heavy (has `@media`).
