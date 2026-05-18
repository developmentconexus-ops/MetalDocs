import { useCallback, useEffect, useMemo, useRef, useState, type Dispatch, type SetStateAction } from 'react';
import type { Comment } from '@metaldocs/editor-ui';
import { toast } from 'sonner';
import {
  rowToLibraryComment,
  useDocumentCommentMutations,
  useDocumentCommentsQuery,
} from '../../queries/useDocumentCommentsQuery';
import { resolveQueryError } from '../../../../lib/api';

export function partitionByMarkers(comments: Comment[], markerIds: Set<number>): { live: Comment[]; orphans: Comment[] } {
  const live: Comment[] = [];
  const orphans: Comment[] = [];
  for (const comment of comments) {
    if (markerIds.has(comment.id)) live.push(comment);
    else orphans.push(comment);
  }
  return { live, orphans };
}

export function useDocumentComments(documentID: string, authorDisplay: string): {
  comments: Comment[];
  orphans: Comment[];
  loading: boolean;
  loadError: string | null;
  add: (c: Comment) => Promise<void>;
  resolve: (c: Comment) => Promise<void>;
  reopen: (c: Comment) => Promise<void>;
  remove: (c: Comment) => Promise<void>;
  reply: (replyC: Comment, parent: Comment) => Promise<void>;
  retry: () => Promise<void>;
  setComments: Dispatch<SetStateAction<Comment[]>>;
} {
  const query = useDocumentCommentsQuery(documentID);
  const mutations = useDocumentCommentMutations(documentID, authorDisplay);
  const [localComments, setLocalComments] = useState<Comment[] | null>(null);
  const lastErrorMessage = useRef<string | null>(null);

  const serverComments = useMemo(
    () => (query.data ?? []).map(rowToLibraryComment),
    [query.data],
  );
  const comments = localComments ?? serverComments;
  const orphans = useMemo(() => [] as Comment[], []);
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

  const add = useCallback(async (c: Comment) => {
    try {
      await mutations.addMutation.mutateAsync(c);
      setLocalComments(null);
    } catch {
      toast.error('Failed to add comment.');
    }
  }, [mutations.addMutation]);

  const resolve = useCallback(async (c: Comment) => {
    try {
      await mutations.resolveMutation.mutateAsync(c);
      setLocalComments(null);
    } catch {
      toast.error('Failed to resolve comment.');
    }
  }, [mutations.resolveMutation]);

  const reopen = useCallback(async (c: Comment) => {
    try {
      await mutations.reopenMutation.mutateAsync(c);
      setLocalComments(null);
    } catch {
      toast.error('Failed to reopen comment.');
    }
  }, [mutations.reopenMutation]);

  const remove = useCallback(async (c: Comment) => {
    try {
      await mutations.deleteMutation.mutateAsync(c);
      setLocalComments(null);
    } catch {
      toast.error('Failed to delete comment.');
    }
  }, [mutations.deleteMutation]);

  const reply = useCallback(async (replyC: Comment, parent: Comment) => {
    try {
      await mutations.replyMutation.mutateAsync({ reply: replyC, parent });
      setLocalComments(null);
    } catch {
      toast.error('Failed to reply to comment.');
    }
  }, [mutations.replyMutation]);

  const setComments: Dispatch<SetStateAction<Comment[]>> = useCallback((next) => {
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
    orphans,
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
