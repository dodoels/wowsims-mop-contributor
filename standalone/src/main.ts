import axios from 'axios';
import { spawn, ChildProcess, execSync } from 'child_process';
import { app, BrowserWindow, dialog, ipcMain, shell } from 'electron';
import * as fs from 'fs';
import * as net from 'net';
import * as path from 'path';

// Use software rendering for maximum compatibility and to prevent black screen issues
// These switches fix GPU crashes on various Windows systems
app.commandLine.appendSwitch('disable-gpu');
app.commandLine.appendSwitch('disable-gpu-compositing');
app.commandLine.appendSwitch('use-gl', 'swiftshader');
app.commandLine.appendSwitch('disable-gpu-sandbox');
app.commandLine.appendSwitch('no-sandbox');
app.commandLine.appendSwitch('disable-dev-shm-usage');
// Disable V8 code caching which can cause issues with some GPU drivers
app.commandLine.appendSwitch('v8-cache-options', 'none');
// Additional GPU stability options
app.commandLine.appendSwitch('disable-software-rasterizer', 'false');
app.commandLine.appendSwitch('enable-native-gpu-memory-buffers', 'true');

/**
 * Application configuration
 */
interface AppConfig {
    serverUrl: string;
    healthCheckUrl: string;
    readonly executableName: string;
    readonly startupTimeout: number;
    readonly healthCheckInterval: number;
    readonly initialDelay: number;
    readonly preferredPort: number;
    readonly maxPortTries: number;
}

// Check if we're in development mode (Electron is run with --dev flag)
const isDevelopment = !app.isPackaged;

const CONFIG: AppConfig = {
    serverUrl: 'http://127.0.0.1:3333', // Will be updated dynamically
    healthCheckUrl: 'http://127.0.0.1:3333/version', // Will be updated dynamically
    executableName: app.isPackaged
        ? (process.platform === 'win32' ? 'server.exe' : 'wowsimmop') // this is for release package
        : (process.platform === 'win32' ? 'wowsimmop-windows.exe' : 'wowsimmop'), // this is for local dev
    startupTimeout: 30000, // 30 seconds
    healthCheckInterval: 2000, // 2 seconds
    initialDelay: 1000, // 1 seconds
    preferredPort: 3333,
    maxPortTries: 10 // Try ports 3333-3342
};

/**
 * Global application state
 */
let mainWindow: BrowserWindow | null = null;
let wowsimsProcess: ChildProcess | null = null;
let isQuitting = false;
let currentPort = CONFIG.preferredPort;

/**
 * Check if a port is available
 */
function isPortAvailable(port: number): Promise<boolean> {
    return new Promise((resolve) => {
        const server = net.createServer();

        server.listen(port, '127.0.0.1', () => {
            server.once('close', () => {
                resolve(true);
            });
            server.close();
        });

        server.on('error', () => {
            resolve(false);
        });
    });
}

/**
 * Try to clean up any orphaned processes on specific ports
 */
async function cleanupOrphanedProcesses(): Promise<void> {
    if (process.platform !== 'win32') {
        return;
    }

    console.log('Attempting to clean up orphaned processes on ports 3333-3342...');

    for (let port = 3333; port <= 3342; port++) {
        try {
            // Use netstat to find PIDs listening on each port
            const output = execSync(`netstat -ano | findstr :${port}`, { encoding: 'utf8', windowsHide: true }).toString();

            if (output) {
                // Extract PIDs and kill them
                const lines = output.split('\n').filter((line: string) => line.includes('LISTENING'));
                for (const line of lines) {
                    const match = line.trim().split(/\s+/);
                    const pid = match[match.length - 1];
                    if (pid && pid !== '-1' && !isNaN(parseInt(pid))) {
                        try {
                            console.log(`Killing orphaned process (PID: ${pid}) on port ${port}`);
                            spawn('taskkill', ['/pid', pid, '/f']);
                        } catch (err) {
                            // Ignore errors, process might already be dead
                        }
                    }
                }
            }
        } catch (err) {
            // Ignore netstat errors
        }
    }

    console.log('Orphan cleanup completed');
}

/**
 * Find an available port starting from the preferred port
 */
async function findAvailablePort(): Promise<number> {
    for (let i = 0; i < CONFIG.maxPortTries; i++) {
        const port = CONFIG.preferredPort + i;
        if (await isPortAvailable(port)) {
            console.log(`Found available port: ${port}`);
            return port;
        }
        console.log(`Port ${port} is occupied, trying next...`);
    }
    throw new Error(`No available ports found in range ${CONFIG.preferredPort}-${CONFIG.preferredPort + CONFIG.maxPortTries - 1}`);
}

/**
 * Update configuration with the selected port
 */
function updateConfigWithPort(port: number): void {
    currentPort = port;
    CONFIG.serverUrl = `http://127.0.0.1:${port}`;
    CONFIG.healthCheckUrl = `http://127.0.0.1:${port}/version`;
    console.log(`Updated configuration to use port ${port}`);
}

/**
 * Check if the WoWSims server is responding
 */
async function checkServerHealth(): Promise<boolean> {
    try {
        console.log('Checking server health at:', CONFIG.healthCheckUrl);
        const response = await axios.get(CONFIG.healthCheckUrl, {
            timeout: 5000,
            headers: {
                'User-Agent': 'WoWSims-Electron-App'
            }
        });
        console.log('Health check SUCCESS! Response:', response.status, response.data);
        return response.status === 200;
    } catch (error: any) {
        console.log('Health check failed:', error.code || error.message);
        return false;
    }
}

/**
 * Wait for the WoWSims server to be ready
 */
async function waitForServer(timeout: number = CONFIG.startupTimeout): Promise<void> {
    const startTime = Date.now();

    while (Date.now() - startTime < timeout) {
        if (await checkServerHealth()) {
            console.log('WoWSims server is ready!');
            return;
        }
        await new Promise(resolve => setTimeout(resolve, CONFIG.healthCheckInterval));
    }

    throw new Error('Server startup timeout');
}

/**
 * Start the WoWSims process
 */
function startWowSimsProcess(): Promise<void> {
    return new Promise((resolve, reject) => {
        // Simple approach: look for executable in the same directory as the Electron app
        let executablePath: string;

        if (isDevelopment) {
            // Development: look in project root
            // __dirname in dev points to standalone/dist, so go up one level to standalone, then up one more to project root
            executablePath = path.join(__dirname, '..', '..', CONFIG.executableName);
        } else {
            // Production: executable should be in the app root directory (same level as the main exe)
            executablePath = path.join(path.dirname(process.execPath), CONFIG.executableName);
        }
        console.log('Starting WoWSims process:', executablePath);

        // Check if the executable exists before trying to spawn it
        if (!fs.existsSync(executablePath)) {
            const errorMsg = `WoWSims server not found at: ${executablePath}\n\n`;
            console.error(errorMsg);
            dialog.showErrorBox('WoWSims Server Not Found', errorMsg);
            reject(new Error(`Executable not found: ${executablePath}`));
            return;
        }

        console.log(`Spawning server at ${currentPort} with args: --launch=false --host=localhost:${currentPort}`);

        wowsimsProcess = spawn(executablePath, ['--launch=false', `--host=localhost:${currentPort}`], {
            cwd: path.dirname(executablePath),
            detached: false,
            stdio: ['pipe', 'pipe', 'pipe']  // Use 'pipe' for stdin so we can close it immediately
        });

        // Close stdin immediately to signal EOF to the process
        // This prevents the server from trying to read interactive commands
        if (wowsimsProcess.stdin) {
            wowsimsProcess.stdin.end();
        }

        // Handle process output
        wowsimsProcess.stdout?.on('data', data => {
            const output = data.toString();
            console.log('[SERVER STDOUT]', output.trim());
        });

        wowsimsProcess.stderr?.on('data', data => {
            const output = data.toString();
            console.log('[SERVER STDERR]', output.trim());
        });

        // Handle process errors
        wowsimsProcess.on('error', (error) => {
            console.error('Failed to start WoWSims process:', error);
            reject(error);
        });

        // Handle process exit
        wowsimsProcess.on('exit', (code, signal) => {
            console.log(`WoWSims process exited with code ${code}, signal ${signal}`);
            wowsimsProcess = null;

            if (!isQuitting && mainWindow) {
                showErrorDialog('WoWSims process has stopped unexpectedly. The application will close.');
                app.quit();
            }
        });

        // Give the process a moment to start
        setTimeout(() => {
            if (wowsimsProcess && !wowsimsProcess.killed) {
                resolve();
            } else {
                reject(new Error('Process failed to start'));
            }
        }, 3000);
    });
}

/**
 * Stop the WoWSims process
 */
function stopWowSimsProcess(): void {
    if (wowsimsProcess && !wowsimsProcess.killed) {
        console.log('Stopping WoWSims process (PID: ' + wowsimsProcess.pid + ')...');

        if (process.platform === 'win32' && wowsimsProcess.pid) {
            // On Windows, use taskkill to force termination
            // /f = force, /t = terminate tree (process and children)
            try {
                const taskkill = spawn('taskkill', ['/pid', wowsimsProcess.pid.toString(), '/f', '/t']);

                taskkill.on('close', code => {
                    console.log('taskkill exited with code:', code);
                });

                taskkill.on('error', err => {
                    console.error('taskkill error:', err);
                });

                // Also try to kill the process directly
                wowsimsProcess.kill('SIGKILL');
            } catch (err) {
                console.error('Error killing process:', err);
                wowsimsProcess.kill('SIGKILL');
            }
        } else {
            // On non-Windows, use SIGKILL for more reliable termination
            wowsimsProcess.kill('SIGKILL');
        }

        wowsimsProcess = null;
    }
}

/**
 * Show error dialog to user
 */
function showErrorDialog(message: string): void {
    dialog.showErrorBox('WoWSims MoP Error', message);
}

/**
 * Create the main application window
 */
function createMainWindow(): void {
    // Use .ico for Windows, .png for other platforms
    const iconPath = process.platform === 'win32'
        ? path.join(__dirname, '..', 'assets', 'WoW-Simulator-Icon.png')
        : path.join(__dirname, '..', 'assets', 'WoW-Simulator-Icon.png');

    mainWindow = new BrowserWindow({
        width: 1400,
        height: 900,
        minWidth: 1200,
        minHeight: 800,
        webPreferences: {
            nodeIntegration: false,
            contextIsolation: true,
            webSecurity: false, // Disable web security to allow CORS requests
            devTools: isDevelopment, // Enable dev tools in development mode
            preload: path.join(__dirname, 'preload.js')
        },
        icon: iconPath,
        title: 'WoWSims MoP',
        show: false,
        titleBarStyle: 'default',
        autoHideMenuBar: true // Hide the menu bar (File, Edit, View, etc.)
    });

    // Set taskbar app ID on Windows for proper icon display
    if (process.platform === 'win32') {
        mainWindow.setAppDetails({
            appId: 'WoWSims-MoP',
            appIconPath: iconPath,
            appIconIndex: 0
        });
    }

    // Completely remove the menu bar in production (prevents Alt key from showing it)
    if (!isDevelopment) {
        mainWindow.setMenu(null);
    }

    // Intercept all requests to add authorization headers
    mainWindow.webContents.session.webRequest.onBeforeSendHeaders((details, callback) => {
        if (details.url.startsWith(`http://127.0.0.1:${currentPort}`)) {
            details.requestHeaders['X-WoWSims-Client'] = 'Electron-App';
            details.requestHeaders['X-WoWSims-Auth'] = 'wowsims-desktop-client';
        }
        callback({ requestHeaders: details.requestHeaders });
    });

    // Load the WoWSims application
    mainWindow.loadURL(CONFIG.serverUrl);

    // Show window when ready
    mainWindow.once('ready-to-show', () => {
        mainWindow?.show();

        // Open dev tools in development mode
        if (isDevelopment) {
            mainWindow?.webContents.openDevTools();
        }

        // Ensure the window gets focus
        if (process.platform === 'darwin') {
            app.dock?.show();
        }

        // Force focus and bring to front
        mainWindow?.focus();
        mainWindow?.moveTop();

        // Additional focus methods for better reliability
        if (process.platform === 'win32') {
            mainWindow?.setAlwaysOnTop(true);
            mainWindow?.setAlwaysOnTop(false);
        }
    });

    // Handle window closed
    mainWindow.on('closed', () => {
        mainWindow = null;
    });

    // Handle external links
    mainWindow.webContents.setWindowOpenHandler(({ url }) => {
        shell.openExternal(url);
        return { action: 'deny' };
    });

    // Prevent navigation away from our app
    mainWindow.webContents.on('will-navigate', (event, navigationUrl) => {
        const parsedUrl = new URL(navigationUrl);
        if (parsedUrl.origin !== CONFIG.serverUrl) {
            event.preventDefault();
            shell.openExternal(navigationUrl);
        }
    });

    // Fix for input focus issues after window has been inactive
    // When window regains focus, restore input capability
    mainWindow.on('focus', () => {
        mainWindow?.webContents.send('window-focus');
    });

    // Handle visibility changes to restore input focus if needed
    mainWindow.on('show', () => {
        mainWindow?.webContents.send('window-show');
    });

    // Ensure input elements can receive focus when window is focused
    mainWindow.on('enter-full-screen', () => {
        mainWindow?.webContents.send('window-focus');
    });

    mainWindow.on('leave-full-screen', () => {
        mainWindow?.webContents.send('window-focus');
    });
}

/**
 * Setup IPC handlers for renderer process communication
 */
function setupIPCHandlers(): void {
    // Handle opening external URLs
    ipcMain.handle('open-external', async (_event, url: string) => {
        await shell.openExternal(url);
    });
}

/**
 * Create loading window
 */
function createLoadingWindow(): BrowserWindow {
    const loadingWindow = new BrowserWindow({
        width: 400,
        height: 300,
        frame: false,
        alwaysOnTop: true,
        transparent: true,
        webPreferences: {
            nodeIntegration: false,
            contextIsolation: true
        }
    });

    const loadingHtml = `
        <!DOCTYPE html>
        <html>
        <head>
            <style>
                body {
                    margin: 0;
                    padding: 0;
                    background: rgba(0, 0, 0, 0.8);
                    color: white;
                    font-family: Arial, sans-serif;
                    display: flex;
                    flex-direction: column;
                    justify-content: center;
                    align-items: center;
                    height: 100vh;
                    border-radius: 10px;
                }
                .spinner {
                    border: 4px solid #f3f3f3;
                    border-top: 4px solid #3498db;
                    border-radius: 50%;
                    width: 40px;
                    height: 40px;
                    animation: spin 2s linear infinite;
                    margin-bottom: 20px;
                }
                @keyframes spin {
                    0% { transform: rotate(0deg); }
                    100% { transform: rotate(360deg); }
                }
                h2 { margin: 10px 0; }
                p { margin: 5px 0; text-align: center; }
            </style>
        </head>
        <body>
            <div class="spinner"></div>
            <h2>WoWSims</h2>
            <p>Launching...</p>
            <p>Please wait</p>
        </body>
        </html>
    `;

    loadingWindow.loadURL(`data:text/html;charset=utf-8,${encodeURIComponent(loadingHtml)}`);
    return loadingWindow;
}

/**
 * Initialize the application
 */
async function initializeApp(): Promise<void> {
    console.log('Electron app ready, starting WoWSims...');

    // Setup IPC handlers first
    setupIPCHandlers();

    const loadingWindow = createLoadingWindow();

    try {
        // Find an available port
        const availablePort = await findAvailablePort();
        updateConfigWithPort(availablePort);

        // Start WoWSims process with the available port
        await startWowSimsProcess();
        console.log('WoWSims process started, giving server time to initialize...');

        // Give the server time to start up before health checking
        await new Promise(resolve => setTimeout(resolve, CONFIG.initialDelay));
        console.log('Starting health checks...');

        // Wait for server to be ready
        await waitForServer();
        console.log('Server is ready, creating main window...');

        // Close loading window and create main window
        loadingWindow.close();
        createMainWindow();

        // Refocus the app after a brief delay to ensure it gets focus after WoWSims executable
        setTimeout(() => {
            if (mainWindow) {
                mainWindow.focus();
                mainWindow.moveTop();
                if (process.platform === 'win32') {
                    mainWindow.setAlwaysOnTop(true);
                    mainWindow.setAlwaysOnTop(false);
                }
            }
        }, 1000);

    } catch (error: any) {
        console.error('Failed to start application:', error);
        console.error('Stack trace:', error.stack);
        loadingWindow.close();
        showErrorDialog(`Failed to start WoWSims: ${error.message}`);
        app.quit();
    }
}

// App event handlers
app.whenReady().then(async () => {
    // Clean up any orphaned processes before starting
    await cleanupOrphanedProcesses();
    await initializeApp();
});

app.on('window-all-closed', () => {
    isQuitting = true;
    stopWowSimsProcess();
    app.quit();
});

app.on('before-quit', () => {
    isQuitting = true;
    stopWowSimsProcess();
});

app.on('activate', () => {
    if (BrowserWindow.getAllWindows().length === 0) {
        createMainWindow();
    }
});

// Prevent multiple instances
const gotTheLock = app.requestSingleInstanceLock();

if (!gotTheLock) {
    console.log('Another instance is already running');

    // Try to clean up orphaned processes
    cleanupOrphanedProcesses().then(() => {
        // Show error dialog
        dialog.showErrorBox(
            'WoWSims MoP already running',
            'Another instance of WoWSims MoP is already running.\n\n' +
            'If you see this error but the application is not visible:\n' +
            '1. Open Task Manager (Ctrl+Shift+Esc)\n' +
            '2. Search for "wowsimmop" or "server", right-click and select "End task"\n' +
            '3. Search for "electron", right-click and select "End task"\n' +
            '4. Close Task Manager and restart the application'
        );
        app.quit();
    });
} else {
    app.on('second-instance', () => {
        // Someone tried to run a second instance, focus our window instead
        if (mainWindow) {
            if (mainWindow.isMinimized()) mainWindow.restore();
            mainWindow.focus();
        }
    });
}
