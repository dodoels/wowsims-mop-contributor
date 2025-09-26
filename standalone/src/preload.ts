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

// Declare global types for renderer process
declare global {
    interface Window {
        electronAPI: ElectronAPI;
    }
}
