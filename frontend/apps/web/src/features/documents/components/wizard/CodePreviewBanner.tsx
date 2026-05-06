import styles from './CodePreviewBanner.module.css';

export type CodePreviewBannerProps = {
  kicker: string;
  code: string;
  caption: string;
};

export function CodePreviewBanner({ kicker, code, caption }: CodePreviewBannerProps): JSX.Element {
  return (
    <div className={styles.banner}>
      <div className="kicker">{kicker}</div>
      <div className={`${styles.code} mono`}>{code}</div>
      <div className="caption">{caption}</div>
    </div>
  );
}

export default CodePreviewBanner;
