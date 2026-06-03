import { resolveQueryError } from "../../../lib/api/resolveQueryError";
import { useUserMembershipsQuery } from "../queries/useUserMembershipsQuery";
import { SP_DATE_FORMATTER } from "../constants";
import styles from "./UserMembershipsTable.module.css";

interface UserMembershipsTableProps {
  userId: string;
}

export default function UserMembershipsTable({ userId }: UserMembershipsTableProps) {
  const { data, isLoading, error } = useUserMembershipsQuery(userId);
  const items = data?.items ?? [];

  if (isLoading) {
    return (
      <div className={styles.empty} aria-busy="true">
        Carregando áreas…
      </div>
    );
  }

  if (error) {
    return (
      <div role="alert" className={styles.error}>
        {resolveQueryError(error, "Não foi possível carregar as áreas do usuário.")}
      </div>
    );
  }

  if (items.length === 0) {
    return (
      <div className={styles.empty}>
        Este usuário ainda não pertence a nenhuma área.
      </div>
    );
  }

  return (
    <div className={styles.wrapper}>
      <table className={styles.table}>
        <thead>
          <tr>
            <th>Área</th>
            <th>Função na área</th>
            <th>Concedido em</th>
            <th>Por</th>
          </tr>
        </thead>
        <tbody>
          {items.map((m) => (
            <tr key={`${m.areaCode}:${m.role}`}>
              <td>
                <span className={styles.code}>{m.areaCode}</span>
              </td>
              <td>
                <span className={styles.role}>{m.role}</span>
              </td>
              <td>
                {m.grantedAt ? SP_DATE_FORMATTER.format(new Date(m.grantedAt)) : "—"}
              </td>
              <td>{m.grantedBy ?? "—"}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
