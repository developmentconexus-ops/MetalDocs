import { Plugin } from '@tiptap/pm/state';

/**
 * filterTransactionGuard — blocks edits to nodes marked `sdtContentLocked`
 * (locked structured-document-tag content) in template-draft mode.
 * Raw ProseMirror plugin; registered internally by MetalDocsEditor for
 * mode==='template-draft'. Kept inside the wall so the vendor plugin API
 * never leaks to the app.
 */
export function filterTransactionGuard() {
  return new Plugin({
    filterTransaction(tr, state) {
      if (!tr.docChanged) return true;
      let allowed = true;
      tr.steps.forEach((_, i) => {
        const map = tr.mapping.maps[i];
        map.forEach((oldStart, oldEnd) => {
          state.doc.nodesBetween(oldStart, Math.min(oldEnd, state.doc.content.size), (node) => {
            if (node.attrs?.sdtLock === 'sdtContentLocked') {
              allowed = false;
              return false;
            }
            return true;
          });
        });
      });
      return allowed;
    },
  });
}
