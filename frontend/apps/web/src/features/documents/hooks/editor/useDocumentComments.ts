import { useCallback, useEffect, useMemo, useRef, useState, type Dispatch, type SetStateAction } from 'react';
import type { EditorComment } from '@metaldocs/editor-ui';
import { toast } from 'sonner';
import {
  rowToEditorComment,
  useDocumentCommentMutations,
  useDocumentCommentsQuery,
} from '../../queries/useDocumentCommentsQuery';
import { resolveQueryError } from '../../../../lib/api';

export function useDocumentComments(documentID: string, authorDisplay: string): {
  comments: EditorComment[];
  loading: boolean;
  loadError: string | null;
  add: (c: EditorComment) => Promise<void>;
  resolve: (c: EditorComment) => Promise<void>;
  reopen: (c: EditorComment) => Promise<void>;
  remove: (c: EditorComment) => Promise<void>;
  reply: (replyC: EditorComment, parent: EditorComment) => Promise<void>;
  retry: () => Promise<void>;
  setComments: Dispatch<SetStateAction<EditorComment[]>>;
} {
  const query = useDocumentCommentsQuery(documentID);
  const mutations = useDocumentCommentMutations(documentID, authorDisplay);
  const [localComments, setLocalComments] = useState<EditorComment[] | null>(null);
  const lastErrorMessage = useRef<string | null>(null);

  const serverComments = useMemo(
    () => (query.data ?? []).map(rowToEditorComment),
    [query.data],
  );
  const comments = localComments ?? serverComments;
  const loadError = useMemo(
    () => (query.isError ? resolveQueryError(query.error, 'Falha ao carregar comentários.') : null),
    [query.error, query.isError],
  );

  useEffect(() => {
    if (!loadError) {
      lastErrorMessage.current = null;
      return;
    }
    if (lastErrorMessage.current === loadError) {
      return;
    }
    lastErrorMessage.current = loadError;
    toast.error(loadError);
  }, [loadError]);

  const add = useCallback(async (c: EditorComment) => {
    try {
      await mutations.addMutation.mutateAsync(c);
      setLocalComments(null);
    } catch {
      toast.error('Failed to add comment.');
    }
  }, [mutations.addMutation]);

  const resolve = useCallback(async (c: EditorComment) => {
    try {
      await mutations.resolveMutation.mutateAsync(c);
      setLocalComments(null);
    } catch {
      toast.error('Failed to resolve comment.');
    }
  }, [mutations.resolveMutation]);

  const reopen = useCallback(async (c: EditorComment) => {
    try {
      await mutations.reopenMutation.mutateAsync(c);
      setLocalComments(null);
    } catch {
      toast.error('Failed to reopen comment.');
    }
  }, [mutations.reopenMutation]);

  const remove = useCallback(async (c: EditorComment) => {
    try {
      await mutations.deleteMutation.mutateAsync(c);
      setLocalComments(null);
    } catch {
      toast.error('Failed to delete comment.');
    }
  }, [mutations.deleteMutation]);

  const reply = useCallback(async (replyC: EditorComment, parent: EditorComment) => {
    try {
      await mutations.replyMutation.mutateAsync({ reply: replyC, parent });
      setLocalComments(null);
    } catch {
      toast.error('Failed to reply to comment.');
    }
  }, [mutations.replyMutation]);

  const setComments: Dispatch<SetStateAction<EditorComment[]>> = useCallback((next) => {
    setLocalComments((prev) => {
      const base = prev ?? serverComments;
      return typeof next === 'function' ? next(base) : next;
    });
  }, [serverComments]);

  const retry = useCallback(async () => {
    await query.refetch();
  }, [query]);

  return {
    comments,
    loading: query.isLoading,
    loadError,
    add,
    resolve,
    reopen,
    remove,
    reply,
    retry,
    setComments,
  };
}
