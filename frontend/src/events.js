export function initLiveEvents() {
  window.runtime.EventsOn("live-data-update", (data) => {
    const fieldRows = document.querySelectorAll(".field-row");

    fieldRows.forEach((row) => {
      const identifier = row.querySelector(".f-id").value;

      if (identifier === data.identifier) {
        const valueDisplay = row.querySelector(".CasparCGKeepAlive");

        if (data.identifier === "CasparCGKeepAlive") {
          valueDisplay.innerText = data.value.IsAlive
            ? "🟢 Online"
            : "🔴 Offline";
        } else {
          // Fallback for other event types
          valueDisplay.innerText = JSON.stringify(data.value);
        }

        valueDisplay.style.color = "#2ecc71";
        setTimeout(() => (valueDisplay.style.color = ""), 300);
      }
    });
  });
}
