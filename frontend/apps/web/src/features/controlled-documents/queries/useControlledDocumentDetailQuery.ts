import { useQuery } from "@tanstack/react-query";
import { QK } from "../../../lib/queryKeys";
import { fetchControlledDocument } from "../api/controlledDocuments";
import { STALE_FIVE_MINUTES } from "../../documents/queries/_constants";

export function useControlledDocumentDetailQuery(id: string) {
  return useQuery({
    queryKey: QK.controlledDocuments.detail(id),
    queryFn: () => fetchControlledDocument(id),
    enabled: Boolean(id),
    staleTime: STALE_FIVE_MINUTES,
  });
}
