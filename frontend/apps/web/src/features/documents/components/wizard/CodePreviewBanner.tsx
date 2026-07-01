import { usePreviewCodeQuery } from '../../../controlled-documents/queries/usePreviewCodeQuery';
import { formatCodePreview } from '../../lib/codePreview';
import styles from './CodePreviewBanner.module.css';

export type CodePreviewBannerProps = {
  profileCode: string | null;
  areaCode: string;
};

export function CodePreviewBanner({
  profileCode,
  areaCode,
}: CodePreviewBannerProps): JSX.Element {
  const effectiveAreaCode = areaCode !== '' ? areaCode : null;
  const ready = profileCode !== null && effectiveAreaCode !== null;
  const { data, isLoading } = usePreviewCodeQuery(profileCode, effectiveAreaCode);

  const code = formatCodePreview({
    ready,
    isLoading,
    code: data?.code,
    profileCode,
    areaCode: effectiveAreaCode,
  });

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

