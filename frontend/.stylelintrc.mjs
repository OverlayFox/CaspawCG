export default {
  extends: ["stylelint-config-standard"],
  rules: {
    // The stylesheet is organized by component/section (with banner
    // comments), not by ascending selector specificity. Since this rule
    // only fires where specificities genuinely differ, reordering to
    // satisfy it would be a no-op for rendering (the more specific rule
    // always wins regardless of source order) but would fragment the
    // section-based layout for no functional benefit.
    "no-descending-specificity": null,
  },
};
