type ContentNode = { text?: string; content?: ContentNode[] };

/**
 * Flatten a DocumentCommentResponse.content (ProseMirror JSON node array) into
 * plain display text. Read-only: the cockpit shows comments, it does not edit them.
 */
export function commentPlainText(nodes: unknown): string {
  if (!Array.isArray(nodes)) {
    return '';
  }
  const parts: string[] = [];
  const walk = (node: ContentNode) => {
    if (typeof node?.text === 'string') {
      parts.push(node.text);
    }
    if (Array.isArray(node?.content)) {
      node.content.forEach(walk);
    }
  };
  (nodes as ContentNode[]).forEach(walk);
  return parts.join(' ').trim();
}
