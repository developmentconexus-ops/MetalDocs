import { registerValidateTemplate } from './validate-template.js';
import { registerRenderRoutes } from './render.js';
import { registerConvertPDF } from './convert-pdf.js';
import { registerFanoutRoute } from './fanout.js';
export function registerRoutes(app, env, s3Factory) {
    registerValidateTemplate(app, env, s3Factory);
    registerRenderRoutes(app, env, s3Factory);
    registerConvertPDF(app, env, s3Factory);
    registerFanoutRoute(app, env, s3Factory);
}
