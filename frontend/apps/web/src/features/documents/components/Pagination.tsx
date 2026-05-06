import styles from './Pagination.module.css';

type PaginationProps = {
  page: number;
  totalPages: number;
  onPageChange: (page: number) => void;
};

export function Pagination({ page, totalPages, onPageChange }: PaginationProps) {
  const safeTotalPages = Math.max(1, totalPages);

  return (
    <div className={styles.root}>
      <button
        type="button"
        className={styles.button}
        onClick={() => onPageChange(page - 1)}
        disabled={page <= 1}
      >
        Anterior
      </button>
      <span className={styles.indicator}>
        Página {page} de {safeTotalPages}
      </span>
      <button
        type="button"
        className={styles.button}
        onClick={() => onPageChange(page + 1)}
        disabled={page >= safeTotalPages}
      >
        Próxima
      </button>
    </div>
  );
}
