const formEl = document.getElementById('analyzerForm');
const inputEl = document.getElementById('urlInputField');
const btnEl = document.getElementById('analyzeBtn');
const btnTextEl = document.querySelector('.analyze-button-text');
const btnLoadingEl = document.querySelector('.analyze-loading-icon');
const resultsSection = document.getElementById('resultsSection');
const resultsContainer = document.getElementById('resultsContainer');
const errorSection = document.getElementById('errorSection');
const errorMessageEl = document.getElementById('errorMessage');

formEl.addEventListener('submit', async (e) => {
    e.preventDefault();

    const url = inputEl.value.trim();

    const { valid, message } = validateURLWithMessage(url);
    if (!valid) {
        showError(message);
        return;
    }

    await analyzeURL(url);
});

function isValidURL(str) {
    try {
        const parsed = new URL(str);
        return parsed.protocol === 'http:' || parsed.protocol === 'https:';
    } catch (err) {
        return false;
    }
}

function validateURLWithMessage(url) {
    if (!url) {
        return { valid: false, message: 'Please enter a URL to analyze' };
    }

    if (!url.startsWith('http://') && !url.startsWith('https://')) {
        return { valid: false, message: 'URL must start with http:// or https://' };
    }

    if (!isValidURL(url)) {
        return { valid: false, message: 'Please enter a valid URL format' };
    }

    return { valid: true, message: '' };
}

async function analyzeURL(url) {
    setLoadingState(true);
    hideError();
    hideResults();

    try {
        const resp = await fetch('/api/analyze', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ url })
        });

        if (!resp.ok) {
            const text = await resp.text();
            throw new Error(`${resp.status}: ${text}`);
        }

        const result = await resp.json();
        displayResults(result);
    } catch (err) {
        showError(`Analysis failed: ${err.message}`);
    } finally {
        setLoadingState(false);
    }
}

function setLoadingState(loading) {
    btnEl.disabled = loading;
    btnTextEl.style.display = loading ? 'none' : 'inline';
    btnLoadingEl.style.display = loading ? 'inline' : 'none';
}

function showResults() {
    resultsSection.style.display = 'block';
    resultsSection.scrollIntoView({ behavior: 'smooth' });
}

function hideResults() {
    resultsSection.style.display = 'none';
}

function showError(msg) {
    errorMessageEl.textContent = msg;
    errorSection.style.display = 'block';
    errorSection.scrollIntoView({ behavior: 'smooth' });
}

function hideError() {
    errorSection.style.display = 'none';
}

function displayResults(data) {
    resultsContainer.innerHTML = '';

    resultsContainer.appendChild(createResultCard('HTML Version', data.html_version));
    resultsContainer.appendChild(createResultCard('Page Title', data.title || 'No title found'));

    resultsContainer.appendChild(createHeadingsCard(data.headings));

    resultsContainer.appendChild(createLinksCard(data.links));
    resultsContainer.appendChild(createLoginFormCard(data.has_login_form));

    showResults();
}

function createResultCard(label, value) {
    const el = document.createElement('div');
    el.className = 'result-card';
    el.innerHTML = `
        <h3>${label}</h3>
        <div class="result-value">${value}</div>
    `;
    return el;
}

function createHeadingsCard(headings) {
    const el = document.createElement('div');
    el.className = 'result-card';

    const total = Object.values(headings).reduce((sum, val) => sum + val, 0);

    let breakdownHTML = '';
    for (let i = 1; i <= 6; i++) {
        const level = `h${i}`;
        breakdownHTML += `
            <div class="heading-item">
                <div class="heading-level">${level.toUpperCase()}</div>
                <div class="heading-count">${headings[level] || 0}</div>
            </div>
        `;
    }

    el.innerHTML = `
        <h3>Headings Analysis</h3>
        <div class="result-value">Total: ${total} headings</div>
        <div class="headings-breakdown">
            ${breakdownHTML}
        </div>
    `;

    return el;
}

function createLinksCard(links) {
    const el = document.createElement('div');
    el.className = 'result-card';

    const total = links.internal + links.external;

    el.innerHTML = `
        <h3>Links Analysis</h3>
        <div class="result-value">Total: ${total} links</div>
        <div class="links-grid">
            <div class="link-row">
                <div class="link-row-label">Internal</div>
                <div class="link-row-value">${links.internal}</div>
            </div>
            <div class="link-row">
                <div class="link-row-label">External</div>
                <div class="link-row-value">${links.external}</div>
            </div>
            <div class="link-row">
                <div class="link-row-label">Inaccessible</div>
                <div class="link-row-value">${links.inaccessible}</div>
            </div>
        </div>
    `;
    return el;
}

function createLoginFormCard(hasLogin) {
    const el = document.createElement('div');
    el.className = 'result-card';

    const msg = hasLogin ? '✓ Login form detected' : '✗ No login form found';
    const className = hasLogin ? 'has-login' : 'no-login';

    el.innerHTML = `
        <h3>Login Form Detection</h3>
        <div class="login-indicator ${className}">
            ${msg}
        </div>
    `;
    return el;
}

document.addEventListener('DOMContentLoaded', () => {
    inputEl.focus();
});

inputEl.addEventListener('input', () => {
    if (errorSection.style.display === 'block') {
        hideError();
    }
});
