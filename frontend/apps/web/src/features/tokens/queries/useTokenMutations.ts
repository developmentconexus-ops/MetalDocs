import { useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { QK } from '../../../lib/queryKeys';
import { resolveErrorMessage } from '../../../lib/api';
import { createToken, deleteToken, updateToken } from '../api/tokens';
import type {
  CreateTokenDictionaryEntryRequest,
  UpdateTokenDictionaryEntryRequest,
} from '../api/tokensTypes';

export function useTokenMutations() {
  const qc = useQueryClient();
  const invalidate = () => qc.invalidateQueries({ queryKey: QK.tokens.list() });

  const create = useMutation({
    mutationFn: (req: CreateTokenDictionaryEntryRequest) => createToken(req),
    onSuccess: () => {
      toast.success('Token criado.');
      void invalidate();
    },
    onError: (err) => toast.error(resolveErrorMessage(err)),
  });

  const update = useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateTokenDictionaryEntryRequest }) =>
      updateToken(id, req),
    onSuccess: () => {
      toast.success('Token atualizado.');
      void invalidate();
    },
    onError: (err) => toast.error(resolveErrorMessage(err)),
  });

  const remove = useMutation({
    mutationFn: (id: string) => deleteToken(id),
    onSuccess: () => {
      toast.success('Token removido.');
      void invalidate();
    },
    onError: (err) => toast.error(resolveErrorMessage(err)),
  });

  return { create, update, remove };
}
