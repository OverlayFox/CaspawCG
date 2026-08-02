import js from "@eslint/js";
import prettierConfig from "eslint-config-prettier";
import litPlugin from "eslint-plugin-lit";
import globals from "globals";
import tseslint from "typescript-eslint";

export default tseslint.config(
  { ignores: ["dist/**", "wailsjs/**", "node_modules/**"] },
  {
    files: ["src/**/*.ts"],
    extends: [
      js.configs.recommended,
      tseslint.configs.strictTypeChecked,
      tseslint.configs.stylisticTypeChecked,
      litPlugin.configs["flat/recommended"],
    ],
    languageOptions: {
      ecmaVersion: 2022,
      sourceType: "module",
      globals: globals.browser,
      parserOptions: {
        projectService: true,
        tsconfigRootDir: import.meta.dirname,
      },
    },
    rules: {
      // Lit calls event-binding handlers (`@click=${this.method}`) with `this`
      // bound to the host element itself (lit-html's EventPart sets
      // eventContext to the host), so passing an unbound method reference is
      // the standard, safe Lit idiom rather than a real unbound-`this` bug.
      "@typescript-eslint/unbound-method": "off",

      // Lit lifecycle methods (firstUpdated, updated, connectedCallback, ...)
      // are routinely overridden as `async` even though LitElement's own
      // signature returns void; this is expected and safe.
      "@typescript-eslint/no-misused-promises": [
        "error",
        { checksVoidReturn: { attributes: false, inheritedMethods: false } },
      ],

      // One-line arrow event handlers in templates (`@click=${() =>
      // this.dispatchEvent(...)}`) are idiomatic Lit; requiring braces adds
      // noise without a safety benefit.
      "@typescript-eslint/no-confusing-void-expression": [
        "error",
        { ignoreArrowShorthand: true },
      ],

      // Numbers and booleans stringify safely and are used throughout for
      // formatting values into templates.
      "@typescript-eslint/restrict-template-expressions": [
        "error",
        { allowNumber: true, allowBoolean: true },
      ],

      // Conflicts directly with no-non-null-assertion below (it prefers `!`
      // where no-non-null-assertion forbids it); keep the rule that actually
      // flags a real correctness risk.
      "@typescript-eslint/non-nullable-type-assertion-style": "off",

      // Wails' generated Go bindings type slices/objects as always-present,
      // but Go's zero values (nil slices, "", 0) marshal to JSON as
      // null/falsy and land here at runtime regardless of the declared TS
      // type — so `x || default` / `x?.y` guards that look "unnecessary" to
      // the type checker are load-bearing defensive code against that
      // boundary, not dead code.
      "@typescript-eslint/no-unnecessary-condition": "off",
      "@typescript-eslint/prefer-nullish-coalescing": "off",
    },
  },
  prettierConfig,
);
