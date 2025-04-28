// File: pkg/html/docker_live_logs.go
package html

// DockerLiveLogsTemplate defines a Docker logs page with
// radio controls to choose either name or index, an Apply button,
// auto-refresh after Apply only when scrolled to bottom,
// and auto-scroll to bottom on Apply.
// Styles reuse the existing attestation page theme.
const DockerLiveLogsTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>{{.Title}}</title>
  <style>
    body {
      font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
      background-color: #1e1e1e;
      color: #e0e0e0;
      margin: 0;
      padding: 0;
    }
    .container {
      max-width: 1000px;
      margin: 0 auto;
      padding: 20px;
    }
    header {
      margin-bottom: 20px;
    }
    h1 {
      font-size: 28px;
      font-weight: 500;
      margin: 0 0 12px 0;
      color: #ffffff;
    }
    #controls {
      background-color: #252525;
      padding: 12px;
      border-radius: 6px;
      border: 1px solid #333;
      display: flex;
      flex-wrap: wrap;
      align-items: center;
      gap: 12px;
    }
    #controls label {
      font-size: 14px;
      color: #cccccc;
    }
    #controls input[type="text"],
    #controls input[type="number"],
    #controls select {
      background-color: #1e1e1e;
      color: #e0e0e0;
      border: 1px solid #444;
      border-radius: 4px;
      padding: 4px 8px;
      font-size: 14px;
    }
    #controls button {
      background-color: #2c2c2c;
      color: #e0e0e0;
      border: 1px solid #444;
      border-radius: 4px;
      padding: 6px 12px;
      font-size: 14px;
      cursor: pointer;
      transition: background-color 0.2s ease;
    }
    #controls button:hover {
      background-color: #3a3a3a;
    }
    .logs-container {
      background-color: #252525;
      border-radius: 6px;
      border: 1px solid #333;
      padding: 16px;
      margin-top: 16px;
      height: 500px;
      overflow-y: auto;
      white-space: pre-wrap;
      word-break: break-all;
      font-family: 'Consolas', 'Courier New', monospace;
      font-size: 13px;
      line-height: 1.4;
    }
    .button-container {
      margin-top: 12px;
      text-align: right;
    }
    .copy-button {
      background-color: #2c2c2c;
      color: #e0e0e0;
      border: 1px solid #444;
      border-radius: 4px;
      padding: 8px 16px;
      font-size: 14px;
      cursor: pointer;
      transition: background-color 0.2s ease;
    }
    .copy-button:hover {
      background-color: #3a3a3a;
    }
    .toast {
      position: fixed;
      bottom: 20px;
      right: 20px;
      background-color: #333;
      color: #fff;
      padding: 12px 20px;
      border-radius: 4px;
      box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
      display: none;
      z-index: 1000;
    }
    .toast.show {
      display: block;
      animation: fadeInOut 2s ease;
    }
    @keyframes fadeInOut {
      0%,100% { opacity: 0 }
      10%,90% { opacity: 1 }
    }
  </style>
</head>
<body>
  <div class="container">
    <header>
      <h1>{{.Title}}</h1>
      <div id="controls">
        <!-- Choose mode -->
        <label><input type="radio" name="mode" value="name" checked> By Name</label>
        <input type="text" id="nameInput" placeholder="e.g test">
        <label><input type="radio" name="mode" value="index"> By Index</label>
        <input type="number" id="indexInput" min="0" placeholder="0" disabled>
        <!-- Lines selector -->
        <label for="linesSelect">Lines</label>
        <select id="linesSelect">
          <option value="100">100</option>
          <option value="500">500</option>
          <option value="1000" selected>1000</option>
        </select>
        <!-- Apply button -->
        <button id="applyButton">Apply</button>
      </div>
    </header>
    <!-- Log output area -->
    <div class="logs-container" id="logs">
      Choose mode, enter value, then click Apply.
    </div>
    <!-- Copy logs button -->
    <div class="button-container">
      <button class="copy-button" id="copyButton">Copy Logs</button>
    </div>
  </div>
  <div class="toast" id="toast">Logs copied to clipboard</div>
  <script>
    let refreshInterval = null;

    // Toggle inputs and reset when mode changes
    document.querySelectorAll('input[name="mode"]').forEach(radio => {
      radio.addEventListener('change', () => {
        const byName = document.querySelector('input[name="mode"]:checked').value === 'name';
        document.getElementById('nameInput').disabled = !byName;
        document.getElementById('indexInput').disabled = byName;
        clearInterval(refreshInterval);
        refreshInterval = null;
        document.getElementById('logs').textContent = 'Choose mode, enter value, then click Apply.';
      });
    });

    // Stop auto-refresh if user edits inputs
    ['nameInput', 'indexInput', 'linesSelect'].forEach(id => {
      document.getElementById(id).addEventListener('input', () => {
        clearInterval(refreshInterval);
        refreshInterval = null;
      });
    });

    // Perform fetch + scroll + auto-refresh
    async function applyFetch() {
      await fetchLogs();
      const logDiv = document.getElementById('logs');
      logDiv.scrollTop = logDiv.scrollHeight;        // auto-scroll on Apply
      if (refreshInterval) clearInterval(refreshInterval);
      refreshInterval = setInterval(fetchLogs, 2000);
    }

    // Core fetch logic: only updates if user is at bottom
    async function fetchLogs() {
      const logDiv = document.getElementById('logs');
      const atBottom = logDiv.scrollHeight - logDiv.scrollTop <= logDiv.clientHeight;
      if (!atBottom) return;

      const mode = document.querySelector('input[name="mode"]:checked').value;
      const lines = document.getElementById('linesSelect').value;
      let url = '/docker_logs?lines=' + lines;
      if (mode === 'name') {
        const name = document.getElementById('nameInput').value.trim();
        if (!name) { logDiv.textContent = 'Please enter a container name.'; return; }
        url += '&name=' + encodeURIComponent(name);
      } else {
        const idx = document.getElementById('indexInput').value.trim();
        if (idx === '') { logDiv.textContent = 'Please enter a container index.'; return; }
        url += '&index=' + encodeURIComponent(idx);
      }

      try {
        const res = await fetch(url);
        const text = await res.text();
        if (!res.ok) {
          try {
            const errObj = JSON.parse(text);
            logDiv.textContent = errObj.error + ': ' + errObj.details;
          } catch {
            logDiv.textContent = text;
          }
          return;
        }
        logDiv.textContent = text;
      } catch (e) {
        console.error(e);
        logDiv.textContent = 'Failed to fetch logs: ' + e.message;
      }
    }

    // Bind Apply and Copy buttons
    document.getElementById('applyButton').addEventListener('click', applyFetch);
    document.getElementById('copyButton').addEventListener('click', () => {
      const text = document.getElementById('logs').textContent;
      navigator.clipboard.writeText(text).then(() => {
        const toast = document.getElementById('toast');
        toast.classList.add('show');
        setTimeout(() => toast.classList.remove('show'), 2000);
      });
    });
  </script>
</body>
</html>`
