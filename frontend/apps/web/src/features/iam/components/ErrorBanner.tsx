import styles from "./ErrorBanner.module.css";

interface ErrorBannerProps {
  message: string;
  onRetry: () => void;
  retryLabel?: string;
}

export default function ErrorBanner({
  message,
  onRetry,
  retryLabel = "Tentar novamente",
}: ErrorBannerProps) {
  return (
    <div role="alert" className={styles.banner}>
      <span>{message}</span>
      <button type="button" className={styles.retryButton} onClick={onRetry}>
        {retryLabel}
      </button>
    </div>
  );
}
