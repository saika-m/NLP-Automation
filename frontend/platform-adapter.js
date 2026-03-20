// Platform Adapter - Handles differences between Electron and Web environments
// Version 2.0.0

class PlatformAdapter {
  constructor() {
    // Detect environment
    this.isElectron = this.detectElectron();
    this.isWeb = !this.isElectron;
    this.appMode = this.getAppMode();
    
    this.init();
  }

  detectElectron() {
    // Check if we're in an Electron renderer process
    return typeof window !== 'undefined' && 
           typeof window.process === 'object' && 
           window.process.type === 'renderer';
  }

  getAppMode() {
    // Try to get from environment variable or detect automatically
    if (typeof process !== 'undefined' && process.env && process.env.APP_MODE) {
      return process.env.APP_MODE;
    }
    return this.isElectron ? 'desktop' : 'web';
  }

  init() {
    console.log(`Platform Adapter initialized: ${this.appMode} mode`);
    
    if (this.isWeb) {
      this.initWebEnvironment();
    }
  }

  initWebEnvironment() {
    // Initialize web-specific polyfills and alternatives
    this.initStoragePolyfill();
    this.initSystemInfoPolyfill();
    this.initFileSystemPolyfill();
  }

  // Storage Adapter
  initStoragePolyfill() {
    if (this.isWeb && typeof window !== 'undefined') {
      window.ElectronStore = class WebStore {
        constructor() {
          this.prefix = 'nlp-automation-';
        }

        get(key, defaultValue = null) {
          try {
            const item = localStorage.getItem(this.prefix + key);
            return item ? JSON.parse(item) : defaultValue;
          } catch (error) {
            console.warn('Failed to get item from storage:', error);
            return defaultValue;
          }
        }

        set(key, value) {
          try {
            localStorage.setItem(this.prefix + key, JSON.stringify(value));
          } catch (error) {
            console.warn('Failed to set item in storage:', error);
          }
        }

        delete(key) {
          try {
            localStorage.removeItem(this.prefix + key);
          } catch (error) {
            console.warn('Failed to delete item from storage:', error);
          }
        }

        clear() {
          try {
            Object.keys(localStorage).forEach(key => {
              if (key.startsWith(this.prefix)) {
                localStorage.removeItem(key);
              }
            });
          } catch (error) {
            console.warn('Failed to clear storage:', error);
          }
        }
      };
    }
  }

  // System Info Adapter
  initSystemInfoPolyfill() {
    if (this.isWeb && typeof window !== 'undefined') {
      window.SystemInfo = class WebSystemInfo {
        static getSystemInfo() {
          return {
            platform: this.getPlatform(),
            arch: this.getArchitecture(),
            hostname: this.getHostname(),
            totalMemory: this.getTotalMemory(),
            freeMemory: this.getFreeMemory(),
            cpus: this.getCPUs(),
            uptime: this.getUptime(),
            userAgent: navigator.userAgent
          };
        }

        static getPlatform() {
          const platform = navigator.platform.toLowerCase();
          if (platform.includes('win')) return 'win32';
          if (platform.includes('mac')) return 'darwin';
          if (platform.includes('linux')) return 'linux';
          return 'unknown';
        }

        static getArchitecture() {
          // Limited detection in browser
          return navigator.userAgent.includes('x64') || navigator.userAgent.includes('64-bit') ? 'x64' : 'x86';
        }

        static getHostname() {
          return window.location.hostname || 'localhost';
        }

        static getTotalMemory() {
          // Browser can't access actual memory info, return estimated
          return navigator.deviceMemory ? navigator.deviceMemory * 1024 * 1024 * 1024 : 8 * 1024 * 1024 * 1024;
        }

        static getFreeMemory() {
          // Browser can't access actual memory info, return estimated
          return this.getTotalMemory() * 0.6; // Assume 60% free
        }

        static getCPUs() {
          return navigator.hardwareConcurrency || 4;
        }

        static getUptime() {
          return performance.now() / 1000; // Uptime since page load
        }
      };
    }
  }

  // File System Adapter
  initFileSystemPolyfill() {
    if (this.isWeb && typeof window !== 'undefined') {
      window.FileSystemAdapter = class WebFileSystem {
        static async selectDirectory() {
          if ('showDirectoryPicker' in window) {
            try {
              const dirHandle = await window.showDirectoryPicker();
              return { 
                canceled: false, 
                filePaths: [dirHandle.name],
                handle: dirHandle
              };
            } catch (error) {
              if (error.name === 'AbortError') {
                return { canceled: true, filePaths: [] };
              }
              throw error;
            }
          } else {
            // Fallback for browsers without File System Access API
            alert('Directory selection not supported in this browser. Please use the text input.');
            return { canceled: true, filePaths: [] };
          }
        }

        static async selectFile(options = {}) {
          if ('showOpenFilePicker' in window) {
            try {
              const fileHandles = await window.showOpenFilePicker({
                types: options.filters || [],
                multiple: false
              });
              return { 
                canceled: false, 
                filePaths: [fileHandles[0].name],
                handles: fileHandles
              };
            } catch (error) {
              if (error.name === 'AbortError') {
                return { canceled: true, filePaths: [] };
              }
              throw error;
            }
          } else {
            // Fallback using input element
            return new Promise((resolve) => {
              const input = document.createElement('input');
              input.type = 'file';
              input.accept = options.filters ? options.filters.map(f => '.' + f.extensions.join(',.')).join(',') : '*';
              
              input.onchange = (e) => {
                const files = Array.from(e.target.files);
                if (files.length > 0) {
                  resolve({
                    canceled: false,
                    filePaths: files.map(f => f.name),
                    files: files
                  });
                } else {
                  resolve({ canceled: true, filePaths: [] });
                }
              };
              
              input.oncancel = () => {
                resolve({ canceled: true, filePaths: [] });
              };
              
              input.click();
            });
          }
        }

        static async saveFile(options = {}) {
          if ('showSaveFilePicker' in window) {
            try {
              const fileHandle = await window.showSaveFilePicker({
                suggestedName: options.defaultPath || 'untitled.txt',
                types: options.filters || []
              });
              return { 
                canceled: false, 
                filePath: fileHandle.name,
                handle: fileHandle
              };
            } catch (error) {
              if (error.name === 'AbortError') {
                return { canceled: true, filePath: '' };
              }
              throw error;
            }
          } else {
            // Fallback using download
            return { canceled: false, filePath: options.defaultPath || 'untitled.txt' };
          }
        }

        static getHomeDirectory() {
          return '/home/user'; // Simulated home directory for web
        }

        static getDesktopDirectory() {
          return '/home/user/Desktop';
        }

        static getDownloadsDirectory() {
          return '/home/user/Downloads';
        }

        static getDocumentsDirectory() {
          return '/home/user/Documents';
        }
      };
    }
  }

  // IPC Adapter
  getIPCAdapter() {
    if (this.isElectron) {
      const { ipcRenderer } = require('electron');
      return ipcRenderer;
    } else {
      // Web IPC simulation - could be WebSocket or other communication method
      return {
        invoke: async (channel, ...args) => {
          console.log(`Web IPC invoke: ${channel}`, args);
          
          // Handle specific channels with web alternatives
          switch (channel) {
            case 'select-directory':
              return await window.FileSystemAdapter.selectDirectory();
            case 'select-file':
              return await window.FileSystemAdapter.selectFile(args[0]);
            case 'save-file':
              return await window.FileSystemAdapter.saveFile(args[0]);
            case 'show-notification':
              return this.showWebNotification(args[0]);
            case 'get-app-version':
              return '2.0.0-web';
            case 'get-system-info':
              return window.SystemInfo.getSystemInfo();
            default:
              console.warn(`Unhandled IPC channel: ${channel}`);
              return null;
          }
        },
        
        send: (channel, ...args) => {
          console.log(`Web IPC send: ${channel}`, args);
          
          // Handle window controls
          switch (channel) {
            case 'minimize-window':
              console.log('Minimize not available in web mode');
              break;
            case 'maximize-window':
              if (document.documentElement.requestFullscreen) {
                document.documentElement.requestFullscreen();
              }
              break;
            case 'close-window':
              window.close();
              break;
            case 'open-external':
              window.open(args[0], '_blank');
              break;
            default:
              console.warn(`Unhandled IPC send channel: ${channel}`);
          }
        },
        
        on: (channel, callback) => {
          console.log(`Web IPC listener registered for: ${channel}`);
          // In web mode, we could store these and trigger them via other means
        }
      };
    }
  }

  // Shell Adapter
  getShellAdapter() {
    if (this.isElectron) {
      const { shell } = require('electron');
      return shell;
    } else {
      return {
        openExternal: (url) => {
          window.open(url, '_blank');
        },
        showItemInFolder: (path) => {
          console.log(`Show in folder not available in web mode: ${path}`);
          alert(`File location: ${path}`);
        }
      };
    }
  }

  // HTTP Client Adapter
  getHTTPClient() {
    if (this.isElectron) {
      return require('axios');
    } else {
      // Use fetch with axios-like interface
      return {
        get: async (url, config = {}) => {
          const response = await fetch(url, {
            method: 'GET',
            headers: config.headers,
            signal: config.timeout ? AbortSignal.timeout(config.timeout) : undefined
          });
          
          if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
          }
          
          const data = await response.json();
          return { status: response.status, data };
        },
        
        post: async (url, data, config = {}) => {
          const response = await fetch(url, {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
              ...config.headers
            },
            body: JSON.stringify(data),
            signal: config.timeout ? AbortSignal.timeout(config.timeout) : undefined
          });
          
          if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
          }
          
          const responseData = await response.json();
          return { status: response.status, data: responseData };
        }
      };
    }
  }

  // Store Adapter
  getStore() {
    if (this.isElectron) {
      return require('electron-store');
    } else {
      return window.ElectronStore;
    }
  }

  // System Info Adapter
  getSystemInfo() {
    if (this.isElectron) {
      const os = require('os');
      return {
        platform: () => os.platform(),
        arch: () => os.arch(),
        hostname: () => os.hostname(),
        totalmem: () => os.totalmem(),
        freemem: () => os.freemem(),
        cpus: () => os.cpus(),
        uptime: () => os.uptime(),
        homedir: () => os.homedir()
      };
    } else {
      const systemInfo = window.SystemInfo.getSystemInfo();
      return {
        platform: () => systemInfo.platform,
        arch: () => systemInfo.arch,
        hostname: () => systemInfo.hostname,
        totalmem: () => systemInfo.totalMemory,
        freemem: () => systemInfo.freeMemory,
        cpus: () => Array(systemInfo.cpus).fill({ model: 'Web CPU' }),
        uptime: () => systemInfo.uptime,
        homedir: () => window.FileSystemAdapter.getHomeDirectory()
      };
    }
  }

  // Notification adapter
  showWebNotification(options) {
    if ('Notification' in window) {
      if (Notification.permission === 'granted') {
        new Notification(options.title || 'NLP-Automation', {
          body: options.body || '',
          icon: '/favicon.ico',
          silent: options.silent || false
        });
        return true;
      } else if (Notification.permission !== 'denied') {
        Notification.requestPermission().then(permission => {
          if (permission === 'granted') {
            new Notification(options.title || 'NLP-Automation', {
              body: options.body || '',
              icon: '/favicon.ico',
              silent: options.silent || false
            });
          }
        });
      }
    }
    return false;
  }

  // File download helper for web
  downloadFile(filename, content, mimeType = 'text/plain') {
    const blob = new Blob([content], { type: mimeType });
    const url = URL.createObjectURL(blob);
    
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    a.style.display = 'none';
    
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    
    URL.revokeObjectURL(url);
  }

  // Clipboard helper for web
  async writeToClipboard(text) {
    if (navigator.clipboard && navigator.clipboard.writeText) {
      await navigator.clipboard.writeText(text);
    } else {
      // Fallback for older browsers
      const textArea = document.createElement('textarea');
      textArea.value = text;
      textArea.style.position = 'fixed';
      textArea.style.left = '-999999px';
      textArea.style.top = '-999999px';
      
      document.body.appendChild(textArea);
      textArea.focus();
      textArea.select();
      
      try {
        document.execCommand('copy');
      } finally {
        document.body.removeChild(textArea);
      }
    }
  }
}

// Initialize platform adapter
const platformAdapter = new PlatformAdapter();

// Export for use
if (typeof window !== 'undefined') {
  window.platformAdapter = platformAdapter;
}

if (typeof module !== 'undefined' && module.exports) {
  module.exports = PlatformAdapter;
}
