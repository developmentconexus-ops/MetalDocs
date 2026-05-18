import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { Comment } from '@metaldocs/editor-ui';
import { QK } from '../../../lib/queryKeys';
import {
  createComment,
  deleteComment,
  listComments,
  patchComment,
  type CommentRow,
} from '../api/documents';

function toInitials(author: string): string {
  return author
    .trim()
    .split(/\s+/)
    .filter(Boolean)
    .map((token) => token[0]?.toUpperCase() ?? '')
    .join('')
    .slice(0, 2);
}

export function rowToLibraryComment(row: CommentRow): Comment {
  return {
    id: row.library_comment_id,
    parentId: row.parent_library_id ?? undefined,
    author: row.author,
    initials: toInitials(row.author),
    date: row.created_at,
    content: row.content as Comment['content'],
    done: row.done,
  };
}

function libraryCommentToPayloadContent(comment: Comment): unknown[] {
  return comment.content as unknown[];
}

export function useDocumentCommentsQuery(documentID: string) {
  return useQuery({
    queryKey: QK.documents.comments(documentID),
    queryFn: () => listComments(documentID),
    enabled: Boolean(documentID),
    staleTime: 30_000,
  });
}

export function useDocumentCommentMutations(documentID: string, authorDisplay: string) {
  const queryClient = useQueryClient();
  const queryKey = QK.documents.comments(documentID);

  const replaceComment = (comment: CommentRow) => {
    queryClient.setQueryData<CommentRow[]>(queryKey, (rows = []) =>
      rows.map((row) => (row.library_comment_id === comment.library_comment_id ? comment : row)),
    );
  };

  const addMutation = useMutation({
    mutationFn: (comment: Comment) => createComment(documentID, {
      library_comment_id: comment.id,
      author_display: authorDisplay,
      content: libraryCommentToPayloadContent(comment),
    }),
    onMutate: async (comment) => {
      await queryClient.cancelQueries({ queryKey });
      const previous = queryClient.getQueryData<CommentRow[]>(queryKey) ?? [];
      queryClient.setQueryData<CommentRow[]>(queryKey, [
        ...previous,
        {
          id: `optimistic-${comment.id}`,
          library_comment_id: comment.id,
          parent_library_id: comment.parentId ?? null,
          author: comment.author,
          author_id: '',
          content: libraryCommentToPayloadContent(comment),
          done: Boolean(comment.done),
          created_at: comment.date ?? new Date().toISOString(),
          updated_at: comment.date ?? new Date().toISOString(),
          resolved_at: comment.done ? new Date().toISOString() : null,
        },
      ]);
      return { previous };
    },
    onError: (_err, _comment, context) => {
      queryClient.setQueryData(queryKey, context?.previous ?? []);
    },
    onSuccess: replaceComment,
    onSettled: () => {
      void queryClient.invalidateQueries({ queryKey });
    },
  });

  const replyMutation = useMutation({
    mutationFn: ({ reply, parent }: { reply: Comment; parent: Comment }) => createComment(documentID, {
      library_comment_id: reply.id,
      parent_library_id: parent.id,
      author_display: authorDisplay,
      content: libraryCommentToPayloadContent(reply),
    }),
    onSuccess: (row) => {
      queryClient.setQueryData<CommentRow[]>(queryKey, (rows = []) => [...rows, row]);
    },
    onSettled: () => {
      void queryClient.invalidateQueries({ queryKey });
    },
  });

  const resolveMutation = useMutation({
    mutationFn: (comment: Comment) => patchComment(documentID, comment.id, { done: true }),
    onSuccess: replaceComment,
    onSettled: () => {
      void queryClient.invalidateQueries({ queryKey });
    },
  });

  const reopenMutation = useMutation({
    mutationFn: (comment: Comment) => patchComment(documentID, comment.id, { done: false }),
    onSuccess: replaceComment,
    onSettled: () => {
      void queryClient.invalidateQueries({ queryKey });
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (comment: Comment) => deleteComment(documentID, comment.id),
    onSuccess: (_void, comment) => {
      queryClient.setQueryData<CommentRow[]>(queryKey, (rows = []) =>
        rows.filter((row) => row.library_comment_id !== comment.id && row.parent_library_id !== comment.id),
      );
    },
    onSettled: () => {
      void queryClient.invalidateQueries({ queryKey });
    },
  });

  return {
    addMutation,
    replyMutation,
    resolveMutation,
    reopenMutation,
    deleteMutation,
  };
}

