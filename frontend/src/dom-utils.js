/**
 * DOMUtils — lightweight wrappers around common DOM operations.
 * Keeps repetitive boilerplate out of the manager files.
 */
export const DOMUtils = {
  createElement(tag, className = "", innerHTML = "") {
    const element = document.createElement(tag);
    if (className) element.className = className;
    if (innerHTML) element.innerHTML = innerHTML;
    return element;
  },

  createOptionsHTML(options) {
    return options
      .map((opt) => `<option value="${opt}">${opt}</option>`)
      .join("");
  },

  querySelector(selector, parent = document) {
    return parent.querySelector(selector);
  },

  querySelectorAll(selector, parent = document) {
    return parent.querySelectorAll(selector);
  },
};
