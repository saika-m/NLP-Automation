// Tashi Electron Main Process
// Version 2.0.0

const { app, BrowserWindow, ipcMain, dialog, shell, Menu } = require('electron');
const path = require('path');
const isDev = process.env.NODE_ENV === 'development';

class TashiMain {
  constructor() {
    this.mainWindow = null;
    this.isQuiting = false;
    
    this.init();
  }

  init() {
    // Set app user model id for Windows
    if (process.platform === 'win32') {
      app.setAppUserModelId('com.tashi.desktop');
    }

    // Configure security
    this.setupSecurity();
    
    // Set up event handlers
    this.setupEventHandlers();
    
    // Create application menu
    this.createMenu();
  }

  setupSecurity() {
    // Prevent new window creation from renderer
    app.on('web-contents-created', (event, contents) => {
      contents.on('new-window', (event, url) => {
        event.preventDefault();
        shell.openExternal(url);
      });
    });
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
        webSecurity: !isDev
      },
      titleBarStyle: process.platform === 'darwin' ? 'hiddenInset' : 'default',
      frame: process.platform !== 'darwin',
      backgroundColor: '#f8f9fa'
    });

    // Load the app
    this.mainWindow.loadFile('index.html');

    // Show window when ready
    this.mainWindow.once('ready-to-show', () => {
      this.mainWindow.show();
      
      // Open DevTools in development
      if (isDev) {
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

    // Handle window focus
    this.mainWindow.on('focus', () => {
      this.mainWindow.webContents.send('window-focused');
    });

    this.mainWindow.on('blur', () => {
      this.mainWindow.webContents.send('window-blurred');
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
              this.mainWindow.webContents.send('new-task');
            }
          },
          {
            label: 'Open Working Directory',
            accelerator: 'CmdOrCtrl+O',
            click: () => {
              this.selectDirectory();
            }
          },
          { type: 'separator' },
          {
            label: 'Settings',
            accelerator: 'CmdOrCtrl+,',
            click: () => {
              this.mainWindow.webContents.send('open-settings');
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
              this.mainWindow.webContents.send('switch-view', 'dashboard');
            }
          },
          {
            label: 'Commands',
            accelerator: 'CmdOrCtrl+2',
            click: () => {
              this.mainWindow.webContents.send('switch-view', 'commands');
            }
          },
          {
            label: 'Office Suite',
            accelerator: 'CmdOrCtrl+3',
            click: () => {
              this.mainWindow.webContents.send('switch-view', 'office');
            }
          },
          {
            label: 'System',
            accelerator: 'CmdOrCtrl+4',
            click: () => {
              this.mainWindow.webContents.send('switch-view', 'system');
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
        label: 'Tools',
        submenu: [
          {
            label: 'Refresh All',
            accelerator: 'F5',
            click: () => {
              this.mainWindow.webContents.send('refresh-all');
            }
          },
          {
            label: 'Clear Cache',
            click: () => {
              this.clearCache();
            }
          },
          { type: 'separator' },
          {
            label: 'Export Commands',
            accelerator: 'CmdOrCtrl+E',
            click: () => {
              this.mainWindow.webContents.send('export-commands');
            }
          },
          {
            label: 'Import Template',
            click: () => {
              this.importTemplate();
            }
          }
        ]
      },
      {
        label: 'Help',
        submenu: [
          {
            label: 'About Tashi',
            click: () => {
              this.showAbout();
            }
          },
          {
            label: 'Documentation',
            click: () => {
              shell.openExternal('https://github.com/tashi-app/tashi');
            }
          },
          {
            label: 'Report Bug',
            click: () => {
              shell.openExternal('https://github.com/tashi-app/tashi/issues');
            }
          },
          { type: 'separator' },
          {
            label: 'Toggle Developer Tools',
            accelerator: process.platform === 'darwin' ? 'Alt+Cmd+I' : 'Ctrl+Shift+I',
            click: () => {
              this.mainWindow.webContents.toggleDevTools();
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
      const notification = new Notification({
        title: options.title || 'Tashi',
        body: options.body || '',
        icon: this.getIconPath(),
        silent: options.silent || false
      });
      
      notification.show();
      return true;
    });

    // Show message box
    ipcMain.handle('show-message-box', async (event, options) => {
      const result = await dialog.showMessageBox(this.mainWindow, {
        type: options.type || 'info',
        title: options.title || 'Tashi',
        message: options.message || '',
        detail: options.detail || '',
        buttons: options.buttons || ['OK'],
        defaultId: options.defaultId || 0,
        cancelId: options.cancelId || 0
      });
      return result;
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

  async selectDirectory() {
    const result = await dialog.showOpenDialog(this.mainWindow, {
      properties: ['openDirectory'],
      title: 'Select Working Directory'
    });
    
    if (!result.canceled && result.filePaths.length > 0) {
      this.mainWindow.webContents.send('directory-selected', result.filePaths[0]);
    }
  }

  async importTemplate() {
    const result = await dialog.showOpenDialog(this.mainWindow, {
      properties: ['openFile'],
      filters: [
        { name: 'JSON Files', extensions: ['json'] },
        { name: 'All Files', extensions: ['*'] }
      ],
      title: 'Import Template'
    });
    
    if (!result.canceled && result.filePaths.length > 0) {
      this.mainWindow.webContents.send('import-template', result.filePaths[0]);
    }
  }

  async clearCache() {
    const result = await dialog.showMessageBox(this.mainWindow, {
      type: 'question',
      title: 'Clear Cache',
      message: 'Are you sure you want to clear the application cache?',
      detail: 'This will remove all stored data including settings and templates.',
      buttons: ['Cancel', 'Clear Cache'],
      defaultId: 0,
      cancelId: 0
    });
    
    if (result.response === 1) {
      // Clear cache
      await this.mainWindow.webContents.session.clearStorageData();
      this.mainWindow.webContents.send('cache-cleared');
      
      // Show success message
      await dialog.showMessageBox(this.mainWindow, {
        type: 'info',
        title: 'Cache Cleared',
        message: 'Application cache has been cleared successfully.',
        buttons: ['OK']
      });
    }
  }

  showAbout() {
    const aboutWindow = new BrowserWindow({
      width: 400,
      height: 300,
      resizable: false,
      parent: this.mainWindow,
      modal: true,
      show: false,
      webPreferences: {
        nodeIntegration: true,
        contextIsolation: false
      }
    });

    aboutWindow.loadURL(`data:text/html;charset=utf-8,
      <!DOCTYPE html>
      <html>
        <head>
          <style>
            body {
              font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
              margin: 0;
              padding: 20px;
              background: #f8f9fa;
              text-align: center;
            }
            .logo {
              font-size: 48px;
              margin-bottom: 10px;
            }
            .title {
              font-size: 24px;
              font-weight: bold;
              margin-bottom: 5px;
              color: #1192DC;
            }
            .version {
              font-size: 14px;
              color: #666;
              margin-bottom: 20px;
            }
            .description {
              font-size: 14px;
              color: #333;
              line-height: 1.5;
              margin-bottom: 20px;
            }
            .footer {
              font-size: 12px;
              color: #666;
            }
          </style>
        </head>
        <body>
          <div class="logo">🚀</div>
          <div class="title">Tashi</div>
          <div class="version">Version ${app.getVersion()}</div>
          <div class="description">
            AI-Powered Automation Assistant<br>
            Simplifying complex tasks with intelligent command generation
          </div>
          <div class="footer">
            Made with ❤️ for productivity
          </div>
        </body>
      </html>
    `);

    aboutWindow.once('ready-to-show', () => {
      aboutWindow.show();
    });

    aboutWindow.setMenuBarVisibility(false);
  }
}

// Initialize the application
const tashiMain = new TashiMain();

// Handle certificate errors
app.on('certificate-error', (event, webContents, url, error, certificate, callback) => {
  if (isDev) {
    // In development, ignore certificate errors
    event.preventDefault();
    callback(true);
  } else {
    // In production, use default behavior
    callback(false);
  }
});

// Security: Prevent navigation to external URLs
app.on('web-contents-created', (event, contents) => {
  contents.on('will-navigate', (event, navigationUrl) => {
    const parsedUrl = new URL(navigationUrl);
    
    if (parsedUrl.origin !== 'file://') {
      event.preventDefault();
    }
  });
});

// Export for testing
module.exports = TashiMain;
