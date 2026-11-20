/**
 * Feature-based の依存ルール：app → features → shared
 * @type {import('dependency-cruiser').IConfiguration}
 */
module.exports = {
  forbidden: [
    {
      name: "shared-must-not-depend-on-features-or-app",
      severity: "error",
      from: { path: "^src/shared/" },
      to: { path: "^src/(features|app)/" },
    },
    {
      name: "features-must-not-depend-on-app",
      severity: "error",
      from: { path: "^src/features/" },
      to: { path: "^src/app/" },
    },
    {
      name: "no-cross-feature-imports",
      comment: "feature 同士は import しない。共通にしたいものは shared に移す",
      severity: "error",
      from: { path: "^src/features/([^/]+)/" },
      to: { path: "^src/features/([^/]+)/", pathNot: "^src/features/$1/" },
    },
    {
      name: "features-only-via-public-api",
      comment: "feature は index.ts 経由でだけ import する",
      severity: "error",
      from: { path: "^src/app/" },
      to: { path: "^src/features/[^/]+/.+", pathNot: "^src/features/[^/]+/index\\.ts$" },
    },
    {
      name: "no-circular",
      severity: "error",
      from: {},
      to: { circular: true },
    },
    {
      name: "not-to-unresolvable",
      severity: "error",
      from: {},
      to: { couldNotResolve: true },
    },
    {
      name: "no-non-package-json",
      comment: "package.json に書いていないパッケージは import しない",
      severity: "error",
      from: {},
      to: { dependencyTypes: ["npm-no-pkg", "npm-unknown"] },
    },
  ],
  options: {
    doNotFollow: { path: "node_modules" },
    exclude: { path: "src/shared/api/generated" },
    tsConfig: { fileName: "tsconfig.json" },
    tsPreCompilationDeps: true,
    enhancedResolveOptions: {
      exportsFields: ["exports"],
      conditionNames: ["import", "require", "node", "default", "types"],
    },
  },
};
