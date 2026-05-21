import { usePreviewCodeQuery } from '../../../controlled-documents/queries/usePreviewCodeQuery';
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
  const effectiveAreaCode = areaCode !== '' ? areaCode : null;
  const ready = profileCode !== null && effectiveAreaCode !== null;
  const { data, isLoading } = usePreviewCodeQuery(profileCode, effectiveAreaCode);

  const code = !ready
    ? `${PLACEHOLDER}-${PLACEHOLDER}-${PLACEHOLDER}`
    : isLoading
    ? `${profileCode}-${areaCode}-…`
    : data?.code ?? `${profileCode}-${areaCode}-${PLACEHOLDER}`;

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

