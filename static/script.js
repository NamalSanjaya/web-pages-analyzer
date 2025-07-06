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

