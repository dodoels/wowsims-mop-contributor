import { contextBridge, ipcRenderer } from 'electron';

/**
 * Electron API interface exposed to renderer process
 */
interface ElectronAPI {
    openExternal: (url: string) => Promise<void>;
    getVersion: () => string;
}

// Expose protected methods that allow the renderer process to use
// the ipcRenderer without exposing the entire object
contextBridge.exposeInMainWorld('electronAPI', {
    /**
     * Open URL in external browser
     */
    openExternal: async (url: string): Promise<void> => {
        await ipcRenderer.invoke('open-external', url);
    },

    /**
     * Get Electron version
     */
    getVersion: (): string => {
        return process.versions.electron;
    }
} as ElectronAPI);

// Log that preload script has loaded
console.log('Preload script loaded successfully');

// Listen for focus events from main process and restore input focus
ipcRenderer.on('window-focus', () => {
    // Try to restore focus to active input or first available input
    const activeElement = document.activeElement as HTMLElement;
    if (activeElement && activeElement.tagName === 'INPUT') {
        (activeElement as HTMLInputElement).focus();
    } else {
        // Find and focus first available input or textarea
        const inputs = document.querySelectorAll('input, textarea, [contenteditable="true"]') as NodeListOf<HTMLElement>;
        if (inputs.length > 0) {
            inputs[0].focus();
        }
    }
});

ipcRenderer.on('window-show', () => {
    // Re-enable pointer events in case they got disabled
    document.body.style.pointerEvents = 'auto';
    const root = document.getElementById('root');
    if (root) {
        root.style.pointerEvents = 'auto';
    }
});

// Declare global types for renderer process
declare global {
    interface Window {
        electronAPI: ElectronAPI;
    }
}
