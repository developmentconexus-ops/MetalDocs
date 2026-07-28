import { usePreviewCodeQuery } from '../../../controlled-documents/queries/usePreviewCodeQuery';
import { CODE_PREVIEW_ERROR_MESSAGE, formatCodePreview } from '../../lib/codePreview';
import { resolveQueryError } from '../../../../lib/api';
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
  const { data, isLoading, isError, error, refetch } = usePreviewCodeQuery(
    profileCode,
    effectiveAreaCode,
  );

  const preview = formatCodePreview({
    ready,
    isLoading,
    isError,
    code: data?.code,
    profileCode,
    areaCode: effectiveAreaCode,
  });

  // F-QA4-1: a failed preview is an explicit error, never the "???" placeholder
  // that also means "nothing selected yet". Loading stays visually distinct
  // (PRC-TI-…) so the three states never collapse into one.
  if (preview.state === 'error') {
    return (
      <div className={styles.banner} role="alert" aria-live="assertive" aria-atomic="true">
        <div className="kicker">Código gerado · falha</div>
        <div className={styles.errorText}>{CODE_PREVIEW_ERROR_MESSAGE}</div>
        <div className="caption">
          {resolveQueryError(error, 'Não foi possível consultar a próxima sequência.')}{' '}
          <button type="button" className="btn btn-sm" onClick={() => void refetch()}>
            Tentar novamente
          </button>
        </div>
      </div>
    );
  }

  const kicker = ready
    ? `Código gerado · próximo em (${profileCode}, ${areaCode})`
    : 'Código gerado · selecione perfil e área';

  return (
    <div className={styles.banner} title="Código final atribuído ao confirmar">
      <div className="kicker">{kicker}</div>
      <div className={`${styles.code} mono`} aria-busy={preview.state === 'loading' || undefined}>
        {preview.text}
      </div>
    </div>
  );
}

export default CodePreviewBanner;
