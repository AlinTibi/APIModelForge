import './style.css';

import {GenerateModels, ValidateJSON, OpenJSONFile, SaveTextFile} from '../wailsjs/go/main/App';

interface GeneratedFile {
    language: string;
    label: string;
    code: string;
    fileName: string;
    modelCount: number;
}

interface GenerateResponse {
    success: boolean;
    error?: string;
    line?: number;
    column?: number;
    files?: GeneratedFile[];
}

interface ValidationResult {
    valid: boolean;
    message?: string;
    error?: string;
    line?: number;
    column?: number;
}

interface OpenFileResult {
    cancelled: boolean;
    content?: string;
    fileName?: string;
    error?: string;
}

interface SaveFileResult {
    cancelled: boolean;
    path?: string;
    error?: string;
}

const jsonInput = document.getElementById('json-input') as HTMLTextAreaElement;
const rootNameInput = document.getElementById('root-name') as HTMLInputElement;
const codeOutput = document.getElementById('code-output') as HTMLElement;
const languageTabs = document.getElementById('language-tabs') as HTMLElement;
const modelCountEl = document.getElementById('model-count') as HTMLElement;
const statusBar = document.querySelector('.status-bar') as HTMLElement;
const statusIcon = document.getElementById('status-icon') as HTMLElement;
const statusText = document.getElementById('status-text') as HTMLElement;

const btnLoad = document.getElementById('btn-load') as HTMLButtonElement;
const btnValidate = document.getElementById('btn-validate') as HTMLButtonElement;
const btnClear = document.getElementById('btn-clear') as HTMLButtonElement;
const btnGenerate = document.getElementById('btn-generate') as HTMLButtonElement;
const btnCopy = document.getElementById('btn-copy') as HTMLButtonElement;
const btnSave = document.getElementById('btn-save') as HTMLButtonElement;

let currentFiles: GeneratedFile[] = [];
let activeLanguage: string | null = null;
let hasGeneratedOnce = false;

function setStatus(message: string, kind: 'idle' | 'success' | 'error' = 'idle'): void {
    statusText.textContent = message;
    statusBar.classList.remove('success-state', 'error-state');
    statusIcon.classList.remove('success', 'error');
    if (kind === 'success') {
        statusBar.classList.add('success-state');
        statusIcon.classList.add('success');
    } else if (kind === 'error') {
        statusBar.classList.add('error-state');
        statusIcon.classList.add('error');
    }
}

function setBusy(busy: boolean): void {
    btnGenerate.disabled = busy;
    btnValidate.disabled = busy;
    btnLoad.disabled = busy;
}

function renderTabs(): void {
    languageTabs.innerHTML = '';
    for (const file of currentFiles) {
        const tab = document.createElement('div');
        tab.className = 'tab' + (file.language === activeLanguage ? ' active' : '');
        tab.textContent = file.label;
        tab.addEventListener('click', () => {
            activeLanguage = file.language;
            renderTabs();
            renderActiveCode();
        });
        languageTabs.appendChild(tab);
    }
}

function renderActiveCode(): void {
    const file = currentFiles.find((f) => f.language === activeLanguage);
    if (!file) {
        codeOutput.textContent = '// Generated code will appear here after you click Generate.';
        modelCountEl.textContent = '';
        btnCopy.disabled = true;
        btnSave.disabled = true;
        return;
    }
    codeOutput.textContent = file.code;
    const noun = file.modelCount === 1 ? 'model' : 'models';
    modelCountEl.textContent = `${file.modelCount} ${noun}`;
    btnCopy.disabled = false;
    btnSave.disabled = false;
}

async function generate(): Promise<void> {
    const text = jsonInput.value;
    if (text.trim() === '') {
        setStatus('Enter or load some JSON first.', 'error');
        return;
    }

    setBusy(true);
    setStatus('Generating…');
    try {
        const res = (await GenerateModels({json: text, rootName: rootNameInput.value})) as unknown as GenerateResponse;
        if (!res.success) {
            currentFiles = [];
            activeLanguage = null;
            renderTabs();
            renderActiveCode();
            const location = res.line ? ` (line ${res.line}, column ${res.column})` : '';
            setStatus(`Invalid JSON${location}: ${res.error}`, 'error');
            return;
        }

        currentFiles = res.files ?? [];
        if (!activeLanguage || !currentFiles.some((f) => f.language === activeLanguage)) {
            activeLanguage = currentFiles[0]?.language ?? null;
        }
        renderTabs();
        renderActiveCode();
        hasGeneratedOnce = true;

        const totalModels = currentFiles[0]?.modelCount ?? 0;
        const noun = totalModels === 1 ? 'model' : 'models';
        setStatus(`Generated ${currentFiles.length} language outputs — ${totalModels} ${noun}.`, 'success');
    } catch (err) {
        setStatus(`Generation failed: ${String(err)}`, 'error');
    } finally {
        setBusy(false);
    }
}

async function validate(): Promise<void> {
    const text = jsonInput.value;
    if (text.trim() === '') {
        setStatus('Enter or load some JSON first.', 'error');
        return;
    }

    setBusy(true);
    try {
        const res = (await ValidateJSON(text)) as unknown as ValidationResult;
        if (res.valid) {
            setStatus(res.message ?? 'Valid JSON.', 'success');
        } else {
            const location = res.line ? ` (line ${res.line}, column ${res.column})` : '';
            setStatus(`Invalid JSON${location}: ${res.error}`, 'error');
        }
    } catch (err) {
        setStatus(`Validation failed: ${String(err)}`, 'error');
    } finally {
        setBusy(false);
    }
}

function clearInput(): void {
    jsonInput.value = '';
    jsonInput.focus();
    setStatus('Input cleared.');
}

async function loadFile(): Promise<void> {
    setBusy(true);
    try {
        const res = (await OpenJSONFile()) as unknown as OpenFileResult;
        if (res.cancelled) {
            return;
        }
        if (res.error) {
            setStatus(`Could not open file: ${res.error}`, 'error');
            return;
        }
        jsonInput.value = res.content ?? '';
        setStatus(`Loaded ${res.fileName ?? 'file'}.`, 'success');
    } catch (err) {
        setStatus(`Could not open file: ${String(err)}`, 'error');
    } finally {
        setBusy(false);
    }
}

async function copyActiveCode(): Promise<void> {
    const file = currentFiles.find((f) => f.language === activeLanguage);
    if (!file) return;
    try {
        await navigator.clipboard.writeText(file.code);
        setStatus(`Copied ${file.label} code to clipboard.`, 'success');
    } catch (err) {
        setStatus(`Copy failed: ${String(err)}`, 'error');
    }
}

async function saveActiveCode(): Promise<void> {
    const file = currentFiles.find((f) => f.language === activeLanguage);
    if (!file) return;
    try {
        const res = (await SaveTextFile(file.fileName, file.code)) as unknown as SaveFileResult;
        if (res.cancelled) {
            return;
        }
        if (res.error) {
            setStatus(`Save failed: ${res.error}`, 'error');
            return;
        }
        setStatus(`Saved to ${res.path}.`, 'success');
    } catch (err) {
        setStatus(`Save failed: ${String(err)}`, 'error');
    }
}

// --- Wiring ---

btnGenerate.addEventListener('click', () => void generate());
btnValidate.addEventListener('click', () => void validate());
btnClear.addEventListener('click', clearInput);
btnLoad.addEventListener('click', () => void loadFile());
btnCopy.addEventListener('click', () => void copyActiveCode());
btnSave.addEventListener('click', () => void saveActiveCode());

document.addEventListener('keydown', (e) => {
    if (e.ctrlKey && e.key === 'Enter') {
        e.preventDefault();
        void generate();
    }
});

// Regenerate automatically once the user changes the root name, but only
// if they've already generated at least once this session.
rootNameInput.addEventListener('change', () => {
    if (hasGeneratedOnce) {
        void generate();
    }
});

// --- Resizable panes ---

const resizer = document.getElementById('resizer') as HTMLElement;
const paneInput = document.getElementById('pane-input') as HTMLElement;
const panes = document.querySelector('.panes') as HTMLElement;

let resizing = false;

resizer.addEventListener('mousedown', () => {
    resizing = true;
    resizer.classList.add('active');
    document.body.style.userSelect = 'none';
});

window.addEventListener('mousemove', (e) => {
    if (!resizing) return;
    const bounds = panes.getBoundingClientRect();
    const minWidth = 240;
    const maxWidth = bounds.width - 240;
    let newWidth = e.clientX - bounds.left;
    newWidth = Math.max(minWidth, Math.min(maxWidth, newWidth));
    paneInput.style.flex = `0 0 ${newWidth}px`;
});

window.addEventListener('mouseup', () => {
    if (!resizing) return;
    resizing = false;
    resizer.classList.remove('active');
    document.body.style.userSelect = '';
});

renderTabs();
renderActiveCode();
jsonInput.focus();
