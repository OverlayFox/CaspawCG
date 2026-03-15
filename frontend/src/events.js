export function initLiveEvents() {
  console.log("Live events listener is mounting...");

  window.runtime.EventsOn("live-data-update", (data) => {
    console.info("Live data update received:", data);

    if (data.identifier === "CasparCGKeepAlive") {
      updateCasparStatus(data.value);
    }

    if (!document.body.classList.contains("is-live")) return;

    const fieldRows = document.querySelectorAll(".field-row");
    fieldRows.forEach((row) => {
      const identifier = row.querySelector(".f-id").value;

      if (identifier === data.identifier) {
        const valueDisplay = row.querySelector(".live-value-display");

        if (typeof data.value === "object") {
          valueDisplay.innerText = JSON.stringify(data.value);
        } else {
          valueDisplay.innerText = data.value;
        }

        valueDisplay.style.color = "#2ecc71";
        setTimeout(() => (valueDisplay.style.color = ""), 300);
      }
    });
  });
}

function updateCasparStatus(clientData) {
  const { host, port, isAlive } = clientData;
  const container = document.getElementById("caspar-clients-container");
  if (!container) return;

  // Create a safe, unique HTML ID from the host and port (e.g., "caspar-192-168-1-5-5250")
  const safeId = `caspar-${host}-${port}`.replace(/[^a-zA-Z0-9-]/g, "-");

  // Check if we already have a chip for this client
  let chip = document.getElementById(safeId);

  if (!chip) {
    // Build a new chip if it's the first time we are seeing this client
    chip = document.createElement("div");
    chip.id = safeId;
    chip.className = "client-chip";

    const dot = document.createElement("div");
    dot.className = `status-dot ${isAlive ? "status-online" : "status-offline"}`;

    const text = document.createElement("span");
    text.innerText = `${host}:${port}`;

    chip.appendChild(dot);
    chip.appendChild(text);
    container.appendChild(chip);
  } else {
    // If the chip exists, just update the dot color
    const dot = chip.querySelector(".status-dot");
    if (isAlive) {
      dot.classList.remove("status-offline");
      dot.classList.add("status-online");
    } else {
      dot.classList.remove("status-online");
      dot.classList.add("status-offline");
    }
  }
}
