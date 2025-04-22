// File: pkg/html/docker_live_logs.go
package html

// DockerLiveLogsTemplate defines a live‑updating Docker logs page
// that preserves scroll position when user scrolls up.
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
            margin: 0 0 8px 0;
            color: #ffffff;
        }
        #controls {
            background-color: #252525;
            padding: 12px;
            border-radius: 6px;
            border: 1px solid #333;
            display: flex;
            align-items: center;
            gap: 8px;
        }
        #controls label {
            font-size: 14px;
            color: #cccccc;
        }
        #controls select {
            background-color: #1e1e1e;
            color: #e0e0e0;
            border: 1px solid #444;
            border-radius: 4px;
            padding: 4px 8px;
            font-size: 14px;
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
            0% { opacity: 0; }
            10% { opacity: 1; }
            90% { opacity: 1; }
            100% { opacity: 0; }
        }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>{{.Title}}</h1>
            <div id="controls">
                <label for="linesSelect">Show last</label>
                <select id="linesSelect">
                    <option value="100">100</option>
                    <option value="500">500</option>
                    <option value="1000" selected>1000</option>
                </select>
                <label for="linesSelect">lines</label>
            </div>
        </header>
        <div class="logs-container" id="logs">Loading logs...</div>
        <div class="button-container">
            <button class="copy-button" id="copyButton">Copy Logs</button>
        </div>
    </div>
    <div class="toast" id="toast">Logs copied to clipboard</div>
    <script>
        // Fetch logs and update container, preserving scroll if user scrolled up
        async function fetchLogs() {
            const logDiv = document.getElementById('logs');
            // determine if user is at bottom before update
            const atBottom = logDiv.scrollHeight - logDiv.scrollTop === logDiv.clientHeight;
            const count = document.getElementById('linesSelect').value;
            try {
                const res = await fetch('/docker_logs?lines=' + count);
                const text = await res.text();
                logDiv.textContent = text;
                // only auto-scroll if user was at bottom
                if (atBottom) {
                    logDiv.scrollTop = logDiv.scrollHeight;
                }
            } catch (e) {
                console.error(e);
            }
        }

        // copy logs to clipboard
        document.getElementById('copyButton').addEventListener('click', () => {
            const text = document.getElementById('logs').textContent;
            navigator.clipboard.writeText(text).then(() => {
                const toast = document.getElementById('toast');
                toast.classList.add('show');
                setTimeout(() => toast.classList.remove('show'), 2000);
            }).catch(err => console.error('Copy failed', err));
        });

        // re-fetch on selector change
        document.getElementById('linesSelect').addEventListener('change', fetchLogs);
        // initial fetch and interval refresh
        fetchLogs();
        setInterval(fetchLogs, 2000);
    </script>
</body>
</html>`