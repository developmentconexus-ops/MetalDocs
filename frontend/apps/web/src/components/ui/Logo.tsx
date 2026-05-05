type LogoProps = {
  size?: 'sm' | 'md';
};

export function Logo({ size = 'md' }: LogoProps) {
  const markSize = size === 'sm' ? 18 : 22;
  const fontSize = size === 'sm' ? 13 : 15;
  return (
    <span
      style={{
        fontSize,
        fontFamily: 'var(--font-sans)',
        fontWeight: 600,
        letterSpacing: '-0.02em',
        display: 'inline-flex',
        alignItems: 'center',
        gap: 8,
      }}
    >
      <span
        style={{
          width: markSize,
          height: markSize,
          borderRadius: 5,
          background: 'var(--brand)',
          position: 'relative',
          display: 'inline-flex',
          alignItems: 'center',
          justifyContent: 'center',
          flexShrink: 0,
        }}
      >
        <span style={{ position: 'absolute', left: 4, right: 4, top: 5, height: 1.5, background: 'white' }} />
        <span style={{ position: 'absolute', left: 4, right: 4, top: 9, height: 1.5, background: 'white' }} />
        <span style={{ position: 'absolute', left: 4, right: 4, top: 13, height: 1.5, background: 'white' }} />
      </span>
      MetalDocs
    </span>
  );
}
