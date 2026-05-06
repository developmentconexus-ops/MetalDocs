import styles from './CodePreviewBanner.module.css';

export type CodePreviewBannerProps = {
  profileCode: string | null;
  areaCode: string;
};

export function CodePreviewBanner({ profileCode, areaCode }: CodePreviewBannerProps): JSX.Element {
  const profile = profileCode ?? '???';
  const area = areaCode === '' ? '???' : areaCode;
  const code = `${profile}-${area}-???`;
  const ready = profileCode !== null && areaCode !== '';
  const kicker = ready
    ? `Código gerado · próximo em (${profile}, ${area})`
    : 'Código gerado · selecione perfil e área';
  const caption = `≈ ${code} · Código final atribuído ao confirmar.`;

  return (
    <div className={styles.banner} title="Código final atribuído ao confirmar">
      <div className="kicker">{kicker}</div>
      <div className={`${styles.code} mono`}>{code}</div>
      <div className="caption">{caption}</div>
    </div>
  );
}

export default CodePreviewBanner;
