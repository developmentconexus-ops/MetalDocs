import styles from './PageSizeSelector.module.css';

type PageSizeSelectorProps = {
  pageSize: number;
  options: number[];
  onPageSizeChange: (size: number) => void;
};

export function PageSizeSelector({ pageSize, options, onPageSizeChange }: PageSizeSelectorProps) {
  return (
    <label className={styles.root}>
      <span className={styles.label}>Itens por página</span>
      <select
        className={styles.select}
        value={pageSize}
        onChange={(event) => onPageSizeChange(Number(event.target.value))}
      >
        {options.map((option) => (
          <option key={option} value={option}>
            {option}
          </option>
        ))}
      </select>
    </label>
  );
}
