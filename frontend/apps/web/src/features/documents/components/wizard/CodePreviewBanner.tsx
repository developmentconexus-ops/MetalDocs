import styles from './CodePreviewBanner.module.css';

export type CodePreviewBannerProps = {
  profileCode: string | null;
  areaCode: string;
};

const PLACEHOLDER = '???';

export function CodePreviewBanner({
  profileCode,
  areaCode,
}: CodePreviewBannerProps): JSX.Element {
  const ready = profileCode !== null && areaCode !== '';
  const code = `${profileCode ?? PLACEHOLDER}-${areaCode || PLACEHOLDER}-${PLACEHOLDER}`;
  const kicker = ready
    ? `Código gerado · próximo em (${profileCode}, ${areaCode})`
    : 'Código gerado · selecione perfil e área';

  return (
    <div className={styles.banner} title="Código final atribuído ao confirmar">
      <div className="kicker">{kicker}</div>
      <div className={`${styles.code} mono`}>{code}</div>
    </div>
  );
}

export default CodePreviewBanner;
