import { build } from 'esbuild';
import { writeFileSync } from 'node:fs';

// Bundle our own source + @metaldocs/* workspace libraries into one ESM file.
//
// Every real node_modules dependency is left EXTERNAL so it loads natively at
// runtime in its own module format. This is the load-bearing decision: fastify
// is CommonJS, @eigenpal/docx-editor-react is dual-format and relies on
// import.meta.url. Bundling them into a single file forces incompatible module
// systems together (the "Dynamic require" / "import.meta.url undefined" errors).
// Externalizing them keeps our entry as clean ESM and lets Node resolve each
// dependency in its intended format.
//
// Only @metaldocs/* workspace packages are inlined, because shared-tokens ships
// raw TypeScript (no compiled dist) and must be compiled into the bundle.
const externalizeNodeModules = {
  name: 'externalize-node-modules',
  setup(b) {
    // Match bare specifiers (anything not starting with "." or "/").
    b.onResolve({ filter: /^[^./]/ }, (args) => {
      if (args.path.startsWith('@metaldocs/')) {
        return null; // fall through to default resolution → inline into bundle
      }
      return { path: args.path, external: true }; // node_modules dep + node:* builtins
    });
  },
};

const result = await build({
  entryPoints: ['src/index.ts'],
  bundle: true,
  platform: 'node',
  target: 'node20',
  format: 'esm',
  outfile: 'dist/index.js',
  sourcemap: true,
  metafile: true,
  logLevel: 'info',
  plugins: [externalizeNodeModules],
});
writeFileSync('dist/meta.json', JSON.stringify(result.metafile));
