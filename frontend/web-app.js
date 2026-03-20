// NLP-Automation Web Application - Pure Web Implementation
// Version 2.0.0

class NLPAutomationWebApp {
  constructor() {
    // Detect if running in Electron
    this.isElectron = this.detectElectron();
    this.isWeb = !this.isElectron;
    
    // Configuration
    this.config = {
      backendUrl: this.getBackendUrl(),
      appMode: this.isElectron ? 'desktop' : 'web'
    };
    
    // Application state
    this.currentView = 'dashboard';
    this.currentWorkingDir = '';
    this.backendStatus = 'offline';
    this.aiStatus = 'offline';
    this.currentCommands = [];
    this.systemInfo = {};
    this.stats = {
      commandsExecuted: 0,
      filesProcessed: 0,
      timeSaved: 0
    };
    this.templates = new Map();
    this.activities = [];
    this.settings = {
      aiModel: 'gpt-4o-mini',
      aiTemperature: 0.3,
      theme: 'light',
      fontSize: 'medium',
      requireConfirmation: true,
      autoBackup: true
    };
    
    this.init();
  }

  detectElectron() {
    return typeof window !== 'undefined' && 
           typeof window.process === 'object' && 
           window.process.type === 'renderer';
  }

  getBackendUrl() {
    // Try different sources for backend URL
    if (this.isElectron && typeof process !== 'undefined' && process.env) {
      return process.env.BACKEND_URL || 'http://localhost:8080';
    }
    
    // For web, try to get from a global config or use same origin
    if (window.TASHI_CONFIG && window.TASHI_CONFIG.backendUrl) {
      return window.TASHI_CONFIG.backendUrl;
    }
    
    // Default to same origin for web, localhost for development
    const isLocalDevelopment = window.location.hostname === 'localhost' || 
                               window.location.hostname === '127.0.0.1';
    
    return isLocalDevelopment ? 'http://localhost:8080' : window.location.origin;
  }

  async init() {
    console.log(`NLP-Automation Web App initialized in ${this.config.appMode} mode`);
    
    await this.initStorage();
    await this.loadSettings();
    await this.setupEventListeners();
    await this.initializeSystemInfo();
    await this.checkBackendStatus();
    await this.loadInitialData();
    
    // Set up periodic updates
    setInterval(() => this.updateSystemInfo(), 5000);
    setInterval(() => this.checkBackendStatus(), 10000);
    setInterval(() => this.updateClock(), 1000);
    
    this.log('info', 'NLP-Automation application initialized successfully');
    
    // Add mode indicator
    this.addModeIndicator();
  }

  addModeIndicator() {
    const indicator = document.createElement('div');
    indicator.className = 'mode-indicator';
    indicator.textContent = this.config.appMode === 'desktop' ? '🖥️ Desktop' : '🌐 Web';
    indicator.style.cssText = `
      position: fixed;
      top: 10px;
      right: 10px;
      padding: 5px 10px;
      background: var(--primary-color, #1192DC);
      color: white;
      border-radius: 15px;
      font-size: 12px;
      font-weight: 500;
      z-index: 1000;
      pointer-events: none;
      opacity: 0.8;
    `;
    document.body.appendChild(indicator);
    
    // Fade out after 3 seconds
    setTimeout(() => {
      indicator.style.transition = 'opacity 1s';
      indicator.style.opacity = '0';
      setTimeout(() => indicator.remove(), 1000);
    }, 3000);
  }

  // Storage Management (Web-compatible)
  async initStorage() {
    this.storage = {
      get: (key, defaultValue = null) => {
        try {
          const item = localStorage.getItem(`nlp-automation-${key}`);
          return item ? JSON.parse(item) : defaultValue;
        } catch (error) {
          console.warn('Failed to get item from storage:', error);
          return defaultValue;
        }
      },
      
      set: (key, value) => {
        try {
          localStorage.setItem(`nlp-automation-${key}`, JSON.stringify(value));
        } catch (error) {
          console.warn('Failed to set item in storage:', error);
        }
      },
      
      delete: (key) => {
        try {
          localStorage.removeItem(`nlp-automation-${key}`);
        } catch (error) {
          console.warn('Failed to delete item from storage:', error);
        }
      }
    };
  }

  async loadSettings() {
    const savedSettings = this.storage.get('settings', {});
    this.settings = { ...this.settings, ...savedSettings };
    this.applySettings();
  }

  async saveSettings() {
    this.storage.set('settings', this.settings);
    this.applySettings();
  }

  applySettings() {
    document.documentElement.setAttribute('data-theme', this.settings.theme);
    document.documentElement.setAttribute('data-font-size', this.settings.fontSize);
    this.updateSettingsUI();
  }

  updateSettingsUI() {
    const elements = {
      aiModel: document.getElementById('ai-model'),
      aiTemperature: document.getElementById('ai-temperature'),
      theme: document.getElementById('theme'),
      fontSize: document.getElementById('font-size'),
      requireConfirmation: document.getElementById('require-confirmation'),
      autoBackup: document.getElementById('auto-backup')
    };

    if (elements.aiModel) elements.aiModel.value = this.settings.aiModel;
    if (elements.aiTemperature) {
      elements.aiTemperature.value = this.settings.aiTemperature;
      const rangeValue = document.querySelector('.range-value');
      if (rangeValue) rangeValue.textContent = this.settings.aiTemperature;
    }
    if (elements.theme) elements.theme.value = this.settings.theme;
    if (elements.fontSize) elements.fontSize.value = this.settings.fontSize;
    if (elements.requireConfirmation) elements.requireConfirmation.checked = this.settings.requireConfirmation;
    if (elements.autoBackup) elements.autoBackup.checked = this.settings.autoBackup;
  }

  // System Info (Web-compatible)
  async initializeSystemInfo() {
    this.systemInfo = {
      platform: this.getPlatform(),
      arch: this.getArchitecture(),
      hostname: this.getHostname(),
      totalMemory: this.getTotalMemory(),
      freeMemory: this.getFreeMemory(),
      cpus: this.getCPUs(),
      uptime: this.getUptime(),
      userAgent: navigator.userAgent
    };

    // Set default working directory
    this.currentWorkingDir = this.getDefaultWorkingDir();
    const workingDirInput = document.getElementById('working-dir');
    if (workingDirInput) {
      workingDirInput.value = this.currentWorkingDir;
    }

    this.updateSystemDisplay();
  }

  getPlatform() {
    const platform = navigator.platform.toLowerCase();
    if (platform.includes('win')) return 'win32';
    if (platform.includes('mac')) return 'darwin';
    if (platform.includes('linux')) return 'linux';
    return 'unknown';
  }

  getArchitecture() {
    return navigator.userAgent.includes('x64') || navigator.userAgent.includes('64-bit') ? 'x64' : 'x86';
  }

  getHostname() {
    return window.location.hostname || 'localhost';
  }

  getTotalMemory() {
    return navigator.deviceMemory ? navigator.deviceMemory * 1024 * 1024 * 1024 : 8 * 1024 * 1024 * 1024;
  }

  getFreeMemory() {
    return this.getTotalMemory() * 0.6; // Estimate 60% free
  }

  getCPUs() {
    return navigator.hardwareConcurrency || 4;
  }

  getUptime() {
    return performance.now() / 1000;
  }

  getDefaultWorkingDir() {
    if (this.isElectron) {
      return '/Users/' + (process.env.USER || 'user');
    }
    return '/home/user'; // Simulated for web
  }

  async updateSystemInfo() {
    this.systemInfo.freeMemory = this.getFreeMemory();
    this.systemInfo.uptime = this.getUptime();
    this.updateSystemDisplay();
  }

  updateSystemDisplay() {
    const elements = {
      platformInfo: document.getElementById('platform-info'),
      memoryInfo: document.getElementById('memory-info'),
      systemPlatform: document.getElementById('system-platform'),
      systemArch: document.getElementById('system-arch'),
      systemMemory: document.getElementById('system-memory')
    };

    if (elements.platformInfo) {
      elements.platformInfo.textContent = `${this.systemInfo.platform} ${this.systemInfo.arch}`;
    }

    if (elements.memoryInfo) {
      const memoryUsage = ((this.systemInfo.totalMemory - this.systemInfo.freeMemory) / this.systemInfo.totalMemory * 100).toFixed(1);
      elements.memoryInfo.textContent = `RAM: ${memoryUsage}%`;
    }

    if (elements.systemPlatform) {
      elements.systemPlatform.textContent = this.systemInfo.platform;
    }

    if (elements.systemArch) {
      elements.systemArch.textContent = this.systemInfo.arch;
    }

    if (elements.systemMemory) {
      const totalGB = (this.systemInfo.totalMemory / 1024 / 1024 / 1024).toFixed(1);
      const freeGB = (this.systemInfo.freeMemory / 1024 / 1024 / 1024).toFixed(1);
      elements.systemMemory.textContent = `${freeGB}GB free of ${totalGB}GB`;
    }

    this.updatePerformanceMetrics();
  }

  updatePerformanceMetrics() {
    const memoryUsage = ((this.systemInfo.totalMemory - this.systemInfo.freeMemory) / this.systemInfo.totalMemory * 100);
    
    const memoryBar = document.querySelector('.metric-item:last-child .metric-fill');
    const memoryValue = document.querySelector('.metric-item:last-child .metric-value');
    
    if (memoryBar) {
      memoryBar.style.width = `${memoryUsage}%`;
    }
    
    if (memoryValue) {
      memoryValue.textContent = `${memoryUsage.toFixed(1)}%`;
    }
  }

  updateClock() {
    const timeInfo = document.getElementById('time-info');
    if (timeInfo) {
      const now = new Date();
      timeInfo.textContent = now.toLocaleTimeString('en-US', { 
        hour12: false,
        hour: '2-digit',
        minute: '2-digit'
      });
    }
  }

  // Backend Communication
  async checkBackendStatus() {
    try {
      const response = await fetch(`${this.config.backendUrl}/api/health`, {
        method: 'GET',
        headers: { 'Content-Type': 'application/json' },
        signal: AbortSignal.timeout(3000)
      });
      
      if (response.ok) {
        this.setBackendStatus('online');
        await this.checkAIStatus();
      } else {
        this.setBackendStatus('offline');
        this.setAIStatus('offline');
      }
    } catch (error) {
      this.setBackendStatus('offline');
      this.setAIStatus('offline');
    }
  }

  async checkAIStatus() {
    try {
      const response = await fetch(`${this.config.backendUrl}/api/ai/status`, {
        method: 'GET',
        headers: { 'Content-Type': 'application/json' },
        signal: AbortSignal.timeout(3000)
      });
      
      if (response.ok) {
        const data = await response.json();
        this.setAIStatus(data.status === 'online' ? 'online' : 'offline');
      } else {
        this.setAIStatus('offline');
      }
    } catch (error) {
      this.setAIStatus('offline');
    }
  }

  setBackendStatus(status) {
    this.backendStatus = status;
    const statusElement = document.getElementById('backend-status');
    if (statusElement) {
      statusElement.className = `status-indicator ${status}`;
      statusElement.querySelector('.status-text').textContent = `Backend: ${status === 'online' ? 'Online' : 'Offline'}`;
    }
  }

  setAIStatus(status) {
    this.aiStatus = status;
    const statusElement = document.getElementById('ai-status');
    if (statusElement) {
      statusElement.className = `status-indicator ${status}`;
      statusElement.querySelector('.status-text').textContent = `AI: ${status === 'online' ? 'Online' : 'Offline'}`;
    }
  }

  // HTTP Client (Web-compatible)
  async apiCall(method, endpoint, data = null, options = {}) {
    const url = `${this.config.backendUrl}${endpoint}`;
    const config = {
      method: method.toUpperCase(),
      headers: {
        'Content-Type': 'application/json',
        ...options.headers
      },
      signal: options.timeout ? AbortSignal.timeout(options.timeout) : undefined
    };

    if (data && ['POST', 'PUT', 'PATCH'].includes(config.method)) {
      config.body = JSON.stringify(data);
    }

    const response = await fetch(url, config);
    
    if (!response.ok) {
      const errorText = await response.text();
      let errorData;
      try {
        errorData = JSON.parse(errorText);
      } catch {
        errorData = { message: errorText || `HTTP ${response.status}: ${response.statusText}` };
      }
      
      const error = new Error(errorData.message || `HTTP ${response.status}`);
      error.response = { status: response.status, data: errorData };
      throw error;
    }

    return {
      status: response.status,
      data: await response.json()
    };
  }

  // File Operations (Web-compatible)
  async selectDirectory() {
    if (this.isElectron) {
      // Use Electron IPC
      const { ipcRenderer } = require('electron');
      return await ipcRenderer.invoke('select-directory');
    } else {
      // Use File System Access API or fallback
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
        // Fallback - show input dialog
        const path = prompt('Enter directory path:', this.currentWorkingDir);
        if (path) {
          return { canceled: false, filePaths: [path] };
        }
        return { canceled: true, filePaths: [] };
      }
    }
  }

  // Clipboard Operations
  async copyToClipboard(text) {
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

  // File Download
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

  // Notifications
  showNotification(title, body, options = {}) {
    if (this.isElectron) {
      const { ipcRenderer } = require('electron');
      return ipcRenderer.invoke('show-notification', { title, body, ...options });
    } else {
      if ('Notification' in window) {
        if (Notification.permission === 'granted') {
          new Notification(title, {
            body,
            icon: '/favicon.ico',
            ...options
          });
          return true;
        } else if (Notification.permission !== 'denied') {
          Notification.requestPermission().then(permission => {
            if (permission === 'granted') {
              new Notification(title, {
                body,
                icon: '/favicon.ico',
                ...options
              });
            }
          });
        }
      }
      return false;
    }
  }

  // Window Controls
  minimizeWindow() {
    if (this.isElectron) {
      const { ipcRenderer } = require('electron');
      ipcRenderer.send('minimize-window');
    } else {
      console.log('Minimize not available in web mode');
    }
  }

  // External Links
  openExternal(url) {
    if (this.isElectron) {
      const { shell } = require('electron');
      shell.openExternal(url);
    } else {
      window.open(url, '_blank');
    }
  }

  // Event Listeners Setup
  async setupEventListeners() {
    // Navigation
    document.querySelectorAll('[data-view]').forEach(navItem => {
      navItem.addEventListener('click', (e) => {
        e.preventDefault();
        const view = e.currentTarget.dataset.view;
        this.switchView(view);
      });
    });

    // Working directory
    const workingDirInput = document.getElementById('working-dir');
    const selectDirBtn = document.getElementById('select-dir-btn');
    
    if (workingDirInput) {
      workingDirInput.addEventListener('change', (e) => {
        this.setWorkingDirectory(e.target.value);
      });
    }

    if (selectDirBtn) {
      selectDirBtn.addEventListener('click', async () => {
        try {
          const result = await this.selectDirectory();
          if (!result.canceled && result.filePaths.length > 0) {
            const path = result.filePaths[0];
            workingDirInput.value = path;
            await this.setWorkingDirectory(path);
          }
        } catch (error) {
          this.showToast('error', 'Failed to select directory');
        }
      });
    }

    // Command generation
    const generateBtn = document.getElementById('generate-commands-btn');
    if (generateBtn) {
      generateBtn.addEventListener('click', () => {
        this.generateCommands();
      });
    }

    // Character counter
    const taskDescription = document.getElementById('task-description');
    if (taskDescription) {
      taskDescription.addEventListener('input', (e) => {
        this.updateCharCounter(e.target.value.length);
      });
    }

    // Settings
    document.querySelectorAll('#settings-view input, #settings-view select').forEach(element => {
      element.addEventListener('change', (e) => {
        const value = e.target.type === 'checkbox' ? e.target.checked : e.target.value;
        this.updateSetting(e.target.id, value);
      });
    });

    // Header buttons
    const refreshBtn = document.getElementById('refresh-all');
    const minimizeBtn = document.getElementById('minimize-btn');

    if (refreshBtn) {
      refreshBtn.addEventListener('click', () => {
        this.refreshAll();
      });
    }

    if (minimizeBtn) {
      minimizeBtn.addEventListener('click', () => {
        this.minimizeWindow();
      });
    }

    // Copy and export buttons
    const copyBtn = document.getElementById('copy-commands-btn');
    const exportBtn = document.getElementById('export-commands-btn');

    if (copyBtn) {
      copyBtn.addEventListener('click', () => {
        this.copyCommands();
      });
    }

    if (exportBtn) {
      exportBtn.addEventListener('click', () => {
        this.exportCommands();
      });
    }
  }

  // Navigation
  switchView(viewName) {
    // Update navigation
    document.querySelectorAll('.nav-item').forEach(item => {
      item.classList.remove('active');
    });
    
    document.querySelector(`[data-view="${viewName}"]`).classList.add('active');

    // Update views
    document.querySelectorAll('.view').forEach(view => {
      view.classList.remove('active');
    });
    
    document.getElementById(`${viewName}-view`).classList.add('active');

    // Update title and breadcrumb
    const viewTitle = document.getElementById('view-title');
    const breadcrumb = document.getElementById('breadcrumb-path');
    
    const titles = {
      dashboard: 'Dashboard',
      commands: 'AI Commands',
      office: 'Office Suite',
      system: 'System Tools',
      web: 'Web Automation',
      testing: 'Testing & Validation',
      settings: 'Settings'
    };

    if (viewTitle) viewTitle.textContent = titles[viewName] || viewName;
    if (breadcrumb) breadcrumb.textContent = `Home > ${titles[viewName] || viewName}`;

    this.currentView = viewName;
  }

  // Working Directory
  async setWorkingDirectory(path) {
    this.currentWorkingDir = path;
    this.addActivity('📁', 'Directory Changed', `Working directory set to ${path}`);
  }

  // Character Counter
  updateCharCounter(length) {
    const charCount = document.getElementById('char-count');
    if (charCount) {
      charCount.textContent = length;
      charCount.style.color = length > 1800 ? '#dc3545' : '#6c757d';
    }
  }

  // Command Generation
  async generateCommands() {
    const description = document.getElementById('task-description').value.trim();
    
    if (!description) {
      this.showToast('warning', 'Please enter a task description');
      return;
    }

    if (this.backendStatus === 'offline') {
      this.showToast('error', 'Backend is offline. Please start the backend server.');
      return;
    }

    const generateBtn = document.getElementById('generate-commands-btn');
    const originalText = generateBtn.querySelector('.btn-text').textContent;
    
    // Show loading state
    generateBtn.disabled = true;
    generateBtn.querySelector('.btn-text').textContent = 'Generating...';
    generateBtn.querySelector('.btn-icon').textContent = '⏳';

    this.showLoading('Generating commands with AI...');

    try {
      const complexity = document.getElementById('task-complexity').value;
      const safetyLevel = document.getElementById('safety-level').value;
      const outputFormat = document.getElementById('output-format').value;
      const dryRun = document.getElementById('dry-run').checked;
      const createBackup = document.getElementById('create-backup').checked;
      const verboseOutput = document.getElementById('verbose-output').checked;

      const response = await this.apiCall('POST', '/api/commands', {
        description,
        working_dir: this.currentWorkingDir,
        complexity,
        safety_level: safetyLevel,
        output_format: outputFormat,
        dry_run: dryRun,
        create_backup: createBackup,
        verbose_output: verboseOutput,
        ai_model: this.settings.aiModel,
        ai_temperature: this.settings.aiTemperature
      });

      const result = response.data.data || response.data;
      this.currentCommands = result.commands || [];
      
      this.displayCommands(result);
      this.updateCommandActions();
      
      // Update stats
      this.stats.commandsExecuted++;
      this.storage.set('stats', this.stats);
      this.updateStatsDisplay();
      
      this.addActivity('🤖', 'Commands Generated', `Generated ${this.currentCommands.length} commands`);
      this.showToast('success', 'Commands generated successfully!');

    } catch (error) {
      this.log('error', 'Failed to generate commands:', error);
      this.showToast('error', `Failed to generate commands: ${error.message}`);
      this.displayError(error);
    } finally {
      // Reset button state
      generateBtn.disabled = false;
      generateBtn.querySelector('.btn-text').textContent = originalText;
      generateBtn.querySelector('.btn-icon').textContent = '✨';
      this.hideLoading();
    }
  }

  displayCommands(result) {
    const container = document.getElementById('commands-container');
    if (!container) return;

    if (!result.commands || result.commands.length === 0) {
      container.innerHTML = `
        <div class="empty-state">
          <div class="empty-icon">❌</div>
          <h3>No Commands Generated</h3>
          <p>Please try a different description or adjust your settings.</p>
        </div>
      `;
      return;
    }

    const commandsHtml = `
      <div class="command-card">
        <div class="command-header">
          <span class="command-title">Generated Commands (${result.commands.length})</span>
          <div class="command-actions">
            ${result.dry_run ? `
              <button class="btn btn-primary" onclick="app.executeCommands()">
                <span class="btn-icon">▶️</span>
                <span class="btn-text">Execute Commands</span>
              </button>
            ` : ''}
          </div>
        </div>
        
        <div class="command-list">
          ${result.commands.map(cmd => `
            <div class="command-item ${cmd.startsWith('#') ? 'comment' : ''}">
              ${this.escapeHtml(cmd)}
            </div>
          `).join('')}
        </div>
        
        ${result.explanation ? `
          <div class="command-explanation">
            <strong>📝 Explanation:</strong><br>
            ${this.escapeHtml(result.explanation)}
          </div>
        ` : ''}
      </div>
    `;

    container.innerHTML = commandsHtml;
  }

  updateCommandActions() {
    const copyBtn = document.getElementById('copy-commands-btn');
    const exportBtn = document.getElementById('export-commands-btn');
    
    if (copyBtn) copyBtn.disabled = this.currentCommands.length === 0;
    if (exportBtn) exportBtn.disabled = this.currentCommands.length === 0;
  }

  async copyCommands() {
    if (this.currentCommands.length === 0) {
      this.showToast('warning', 'No commands to copy');
      return;
    }

    try {
      const commandsText = this.currentCommands.join('\n');
      await this.copyToClipboard(commandsText);
      this.showToast('success', 'Commands copied to clipboard');
    } catch (error) {
      this.log('error', 'Failed to copy commands:', error);
      this.showToast('error', 'Failed to copy commands');
    }
  }

  async exportCommands() {
    if (this.currentCommands.length === 0) {
      this.showToast('warning', 'No commands to export');
      return;
    }

    try {
      const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
      const filename = `nlp-automation-commands-${timestamp}.sh`;
      const commandsText = this.currentCommands.join('\n');
      
      this.downloadFile(filename, commandsText, 'text/plain');
      this.showToast('success', 'Commands exported successfully');
    } catch (error) {
      this.log('error', 'Failed to export commands:', error);
      this.showToast('error', 'Failed to export commands');
    }
  }

  displayError(error) {
    const container = document.getElementById('commands-container');
    if (!container) return;

    container.innerHTML = `
      <div class="command-card">
        <div class="command-header">
          <span class="command-title">❌ Error</span>
        </div>
        <div class="command-explanation" style="background: var(--danger-light); border-color: var(--danger-color);">
          <strong>Failed to generate commands:</strong><br>
          ${this.escapeHtml(error.response?.data?.message || error.message)}<br>
          <small>💡 Try: Check your internet connection or restart the backend</small>
        </div>
      </div>
    `;
  }

  // Stats and Activities Management
  async loadInitialData() {
    this.stats = this.storage.get('stats', this.stats);
    this.updateStatsDisplay();

    this.activities = this.storage.get('activities', []);
    this.updateActivitiesDisplay();

    const savedTemplates = this.storage.get('templates', {});
    this.templates = new Map(Object.entries(savedTemplates));
  }

  updateStatsDisplay() {
    const elements = {
      commandsExecuted: document.getElementById('commands-executed'),
      filesProcessed: document.getElementById('files-processed'),
      timeSaved: document.getElementById('time-saved')
    };

    if (elements.commandsExecuted) {
      elements.commandsExecuted.textContent = this.stats.commandsExecuted.toLocaleString();
    }

    if (elements.filesProcessed) {
      elements.filesProcessed.textContent = this.stats.filesProcessed.toLocaleString();
    }

    if (elements.timeSaved) {
      elements.timeSaved.textContent = `${this.stats.timeSaved}h`;
    }
  }

  updateActivitiesDisplay() {
    const container = document.getElementById('recent-activities');
    if (!container) return;

    const recentActivities = this.activities.slice(0, 5);
    
    if (recentActivities.length === 0) {
      container.innerHTML = `
        <div class="activity-item">
          <span class="activity-icon">🎯</span>
          <div class="activity-content">
            <div class="activity-title">Welcome to NLP-Automation!</div>
            <div class="activity-desc">Your AI-powered automation assistant is ready</div>
          </div>
          <div class="activity-time">Just now</div>
        </div>
      `;
      return;
    }

    container.innerHTML = recentActivities.map(activity => `
      <div class="activity-item">
        <span class="activity-icon">${activity.icon}</span>
        <div class="activity-content">
          <div class="activity-title">${activity.title}</div>
          <div class="activity-desc">${activity.description}</div>
        </div>
        <div class="activity-time">${this.formatTime(activity.timestamp)}</div>
      </div>
    `).join('');
  }

  addActivity(icon, title, description) {
    const activity = {
      icon,
      title,
      description,
      timestamp: new Date().toISOString()
    };

    this.activities.unshift(activity);
    this.activities = this.activities.slice(0, 20); // Keep only last 20 activities
    
    this.storage.set('activities', this.activities);
    this.updateActivitiesDisplay();
  }

  formatTime(timestamp) {
    const date = new Date(timestamp);
    const now = new Date();
    const diffInMinutes = Math.floor((now - date) / (1000 * 60));

    if (diffInMinutes === 0) return 'Just now';
    if (diffInMinutes === 1) return '1 minute ago';
    if (diffInMinutes < 60) return `${diffInMinutes} minutes ago`;
    if (diffInMinutes < 1440) return `${Math.floor(diffInMinutes / 60)} hours ago`;
    return date.toLocaleDateString();
  }

  // Settings Management
  updateSetting(settingId, value) {
    const settingMap = {
      'ai-model': 'aiModel',
      'ai-temperature': 'aiTemperature',
      'theme': 'theme',
      'font-size': 'fontSize',
      'require-confirmation': 'requireConfirmation',
      'auto-backup': 'autoBackup'
    };

    const settingKey = settingMap[settingId];
    if (settingKey) {
      this.settings[settingKey] = value;
      this.saveSettings();
      this.showToast('success', 'Setting updated');
    }
  }

  // System Operations
  async refreshAll() {
    this.showLoading('Refreshing all data...');
    
    try {
      await Promise.all([
        this.checkBackendStatus(),
        this.updateSystemInfo()
      ]);
      
      this.showToast('success', 'All data refreshed successfully');
      this.addActivity('🔄', 'Data Refreshed', 'All system data refreshed');
      
    } catch (error) {
      this.log('error', 'Failed to refresh data:', error);
      this.showToast('error', 'Failed to refresh some data');
    } finally {
      this.hideLoading();
    }
  }

  // UI Helper Methods
  showLoading(message = 'Loading...') {
    const overlay = document.getElementById('loading-overlay');
    const text = document.querySelector('.loading-text');
    
    if (overlay) {
      overlay.classList.remove('hidden');
    }
    
    if (text) {
      text.textContent = message;
    }
  }

  hideLoading() {
    const overlay = document.getElementById('loading-overlay');
    if (overlay) {
      overlay.classList.add('hidden');
    }
  }

  showToast(type, message, duration = 3000) {
    const container = document.getElementById('toast-container');
    if (!container) return;

    const toast = document.createElement('div');
    toast.className = `toast ${type}`;
    toast.innerHTML = `
      <div class="toast-content">
        <span class="toast-icon">${this.getToastIcon(type)}</span>
        <span class="toast-message">${message}</span>
      </div>
    `;

    container.appendChild(toast);

    // Auto remove after duration
    setTimeout(() => {
      toast.remove();
    }, duration);
  }

  getToastIcon(type) {
    const icons = {
      success: '✅',
      error: '❌',
      warning: '⚠️',
      info: 'ℹ️'
    };
    return icons[type] || 'ℹ️';
  }

  escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
  }

  log(level, message, ...args) {
    const timestamp = new Date().toISOString();
    const logMessage = `[${timestamp}] [${level.toUpperCase()}] ${message}`;
    
    if (level === 'error') {
      console.error(logMessage, ...args);
    } else if (level === 'warning') {
      console.warn(logMessage, ...args);
    } else {
      console.log(logMessage, ...args);
    }
  }
}

// Initialize the application when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
  window.app = new NLPAutomationWebApp();
});

// Global functions for onclick handlers
window.switchView = (view) => {
  if (window.app) {
    window.app.switchView(view);
  }
};

// Export for module systems
if (typeof module !== 'undefined' && module.exports) {
  module.exports = NLPAutomationWebApp;
}
