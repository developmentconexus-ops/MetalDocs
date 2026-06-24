import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { EditorComment } from '@metaldocs/editor-ui';
import { QK } from '../../../lib/queryKeys';
import {
  createComment,
  deleteComment,
  listComments,
  patchComment,
  type DocumentComment,
  type DocumentCommentCreateRequest,
} from '../api/documents';

type DocumentCommentContent = DocumentCommentCreateRequest['content'];

export function rowToEditorComment(row: DocumentComment): EditorComment {
  return {
    id: row.library_comment_id,
    parentId: row.parent_library_id ?? undefined,
    author: row.author,
    createdAt: row.created_at,
    body: row.content as unknown,
    resolved: row.done,
  };
}

function libraryCommentToPayloadContent(comment: EditorComment): DocumentCommentContent {
  return comment.body as unknown as DocumentCommentContent;
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

  const replaceComment = (comment: DocumentComment) => {
    queryClient.setQueryData<DocumentComment[]>(queryKey, (rows = []) =>
      rows.map((row) => (row.library_comment_id === comment.library_comment_id ? comment : row)),
    );
  };

  const addMutation = useMutation({
    mutationFn: (comment: EditorComment) => createComment(documentID, {
      library_comment_id: comment.id,
      author_display: authorDisplay,
      content: libraryCommentToPayloadContent(comment) as DocumentCommentCreateRequest['content'],
    }),
    onMutate: async (comment) => {
      await queryClient.cancelQueries({ queryKey });
      const previous = queryClient.getQueryData<DocumentComment[]>(queryKey) ?? [];
      queryClient.setQueryData<DocumentComment[]>(queryKey, [
        ...previous,
        {
          id: `optimistic-${comment.id}`,
          library_comment_id: comment.id,
          parent_library_id: comment.parentId ?? null,
          author: comment.author,
          author_id: '',
          content: libraryCommentToPayloadContent(comment),
          done: Boolean(comment.resolved),
          created_at: comment.createdAt ?? new Date().toISOString(),
          updated_at: comment.createdAt ?? new Date().toISOString(),
          resolved_at: comment.resolved ? new Date().toISOString() : null,
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
    mutationFn: ({ reply, parent }: { reply: EditorComment; parent: EditorComment }) => createComment(documentID, {
      library_comment_id: reply.id,
      parent_library_id: parent.id,
      author_display: authorDisplay,
      content: libraryCommentToPayloadContent(reply) as DocumentCommentCreateRequest['content'],
    }),
    onSuccess: (row) => {
      queryClient.setQueryData<DocumentComment[]>(queryKey, (rows = []) => [...rows, row]);
    },
    onSettled: () => {
      void queryClient.invalidateQueries({ queryKey });
    },
  });

  const resolveMutation = useMutation({
    mutationFn: (comment: EditorComment) => patchComment(documentID, comment.id, { done: true }),
    onSuccess: replaceComment,
    onSettled: () => {
      void queryClient.invalidateQueries({ queryKey });
    },
  });

  const reopenMutation = useMutation({
    mutationFn: (comment: EditorComment) => patchComment(documentID, comment.id, { done: false }),
    onSuccess: replaceComment,
    onSettled: () => {
      void queryClient.invalidateQueries({ queryKey });
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (comment: EditorComment) => deleteComment(documentID, comment.id),
    onSuccess: (_void, comment) => {
      queryClient.setQueryData<DocumentComment[]>(queryKey, (rows = []) =>
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
