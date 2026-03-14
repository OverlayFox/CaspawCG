// 1. Initialize the Grid
const grid = GridStack.init({
  cellHeight: 100, // Height of one grid block
  margin: 10, // Space between blocks
  float: true, // Elements don't automatically fall to the bottom
});

let isLiveMode = false;

// 2. Mocking the data that will eventually come from your Go API
async function fetchApiOptions() {
  // In the future, this will be: return await window.go.main.UIService.GetDropdownOptions();
  return ["Select an option...", "User Data", "System Logs", "Network Stats"];
}

// 2.b Mocking Datasource options for the custom fields
async function fetchDatasources() {
  return await window.go.ui.UIService.GetDataSources();
}

// 3. Updated Add Widget Function
async function addWidget() {
  const options = await fetchApiOptions();
  let optionsHtml = options
    .map((opt) => `<option value="${opt}">${opt}</option>`)
    .join("");

  const widgetElement = `
        <div class="grid-stack-item">
            <div class="grid-stack-item-content widget-card">
                <strong>Dynamic Element</strong>
                
                <select class="api-dropdown edit-only">
                    ${optionsHtml}
                </select>
                
                <button class="action-btn" onclick="alert('Button Clicked!')">Execute</button>
                
                <button class="delete-btn edit-only" onclick="grid.removeWidget(this.closest('.grid-stack-item'))" style="background-color: #e74c3c; margin-left: 5px;">Remove</button>

                <div class="custom-fields-container" style="overflow-y: auto; max-height: 70%; margin-top: 10px;"></div>

                <button class="edit-only" onclick="addField(this)" style="margin-top: 10px; width: 100%;">➕ Add Custom Field</button>
            </div>
        </div>
    `;
  grid.addWidget(widgetElement, { w: 3, h: 3 });
}

// 4. New Function: Add Custom Field
async function addField(buttonElement) {
  const container = buttonElement.previousElementSibling; // targets custom-fields-container
  const sources = await fetchDatasources();
  const sourcesHtml = sources
    .map((s) => `<option value="${s}">${s}</option>`)
    .join("");

  const row = document.createElement("div");
  row.className = "field-row";

  row.innerHTML = `
        <div class="edit-only" style="display: flex; gap: 5px; width: 100%;">
            <input type="text" placeholder="Key" class="f-key" style="width: 25%;">
            <select class="f-type" style="width: 20%;">
                <option value="string">String</option>
                <option value="int">Int</option>
                <option value="float">Float</option>
            </select>
            <input type="text" placeholder="Location" class="f-id" style="width: 25%;">
            <select class="f-source" style="width: 25%;">
                ${sourcesHtml}
            </select>
            <button onclick="this.closest('.field-row').remove()" style="color: #e74c3c; cursor: pointer;">❌</button>
        </div>

        <div class="live-only" style="width: 100%;">
            <strong class="live-key-display"></strong>: <span class="live-value-display">Loading...</span>
        </div>
    `;

  container.appendChild(row);
}

// 5. Updated Toggle Edit/Live Mode Logic
async function toggleMode() {
  isLiveMode = !isLiveMode;
  const modeBtn = document.getElementById("toggle-mode-btn");
  const addBtn = document.getElementById("add-widget-btn");

  if (isLiveMode) {
    grid.enableMove(false);
    grid.enableResize(false);
    document.body.classList.add("is-live");
    modeBtn.innerText = "Current: LIVE MODE";
    modeBtn.classList.add("mode-live");
    addBtn.style.display = "none";

    // --- POPULATE LIVE DATA ---
    const fieldRows = document.querySelectorAll(".field-row");

    fieldRows.forEach(async (row) => {
      // Extract values from the inputs
      const key = row.querySelector(".f-key").value || "Unnamed Key";
      const type = row.querySelector(".f-type").value;
      const identifier = row.querySelector(".f-id").value;
      const source = row.querySelector(".f-source").value;

      // Set the Key label for Live mode
      row.querySelector(".live-key-display").innerText = key;
      const valueDisplay = row.querySelector(".live-value-display");
      valueDisplay.innerText = "Fetching...";

      // Simulate a Go backend delay
      setTimeout(() => {
        // Future Go implementation:
        // const liveData = await window.go.main.UIService.FetchData(identifier, type, source);

        // Dummy logic based on type
        let dummyData;
        if (type === "int") dummyData = Math.floor(Math.random() * 1000);
        else if (type === "float") dummyData = (Math.random() * 100).toFixed(2);
        else dummyData = `Data from ${source}`;

        valueDisplay.innerText = dummyData;
      }, 500);
    });
  } else {
    // Revert to Edit Mode
    grid.enableMove(true);
    grid.enableResize(true);
    document.body.classList.remove("is-live");
    modeBtn.innerText = "Current: EDIT MODE";
    modeBtn.classList.remove("mode-live");
    addBtn.style.display = "block";
  }
}

// Event Listeners
document.getElementById("add-widget-btn").addEventListener("click", addWidget);
document
  .getElementById("toggle-mode-btn")
  .addEventListener("click", toggleMode);
