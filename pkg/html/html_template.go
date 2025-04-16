package html

// HTML template for rendering attestation quote pages with dynamic title, description, and quote content.
const HtmlTemplate = `<!DOCTYPE html>
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
            padding: 20px;
            display: flex;
            flex-direction: column;
            min-height: 100vh;
        }
        .container {
            max-width: 1000px;
            margin: 0 auto;
            width: 100%;
        }
        header {
            margin-bottom: 30px;
        }
        h1 {
            font-size: 32px;
            font-weight: 500;
            margin: 0;
            padding: 0;
            color: #ffffff;
        }
        p.description {
            color: #a0a0a0;
            margin-top: 8px;
        }
        .quote-container {
            position: relative;
            background-color: #252525;
            border-radius: 8px;
            border: 1px solid #333;
            overflow: hidden;
            margin-bottom: 20px;
        }
        .quote-textarea {
            width: 100%;
            min-height: 80px;
            background-color: #252525;
            color: #e0e0e0;
            border: none;
            padding: 16px;
            font-family: 'Consolas', 'Courier New', monospace;
            font-size: 14px;
            line-height: 1.5;
            box-sizing: border-box;
            outline: none;
            overflow-x: auto;
            white-space: pre-wrap;
            word-break: break-all;
        }
        .button-container {
            position: absolute;
            top: 8px;
            right: 8px;
        }
        .copy-button {
            background-color: #2c2c2c;
            color: #e0e0e0;
            border: 1px solid #444;
            border-radius: 4px;
            padding: 6px 12px;
            font-size: 14px;
            cursor: pointer;
            display: flex;
            align-items: center;
            gap: 6px;
            transition: all 0.2s ease;
        }
        .copy-button:hover {
            background-color: #3a3a3a;
        }
        .copy-button:active {
            background-color: #444;
        }
        .copy-icon {
            width: 16px;
            height: 16px;
        }
        .toast {
            position: fixed;
            bottom: 20px;
            right: 20px;
            background-color: #333;
            color: white;
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
        .verification-link {
            text-align: center;
            margin-top: 12px;
            margin-bottom: 24px;
            font-size: 14px;
        }
        .verification-link a {
            color: #70a9ff;
            text-decoration: none;
            transition: color 0.2s ease;
        }
        .verification-link a:hover {
            color: #9cc2ff;
            text-decoration: underline;
        }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>{{.Title}}</h1>
            <p class="description">{{.Description}}</p>
        </header>
        <div class="quote-container">
            <pre class="quote-textarea" id="quoteTextarea">{{.Quote}}</pre>
            <div class="button-container">
                <button class="copy-button" id="copyButton">
                    <svg class="copy-icon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
                        <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>
                    </svg>
                    Copy
                </button>
            </div>
        </div>
        <p class="verification-link">Click <a href="#" id="verifyLink">here</a> to verify the attestation quote</p>
    </div>
    <div class="toast" id="toast">Copied to clipboard</div>
    <script>
        document.addEventListener('DOMContentLoaded', function() {
            const copyButton = document.getElementById('copyButton');
            const quoteTextarea = document.getElementById('quoteTextarea');
            const toast = document.getElementById('toast');

            // Copy the text to clipboard on button click
            copyButton.addEventListener('click', function() {
                const textToCopy = quoteTextarea.textContent;
                navigator.clipboard.writeText(textToCopy)
                    .then(function() {
                        toast.classList.add('show');
                        setTimeout(function() {
                            toast.classList.remove('show');
                        }, 2000);
                    })
                    .catch(function(err) {
                        console.error('Could not copy text: ', err);
                        fallbackCopyTextToClipboard(textToCopy);
                    });
            });

            // Fallback method for copying text for older browsers
            function fallbackCopyTextToClipboard(text) {
                const textArea = document.createElement("textarea");
                textArea.value = text;
                textArea.style.position = "fixed";
                textArea.style.left = "-999999px";
                textArea.style.top = "-999999px";
                document.body.appendChild(textArea);
                textArea.focus();
                textArea.select();
                try {
                    const successful = document.execCommand('copy');
                    if (successful) {
                        toast.classList.add('show');
                        setTimeout(function() {
                            toast.classList.remove('show');
                        }, 2000);
                    }
                } catch (err) {
                    console.error('Fallback: Could not copy text: ', err);
                }
                document.body.removeChild(textArea);
            }
        });
    </script>
</body>
</html>`