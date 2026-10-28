import path from 'node:path'
import { fileURLToPath } from 'node:url'

import tailwindcss from '@tailwindcss/vite'
import { tanstackRouter } from '@tanstack/router-plugin/vite'
import react from '@vitejs/plugin-react'
import { loadEnv,defineConfig,lazyPlugins } from 'vite-plus'

const __dirname = path.dirname(fileURLToPath(import.meta.url))

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const serverUrl =
    process.env.VITE_REACT_APP_SERVER_URL ||
    env.VITE_REACT_APP_SERVER_URL ||
    'http://localhost:3000'
  const isProd = mode === 'production'
  const devProxy = Object.fromEntries(
    (['/api', '/v1', '/mj', '/pg'] as const).map((key) => [
      key,
      { target: serverUrl, changeOrigin: true },
    ])
  ) as Record<string, { target: string; changeOrigin: boolean }>

  return {
    plugins: lazyPlugins(() => [
      react(),
      tailwindcss(),
      tanstackRouter({
        target: 'react',
        // Dev: avoid per-route async chunks (reduces white flash on navigation + faster HMR feedback).
        // Prod: keep route-based code splitting.
        autoCodeSplitting: isProd,
      }),
    ]),
    resolve: {
      alias: {
        '@': path.resolve(__dirname, './src'),
      },
    },
    server: {
      host: '0.0.0.0',
      strictPort: false,
      proxy: devProxy,
    },
    build: {
      outDir: 'dist',
      target: 'es2020',
      // Emit a consolidated license report for bundled dependencies.
      license: true,
    },
    test: {
      environment: 'jsdom',
      server: {
        deps: { inline: [/@lobehub\//, /antd-style/] },
      },
      setupFiles: ['./src/test-setup.ts'],
      // Several heavy jsdom suites (channel-configuration, visual-billing-editor)
      // legitimately take >5s per test on contended CI runners; the vitest
      // default of 5000ms fails whichever of them crosses the line first. The
      // heaviest test measures ~3.2s uncontended, so 20s keeps headroom for the
      // ~4x slowdown observed on shared runners.
      testTimeout: 20000,
      clearMocks: true,
      restoreMocks: true,
      include: [
        'src/**/*.{test,spec}.{ts,tsx}',
        'scripts/oxlint/__tests__/*.test.ts',
      ],
    },
    fmt: {
      arrowParens: 'always',
      bracketSpacing: true,
      endOfLine: 'lf',
      insertFinalNewline: true,
      jsxSingleQuote: true,
      printWidth: 80,
      quoteProps: 'as-needed',
      semi: false,
      singleQuote: true,
      sortImports: true,
      sortPackageJson: true,
      sortTailwindcss: true,
      tabWidth: 2,
      trailingComma: 'es5',
      ignorePatterns: [
        'node_modules',
        'dist',
        'build',
        'coverage',
        '.tanstack',
        '.eslintrc.cjs',
        '**/routeTree.gen.ts',
        'src/routeTree.gen.ts',
      ],
      overrides: [
        {
          files: ['../plugins/**/*.{js,mjs,cjs}'],
          semi: true,
          singleQuote: false,
          printWidth: 160,
        },
      ],
    },
    lint: {
      plugins: ['typescript', 'unicorn', 'oxc', 'react', 'import', 'promise'],
      jsPlugins: [
        {
          name: 'project',
          specifier: './scripts/oxlint/intl-locale.mjs',
        },
        {
          name: 'vite-plus',
          specifier: 'vite-plus/oxlint-plugin',
        },
      ],
      categories: {
        correctness: 'error',
      },
      env: {
        browser: true,
        builtin: true,
        es2024: true,
        node: true,
      },
      ignorePatterns: [
        'node_modules',
        'dist',
        'build',
        'coverage',
        '.tanstack',
        '.eslintrc.cjs',
        'src/components/ui',
        'src/routeTree.gen.ts',
      ],
      rules: {
        curly: ['error', 'multi-line'],
        eqeqeq: [
          'error',
          'always',
          {
            null: 'ignore',
          },
        ],
        'import/first': 'warn',
        'import/newline-after-import': 'warn',
        'import/no-cycle': 'error',
        'import/no-duplicates': [
          'error',
          {
            preferInline: true,
          },
        ],
        'import/no-unassigned-import': 'off',
        'no-array-constructor': 'error',
        'no-console': 'warn',
        'no-duplicate-imports': 'error',
        'no-nested-ternary': 'error',
        'no-new-wrappers': 'error',
        'no-regex-spaces': 'error',
        'no-unsafe-finally': 'warn',
        'no-unsafe-negation': 'error',
        'no-unused-expressions': 'warn',
        'no-unused-labels': 'error',
        'no-unused-vars': [
          'warn',
          {
            args: 'all',
            argsIgnorePattern: '^_',
            caughtErrors: 'all',
            caughtErrorsIgnorePattern: '^_',
            destructuredArrayIgnorePattern: '^_',
            varsIgnorePattern: '^_',
            ignoreRestSiblings: true,
          },
        ],
        'no-useless-catch': 'warn',
        'no-useless-computed-key': 'error',
        'no-useless-escape': 'warn',
        'no-useless-rename': 'warn',
        'no-var': 'warn',
        'object-shorthand': 'warn',
        'oxc/const-comparisons': 'error',
        'prefer-const': 'warn',
        'prefer-object-has-own': 'warn',
        'prefer-object-spread': 'warn',
        'prefer-template': 'warn',
        'project/intl-locale': 'error',
        'promise/catch-or-return': 'error',
        'promise/no-callback-in-promise': 'error',
        'promise/no-new-statics': 'error',
        'promise/no-return-wrap': 'warn',
        'promise/valid-params': 'warn',
        // This project does not use the React Compiler; these rules flag
      // compiler-specific constraints that do not apply to the current code.
      'react/immutability': 'off',
      'react/incompatible-library': 'off',
      'react/preserve-manual-memoization': 'off',
      'react/purity': 'off',
      'react/refs': 'off',
      'react/set-state-in-effect': 'off',
      'react/button-has-type': 'error',
        'react/checked-requires-onchange-or-readonly': 'warn',
        'react/exhaustive-deps': 'error',
        'react/iframe-missing-sandbox': 'error',
        'react/jsx-boolean-value': 'warn',
        'react/jsx-curly-brace-presence': 'warn',
        'react/jsx-fragments': 'warn',
        'react/jsx-key': 'warn',
        'react/jsx-no-duplicate-props': 'error',
        'react/jsx-no-target-blank': 'warn',
        'react/jsx-no-undef': 'error',
        'react/jsx-no-useless-fragment': 'error',
        'react/jsx-props-no-spread-multi': 'warn',
        'react/no-array-index-key': 'error',
        'react/no-danger': 'warn',
        'react/no-unknown-property': 'warn',
        'react/no-unstable-nested-components': [
          'error',
          {
            allowAsProps: true,
          },
        ],
        'react/only-export-components': [
          'error',
          {
            allowConstantExport: true,
            allowExportNames: [
              'getAuthSectionNavItems',
              'getAuthSectionContent',
              'getAuthSectionMeta',
              'getBillingSectionNavItems',
              'getBillingSectionContent',
              'getBillingSectionMeta',
              'getContentSectionNavItems',
              'getContentSectionContent',
              'getContentSectionMeta',
              'getDashboardSectionNavItems',
              'getModelsSectionNavItems',
              'getModelsSectionContent',
              'getModelsSectionMeta',
              'getOperationsSectionNavItems',
              'getOperationsSectionContent',
              'getOperationsSectionMeta',
              'getPolicySectionNavItems',
              'getPolicySectionContent',
              'getPolicySectionMeta',
              'getSecuritySectionNavItems',
              'getSecuritySectionContent',
              'getSecuritySectionMeta',
              'getSiteSectionNavItems',
              'getSiteSectionContent',
              'getSiteSectionMeta',
              'getUsageLogsSectionNavItems',
              'isUsageLogsSectionId',
              'useSuppressSettingsSectionHeader',
            ],
          },
        ],
        'react/react-in-jsx-scope': 'off',
        'react/rules-of-hooks': 'warn',
        'react/self-closing-comp': 'warn',
        'typescript/consistent-type-imports': [
          'error',
          {
            prefer: 'type-imports',
            fixStyle: 'inline-type-imports',
            disallowTypeAnnotations: false,
          },
        ],
        'typescript/no-duplicate-type-constituents': 'error',
        'typescript/no-explicit-any': 'warn',
        'typescript/no-extra-non-null-assertion': 'error',
        'typescript/no-import-type-side-effects': 'error',
        'typescript/no-non-null-assertion': 'error',
        'typescript/no-require-imports': 'error',
        'typescript/no-unnecessary-template-expression': 'error',
        'typescript/no-unnecessary-type-arguments': 'error',
        'typescript/no-unnecessary-type-assertion': 'error',
        'typescript/no-useless-empty-export': 'error',
        'typescript/no-wrapper-object-types': 'error',
        'typescript/prefer-as-const': 'error',
        'typescript/prefer-namespace-keyword': 'error',
        'unicorn/no-single-promise-in-promise-methods': 'error',
        'unicorn/no-unnecessary-await': 'error',
        'unicorn/no-useless-fallback-in-spread': 'warn',
        'unicorn/no-useless-length-check': 'warn',
        'unicorn/no-useless-promise-resolve-reject': 'warn',
        'unicorn/no-useless-spread': 'error',
        'unicorn/prefer-add-event-listener': 'warn',
        'unicorn/prefer-array-find': 'warn',
        'unicorn/prefer-array-flat': 'warn',
        'unicorn/prefer-array-flat-map': 'warn',
        'unicorn/prefer-array-index-of': 'warn',
        'unicorn/prefer-array-some': 'warn',
        'unicorn/prefer-at': 'error',
        'unicorn/prefer-date-now': 'warn',
        'unicorn/prefer-includes': 'warn',
        'unicorn/prefer-keyboard-event-key': 'warn',
        'unicorn/prefer-modern-dom-apis': 'warn',
        'unicorn/prefer-module': 'warn',
        'unicorn/prefer-node-protocol': 'warn',
        'unicorn/prefer-number-properties': 'warn',
        'unicorn/prefer-object-from-entries': 'warn',
        'unicorn/prefer-optional-catch-binding': 'error',
        'unicorn/prefer-query-selector': 'warn',
        'unicorn/prefer-regexp-test': 'warn',
        'unicorn/prefer-set-has': 'warn',
        'unicorn/prefer-spread': 'error',
        'unicorn/prefer-string-replace-all': 'error',
        'unicorn/prefer-string-slice': 'warn',
        'unicorn/prefer-string-starts-ends-with': 'warn',
        'unicorn/prefer-string-trim-start-end': 'warn',
        'unicorn/prefer-structured-clone': 'warn',
        'unicorn/throw-new-error': 'error',
        'vite-plus/prefer-vite-plus-imports': 'error',
      },
      overrides: [
        {
          files: ['src/routes/**/*.{ts,tsx}'],
          rules: {
            'react/only-export-components': 'off',
          },
        },
        {
          files: ['**/scripts/**/*.{js,mjs,cjs}'],
          rules: {
            'no-console': 'off',
            'prefer-node-protocol': 'off',
          },
        },
        {
          files: ['../plugins/**/*.{js,mjs,cjs}'],
          globals: {
            console: 'readonly',
            utils: 'readonly',
          },
          rules: {
            'no-console': 'off',
            'no-underscore-dangle': [
              'warn',
              {
                allow: ['__fileRef'],
              },
            ],
            'no-undef': 'error',
            'no-unused-vars': [
              'error',
              {
                argsIgnorePattern: '^_',
                caughtErrorsIgnorePattern: '^e$',
              },
            ],
            'no-var': 'error',
            'prefer-node-protocol': 'off',
            'project/intl-locale': 'off',
          },
        },
      ],
      options: {
        // Type checking is handled by `tsc -b`; keep lint rules aligned with
        // the project's previous non-type-aware Oxlint policy.
        typeAware: false,
        typeCheck: false,
      },
    },
  }
})
