package html

// HtmlTemplate holds the HTML structure for attestation quote pages.
// It includes a favicon, logo, and a copy-to-clipboard feature.
// It conditionally renders the verification link based on .ShowVerify.
const HtmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>

    <!-- Add favicon for browser tab -->
    <link rel="icon" href="/images/favicon.png" type="image/png">
    
    <style>
        /* Base styles for the entire page */
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
        /* Container centers content and sets max width */
        .container {
            max-width: 1000px;
            margin: 0 auto;
            width: 100%;
        }
        /* Header flex layout to align logo and title */
        header {
            display: flex;
            align-items: center;
            margin-bottom: 30px;
        }
        /* Logo size and spacing */
        .logo {
            width: 40px;
            height: auto;
            margin-right: 12px;
        }
        /* Main title styling */
        h1 {
            font-size: 32px;
            font-weight: 500;
            margin: 0;
            padding: 0;
            color: #ffffff;
        }
        /* Description text color and spacing */
        p.description {
            color: #a0a0a0;
            margin-top: 8px;
        }
        /* Quote box container styles */
        .quote-container {
            position: relative;
            background-color: #252525;
            border-radius: 8px;
            border: 1px solid #333;
            overflow: hidden;
            margin-bottom: 20px;
        }
        /* Preformatted text styling for the quote */
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
        /* Copy button container positioning */
        .button-container {
            position: absolute;
            top: 8px;
            right: 8px;
        }
        /* Copy button base styles */
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
        /* Hover state for copy button */
        .copy-button:hover {
            background-color: #3a3a3a;
        }
        /* Active state for copy button */
        .copy-button:active {
            background-color: #444;
        }
        /* Icon inside the copy button */
        .copy-icon {
            width: 16px;
            height: 16px;
        }
        /* Toast notification base styles */
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
        /* Show animation for toast */
        .toast.show {
            display: block;
            animation: fadeInOut 2s ease;
        }
        /* Verification link styling */
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
            <!-- Display the logo next to the title -->
            <img src="/images/favicon.png" alt="Logo" class="logo">
            <div>
                <h1>{{.Title}}</h1>
                <p class="description">{{.Description}}</p>
            </div>
        </header>
        <div class="quote-container">
            <!-- Pre block shows quote text -->
            <pre class="quote-textarea" id="quoteTextarea">{{.Quote}}</pre>
            <div class="button-container">
                <!-- Button to copy quote to clipboard -->
                <button class="copy-button" id="copyButton">
                    <svg class="copy-icon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
                        <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>
                    </svg>
                    Copy
                </button>
            </div>
        </div>
        {{if .ShowVerify}}
        <p class="verification-link">
            Click <a href="#" id="verifyLink">here</a> to verify the attestation quote
        </p>
        {{end}}
    </div>
    <div class="toast" id="toast">Copied to clipboard</div>
    <script>
        document.addEventListener('DOMContentLoaded', function() {
            // Grab UI elements by their IDs
            const copyButton = document.getElementById('copyButton');
            const quoteTextarea = document.getElementById('quoteTextarea');
            const toast = document.getElementById('toast');

            // Copy text on button click
            copyButton.addEventListener('click', function() {
                const textToCopy = quoteTextarea.textContent;
                navigator.clipboard.writeText(textToCopy)
                    .then(function() {
                        // Show confirmation toast
                        toast.classList.add('show');
                        setTimeout(function() {
                            toast.classList.remove('show');
                        }, 2000);
                    })
                    .catch(function(err) {
                        // Fallback if clipboard API fails
                        console.error('Could not copy text: ', err);
                        fallbackCopyTextToClipboard(textToCopy);
                    });
            });

            // Fallback method using execCommand
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
                    document.execCommand('copy');
                    toast.classList.add('show');
                    setTimeout(function() {
                        toast.classList.remove('show');
                    }, 2000);
                } catch (err) {
                    console.error('Fallback: Could not copy text: ', err);
                }
                document.body.removeChild(textArea);
            }
        });
    </script>
</body>
</html>`
