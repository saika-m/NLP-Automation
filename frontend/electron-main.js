// NLP-Automation Electron Wrapper - Simple wrapper for the web app
// Version 2.0.0

const { app, BrowserWindow, ipcMain, dialog, shell, Menu } = require('electron');
const path = require('path');

class NLPAutomationElectronApp {
  constructor() {
    this.mainWindow = null;
    this.isQuiting = false;
    this.init();
  }

  init() {
    // Set app user model id for Windows
    if (process.platform === 'win32') {
      app.setAppUserModelId('com.nlp-automation.desktop');
    }

    // Set up event handlers
    this.setupEventHandlers();
    
    // Create application menu
    this.createMenu();
  }

  setupEventHandlers() {
    // App ready
    app.whenReady().then(() => {
      this.createWindow();
      
      // macOS specific - recreate window when dock icon is clicked
      app.on('activate', () => {
        if (BrowserWindow.getAllWindows().length === 0) {
          this.createWindow();
        }
      });
    });

    // Quit when all windows are closed
    app.on('window-all-closed', () => {
      if (process.platform !== 'darwin') {
        app.quit();
      }
    });

    // Before quit
    app.on('before-quit', () => {
      this.isQuiting = true;
    });

    // IPC handlers
    this.setupIpcHandlers();
  }

  createWindow() {
    // Create the browser window
    this.mainWindow = new BrowserWindow({
      width: 1400,
      height: 900,
      minWidth: 1200,
      minHeight: 700,
      show: false,
      icon: this.getIconPath(),
      webPreferences: {
        nodeIntegration: true,
        contextIsolation: false,
        enableRemoteModule: true,
        webSecurity: false // Allow CORS for local development
      },
      titleBarStyle: process.platform === 'darwin' ? 'hiddenInset' : 'default',
      frame: process.platform !== 'darwin',
      backgroundColor: '#f8f9fa'
    });

    // Load the web app (same HTML file)
    this.mainWindow.loadFile('index.html');

    // Show window when ready
    this.mainWindow.once('ready-to-show', () => {
      this.mainWindow.show();
      
      // Open DevTools in development
      if (process.env.NODE_ENV === 'development') {
        this.mainWindow.webContents.openDevTools();
      }
    });

    // Handle window closed
    this.mainWindow.on('closed', () => {
      this.mainWindow = null;
    });

    // Handle window close (minimize to tray on macOS)
    this.mainWindow.on('close', (event) => {
      if (process.platform === 'darwin' && !this.isQuiting) {
        event.preventDefault();
        this.mainWindow.hide();
      }
    });

    // Handle external links
    this.mainWindow.webContents.setWindowOpenHandler(({ url }) => {
      shell.openExternal(url);
      return { action: 'deny' };
    });
  }

  getIconPath() {
    if (process.platform === 'win32') {
      return path.join(__dirname, 'assets', 'icon.ico');
    } else if (process.platform === 'darwin') {
      return path.join(__dirname, 'assets', 'icon.icns');
    } else {
      return path.join(__dirname, 'assets', 'icon.png');
    }
  }

  createMenu() {
    const template = [
      {
        label: 'File',
        submenu: [
          {
            label: 'New Task',
            accelerator: 'CmdOrCtrl+N',
            click: () => {
              this.mainWindow.webContents.executeJavaScript('window.app && window.app.switchView("commands")');
            }
          },
          { type: 'separator' },
          {
            label: 'Settings',
            accelerator: 'CmdOrCtrl+,',
            click: () => {
              this.mainWindow.webContents.executeJavaScript('window.app && window.app.switchView("settings")');
            }
          },
          { type: 'separator' },
          {
            label: 'Quit',
            accelerator: process.platform === 'darwin' ? 'Cmd+Q' : 'Ctrl+Q',
            click: () => {
              this.isQuiting = true;
              app.quit();
            }
          }
        ]
      },
      {
        label: 'Edit',
        submenu: [
          { role: 'undo' },
          { role: 'redo' },
          { type: 'separator' },
          { role: 'cut' },
          { role: 'copy' },
          { role: 'paste' },
          { role: 'selectall' }
        ]
      },
      {
        label: 'View',
        submenu: [
          {
            label: 'Dashboard',
            accelerator: 'CmdOrCtrl+1',
            click: () => {
              this.mainWindow.webContents.executeJavaScript('window.app && window.app.switchView("dashboard")');
            }
          },
          {
            label: 'Commands',
            accelerator: 'CmdOrCtrl+2',
            click: () => {
              this.mainWindow.webContents.executeJavaScript('window.app && window.app.switchView("commands")');
            }
          },
          {
            label: 'System',
            accelerator: 'CmdOrCtrl+3',
            click: () => {
              this.mainWindow.webContents.executeJavaScript('window.app && window.app.switchView("system")');
            }
          },
          { type: 'separator' },
          { role: 'reload' },
          { role: 'forceReload' },
          { role: 'toggleDevTools' },
          { type: 'separator' },
          { role: 'resetZoom' },
          { role: 'zoomIn' },
          { role: 'zoomOut' },
          { type: 'separator' },
          { role: 'togglefullscreen' }
        ]
      },
      {
        label: 'Help',
        submenu: [
          {
            label: 'About NLP-Automation',
            click: () => {
              this.showAbout();
            }
          },
          {
            label: 'Documentation',
            click: () => {
              shell.openExternal('https://github.com/nlp-automation-app/nlp-automation');
            }
          }
        ]
      }
    ];

    // macOS specific menu adjustments
    if (process.platform === 'darwin') {
      template.unshift({
        label: app.getName(),
        submenu: [
          { role: 'about' },
          { type: 'separator' },
          { role: 'services', submenu: [] },
          { type: 'separator' },
          { role: 'hide' },
          { role: 'hideothers' },
          { role: 'unhide' },
          { type: 'separator' },
          { role: 'quit' }
        ]
      });

      // Window menu
      template.push({
        label: 'Window',
        submenu: [
          { role: 'minimize' },
          { role: 'close' }
        ]
      });
    }

    const menu = Menu.buildFromTemplate(template);
    Menu.setApplicationMenu(menu);
  }

  setupIpcHandlers() {
    // Directory selection
    ipcMain.handle('select-directory', async () => {
      const result = await dialog.showOpenDialog(this.mainWindow, {
        properties: ['openDirectory'],
        title: 'Select Working Directory'
      });
      return result;
    });

    // File selection
    ipcMain.handle('select-file', async (event, options = {}) => {
      const result = await dialog.showOpenDialog(this.mainWindow, {
        properties: ['openFile'],
        filters: options.filters || [
          { name: 'All Files', extensions: ['*'] }
        ],
        title: options.title || 'Select File'
      });
      return result;
    });

    // Save file dialog
    ipcMain.handle('save-file', async (event, options = {}) => {
      const result = await dialog.showSaveDialog(this.mainWindow, {
        filters: options.filters || [
          { name: 'All Files', extensions: ['*'] }
        ],
        defaultPath: options.defaultPath || 'untitled',
        title: options.title || 'Save File'
      });
      return result;
    });

    // Window controls
    ipcMain.on('minimize-window', () => {
      if (this.mainWindow) {
        this.mainWindow.minimize();
      }
    });

    ipcMain.on('maximize-window', () => {
      if (this.mainWindow) {
        if (this.mainWindow.isMaximized()) {
          this.mainWindow.unmaximize();
        } else {
          this.mainWindow.maximize();
        }
      }
    });

    ipcMain.on('close-window', () => {
      if (this.mainWindow) {
        this.mainWindow.close();
      }
    });

    // Notifications
    ipcMain.handle('show-notification', (event, options) => {
      const { Notification } = require('electron');
      
      if (Notification.isSupported()) {
        const notification = new Notification({
          title: options.title || 'NLP-Automation',
          body: options.body || '',
          icon: this.getIconPath(),
          silent: options.silent || false
        });
        
        notification.show();
        return true;
      }
      return false;
    });

    // Open external URL
    ipcMain.on('open-external', (event, url) => {
      shell.openExternal(url);
    });

    // Show item in folder
    ipcMain.on('show-item-in-folder', (event, fullPath) => {
      shell.showItemInFolder(fullPath);
    });

    // Get app version
    ipcMain.handle('get-app-version', () => {
      return app.getVersion();
    });

    // Get system info
    ipcMain.handle('get-system-info', () => {
      const os = require('os');
      return {
        platform: os.platform(),
        arch: os.arch(),
        release: os.release(),
        hostname: os.hostname(),
        totalMemory: os.totalmem(),
        freeMemory: os.freemem(),
        cpus: os.cpus().length,
        uptime: os.uptime()
      };
    });
  }

  showAbout() {
    dialog.showMessageBox(this.mainWindow, {
      type: 'info',
      title: 'About NLP-Automation',
      message: 'NLP-Automation v2.0.0',
      detail: 'AI-Powered Automation Assistant\nBuilt with Electron and Web Technologies',
      buttons: ['OK']
    });
  }
}

// Initialize the application
new NLPAutomationElectronApp();

// Handle certificate errors in development
app.on('certificate-error', (event, webContents, url, error, certificate, callback) => {
  if (process.env.NODE_ENV === 'development') {
    event.preventDefault();
    callback(true);
  } else {
    callback(false);
  }
});
