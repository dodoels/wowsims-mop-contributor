import { app, BrowserWindow, dialog, shell } from 'electron';
import { spawn, ChildProcess } from 'child_process';
import * as path from 'path';
import axios from 'axios';
import * as net from 'net';
import * as fs from 'fs';

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
const isDevelopment = process.argv.includes('--dev') || process.env.NODE_ENV === 'development';

const CONFIG: AppConfig = {
    serverUrl: 'http://127.0.0.1:3333', // Will be updated dynamically
    healthCheckUrl: 'http://127.0.0.1:3333/version', // Will be updated dynamically
    executableName: app.isPackaged
        ? (process.platform === 'win32' ? 'server.exe' : 'wowsimmop')
        : (process.platform === 'win32' ? 'wowsimmop-windows.exe' : 'wowsimmop'),
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

        wowsimsProcess = spawn(executablePath, ['--launch=false', `--host=localhost:${currentPort}`], {
            cwd: path.dirname(executablePath),
            detached: false,
            stdio: ['ignore', 'pipe', 'pipe']
        });

        // Handle process output
        wowsimsProcess.stdout?.on('data', (data) => {
            const output = data.toString();
            console.log('WoWSims stdout:', output);
        });

        wowsimsProcess.stderr?.on('data', (data) => {
            const output = data.toString();
            console.log('WoWSims stderr:', output);
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
        console.log('Stopping WoWSims process...');

        if (process.platform === 'win32' && wowsimsProcess.pid) {
            // On Windows, use taskkill to ensure clean shutdown
            spawn('taskkill', ['/pid', wowsimsProcess.pid.toString(), '/f', '/t'], { detached: true });
        } else {
            wowsimsProcess.kill('SIGTERM');
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
    mainWindow = new BrowserWindow({
        width: 1400,
        height: 900,
        minWidth: 1200,
        minHeight: 800,
        webPreferences: {
            nodeIntegration: false,
            contextIsolation: true,
            webSecurity: false, // Disable web security to allow NewBeeBox SDK CORS requests
            devTools: isDevelopment, // Enable dev tools in development mode
            preload: path.join(__dirname, 'preload.js')
        },
        icon: path.join(__dirname, '..', 'assets', 'WoW-Simulator-Icon.png'),
        title: 'WoWSims',
        show: false,
        titleBarStyle: 'default',
        autoHideMenuBar: true // Hide the menu bar (File, Edit, View, etc.)
    });

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
        loadingWindow.close();
        showErrorDialog(`Failed to start WoWSims: ${error.message}`);
        app.quit();
    }
}

// App event handlers
app.whenReady().then(initializeApp);

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
    app.quit();
} else {
    app.on('second-instance', () => {
        // Someone tried to run a second instance, focus our window instead
        if (mainWindow) {
            if (mainWindow.isMinimized()) mainWindow.restore();
            mainWindow.focus();
        }
    });
}
