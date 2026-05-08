import styles from "./MiniDocPreview.module.css";

const LINE_COUNT = 8;

export function MiniDocPreview() {
  return (
    <div className={styles.miniDoc}>
      <div className={styles.miniDocBrand} />
      {Array.from({ length: LINE_COUNT }).map((_, k) => (
        <div key={k} className={styles.miniDocLine} />
      ))}
    </div>
  );
}
