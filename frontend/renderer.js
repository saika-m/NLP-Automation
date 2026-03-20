// Tashi Frontend - Modern Renderer
// Version 2.0.0

// Import required modules for Electron
const { ipcRenderer, shell } = require('electron');
const Store = require('electron-store');
const axios = require('axios');
const hljs = require('highlight.js');
const { marked } = require('marked');

// Global application state
class TashiApp {
  constructor() {
    this.store = new Store();
    this.currentView = 'dashboard';
    this.backendUrl = 'http://localhost:8080';
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
    
    // Initialize
    this.init();
  }

  async init() {
    await this.loadSettings();
    await this.setupEventListeners();
    await this.initializeSystemInfo();
    await this.checkBackendStatus();
    await this.loadInitialData();
    
    // Set up periodic updates
    setInterval(() => this.updateSystemInfo(), 5000);
    setInterval(() => this.checkBackendStatus(), 10000);
    setInterval(() => this.updateClock(), 1000);
    
    this.log('info', 'Tashi application initialized successfully');
  }

  async loadSettings() {
    const savedSettings = this.store.get('settings', {});
    this.settings = { ...this.settings, ...savedSettings };
    this.applySettings();
  }

  async saveSettings() {
    this.store.set('settings', this.settings);
    this.applySettings();
  }

  applySettings() {
    // Apply theme
    document.documentElement.setAttribute('data-theme', this.settings.theme);
    
    // Apply font size
    document.documentElement.setAttribute('data-font-size', this.settings.fontSize);
    
    // Update UI elements
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

  async setupEventListeners() {
    // Navigation
    document.querySelectorAll('[data-view]').forEach(navItem => {
      navItem.addEventListener('click', (e) => {
        e.preventDefault();
        const view = e.currentTarget.dataset.view;
        console.log('Navigation clicked:', view);
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
      selectDirBtn.addEventListener('click', () => {
        this.selectDirectory();
      });
    }

    // Directory suggestions
    document.querySelectorAll('.dir-suggestion').forEach(btn => {
      btn.addEventListener('click', (e) => {
        const suggestion = e.currentTarget.textContent.toLowerCase();
        this.setWorkingDirectoryByType(suggestion);
      });
    });

    // Command generation
    const generateBtn = document.getElementById('generate-commands-btn');
    const taskDescription = document.getElementById('task-description');
    const saveTemplateBtn = document.getElementById('save-template-btn');

    if (generateBtn) {
      generateBtn.addEventListener('click', () => {
        this.generateCommands();
      });
    }

    if (taskDescription) {
      taskDescription.addEventListener('input', (e) => {
        this.updateCharCounter(e.target.value.length);
      });
    }

    if (saveTemplateBtn) {
      saveTemplateBtn.addEventListener('click', () => {
        this.saveTemplate();
      });
    }

    // Office form
    const officeForm = document.getElementById('office-form');
    if (officeForm) {
      officeForm.addEventListener('submit', (e) => {
        e.preventDefault();
        this.executeOfficeAction();
      });
    }

    // Settings
    document.querySelectorAll('#settings-view input, #settings-view select').forEach(element => {
      element.addEventListener('change', (e) => {
        this.updateSetting(e.target.id, e.target.type === 'checkbox' ? e.target.checked : e.target.value);
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
        ipcRenderer.send('minimize-window');
      });
    }

    // Result tabs
    document.querySelectorAll('.result-tab').forEach(tab => {
      tab.addEventListener('click', (e) => {
        this.switchResultTab(e.target.dataset.tab);
      });
    });

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

  async initializeSystemInfo() {
    try {
      const os = require('os');
      this.systemInfo = {
        platform: os.platform(),
        arch: os.arch(),
        hostname: os.hostname(),
        totalMemory: os.totalmem(),
        freeMemory: os.freemem(),
        cpus: os.cpus(),
        uptime: os.uptime()
      };

      // Set home directory as default
      this.currentWorkingDir = os.homedir();
      const workingDirInput = document.getElementById('working-dir');
      if (workingDirInput) {
        workingDirInput.value = this.currentWorkingDir;
      }

      this.updateSystemDisplay();
    } catch (error) {
      this.log('error', 'Failed to initialize system info:', error);
    }
  }

  updateSystemDisplay() {
    const elements = {
      platformInfo: document.getElementById('platform-info'),
      memoryInfo: document.getElementById('memory-info'),
      systemPlatform: document.getElementById('system-platform'),
      systemArch: document.getElementById('system-arch'),
      systemMemory: document.getElementById('system-memory'),
      systemStorage: document.getElementById('system-storage')
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

    // Update performance metrics
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

  async updateSystemInfo() {
    try {
      const os = require('os');
      this.systemInfo.freeMemory = os.freemem();
      this.systemInfo.uptime = os.uptime();
      this.updateSystemDisplay();
    } catch (error) {
      this.log('error', 'Failed to update system info:', error);
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

  async checkBackendStatus() {
    try {
      const response = await axios.get(`${this.backendUrl}/api/health`, {
        timeout: 3000
      });
      
      if (response.status === 200) {
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
      const response = await axios.get(`${this.backendUrl}/api/ai/status`, {
        timeout: 3000
      });
      
      if (response.status === 200 && response.data.status === 'online') {
        this.setAIStatus('online');
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

  async loadInitialData() {
    // Load saved stats
    this.stats = this.store.get('stats', this.stats);
    this.updateStatsDisplay();

    // Load saved activities
    this.activities = this.store.get('activities', []);
    this.updateActivitiesDisplay();

    // Load saved templates
    const savedTemplates = this.store.get('templates', {});
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
            <div class="activity-title">Welcome to Tashi!</div>
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
    
    this.store.set('activities', this.activities);
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
    this.onViewChange(viewName);
  }

  async onViewChange(viewName) {
    switch (viewName) {
      case 'office':
        await this.checkOfficeCapabilities();
        break;
      case 'system':
        await this.updateSystemInfo();
        break;
      case 'testing':
        await this.loadTestSuites();
        break;
    }
  }

  async setWorkingDirectory(path) {
    this.currentWorkingDir = path;
    await this.updateDirectoryInfo(path);
    this.addActivity('📁', 'Directory Changed', `Working directory set to ${path}`);
  }

  async setWorkingDirectoryByType(type) {
    const os = require('os');
    let path;

    switch (type) {
      case '🖥️ desktop':
        path = require('path').join(os.homedir(), 'Desktop');
        break;
      case '📥 downloads':
        path = require('path').join(os.homedir(), 'Downloads');
        break;
      case '📄 documents':
        path = require('path').join(os.homedir(), 'Documents');
        break;
      default:
        path = os.homedir();
    }

    const workingDirInput = document.getElementById('working-dir');
    if (workingDirInput) {
      workingDirInput.value = path;
    }

    await this.setWorkingDirectory(path);
  }

  async selectDirectory() {
    try {
      const result = await ipcRenderer.invoke('select-directory');
      if (result && !result.canceled && result.filePaths.length > 0) {
        const path = result.filePaths[0];
        const workingDirInput = document.getElementById('working-dir');
        if (workingDirInput) {
          workingDirInput.value = path;
        }
        await this.setWorkingDirectory(path);
      }
    } catch (error) {
      this.log('error', 'Failed to select directory:', error);
      this.showToast('error', 'Failed to select directory');
    }
  }

  async updateDirectoryInfo(path) {
    const directoryInfo = document.getElementById('directory-info');
    if (!directoryInfo) return;

    try {
      const fs = require('fs');
      const stats = await fs.promises.stat(path);
      
      if (stats.isDirectory()) {
        const files = await fs.promises.readdir(path);
        const fileCount = files.length;
        
        directoryInfo.innerHTML = `
          <strong>📁 ${path}</strong><br>
          <small>📊 ${fileCount} items | 💡 Ready for automation</small>
        `;
        directoryInfo.classList.remove('hidden');
      } else {
        directoryInfo.innerHTML = '<strong>⚠️ Not a valid directory</strong>';
        directoryInfo.classList.remove('hidden');
      }
    } catch (error) {
      directoryInfo.innerHTML = '<strong>⚠️ Directory not accessible</strong>';
      directoryInfo.classList.remove('hidden');
    }
  }

  updateCharCounter(length) {
    const charCount = document.getElementById('char-count');
    if (charCount) {
      charCount.textContent = length;
      charCount.style.color = length > 1800 ? '#dc3545' : '#6c757d';
    }
  }

  async generateCommands() {
    const description = document.getElementById('task-description').value.trim();
    const complexity = document.getElementById('task-complexity').value;
    const safetyLevel = document.getElementById('safety-level').value;
    const outputFormat = document.getElementById('output-format').value;
    const dryRun = document.getElementById('dry-run').checked;
    const createBackup = document.getElementById('create-backup').checked;
    const verboseOutput = document.getElementById('verbose-output').checked;

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
      const response = await axios.post(`${this.backendUrl}/api/commands`, {
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
      this.store.set('stats', this.stats);
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
        
        ${result.safety_warnings && result.safety_warnings.length > 0 ? `
          <div class="safety-warnings">
            <h4>⚠️ Safety Warnings</h4>
            <ul>
              ${result.safety_warnings.map(warning => `<li>${this.escapeHtml(warning)}</li>`).join('')}
            </ul>
          </div>
        ` : ''}
        
        ${result.estimated_time ? `
          <div class="command-explanation">
            <strong>⏱️ Estimated Time:</strong> ${result.estimated_time}
          </div>
        ` : ''}
      </div>
    `;

    container.innerHTML = commandsHtml;
  }

  updateCommandActions() {
    const copyBtn = document.getElementById('copy-commands-btn');
    const exportBtn = document.getElementById('export-commands-btn');
    
    if (copyBtn) {
      copyBtn.disabled = this.currentCommands.length === 0;
    }
    
    if (exportBtn) {
      exportBtn.disabled = this.currentCommands.length === 0;
    }
  }

  async executeCommands() {
    if (this.currentCommands.length === 0) {
      this.showToast('warning', 'No commands to execute');
      return;
    }

    if (this.settings.requireConfirmation) {
      const confirmed = await this.showConfirmation(
        'Execute Commands',
        `Are you sure you want to execute ${this.currentCommands.length} commands?`,
        'This action cannot be undone.'
      );
      
      if (!confirmed) return;
    }

    this.showLoading('Executing commands...');
    
    const executionSection = document.getElementById('execution-section');
    const resultStatus = document.getElementById('result-status');
    const commandOutput = document.getElementById('command-output');
    
    if (executionSection) {
      executionSection.classList.remove('hidden');
    }
    
    if (resultStatus) {
      resultStatus.textContent = 'Running';
      resultStatus.className = 'result-status running';
    }
    
    if (commandOutput) {
      commandOutput.textContent = 'Executing commands...\n';
    }

    try {
      const response = await axios.post(`${this.backendUrl}/api/commands/execute`, {
        commands: this.currentCommands,
        working_dir: this.currentWorkingDir,
        create_backup: this.settings.autoBackup
      });

      const result = response.data.data || response.data;
      this.displayExecutionResults(result);
      
      this.addActivity('⚡', 'Commands Executed', `Successfully executed ${this.currentCommands.length} commands`);
      this.showToast('success', 'Commands executed successfully!');

    } catch (error) {
      this.log('error', 'Failed to execute commands:', error);
      this.showToast('error', `Failed to execute commands: ${error.message}`);
      this.displayExecutionError(error);
    } finally {
      this.hideLoading();
    }
  }

  displayExecutionResults(result) {
    const resultStatus = document.getElementById('result-status');
    const executionTime = document.getElementById('execution-time');
    const commandOutput = document.getElementById('command-output');
    const commandErrors = document.getElementById('command-errors');
    const executionSummary = document.getElementById('execution-summary');

    if (resultStatus) {
      resultStatus.textContent = result.success ? 'Success' : 'Error';
      resultStatus.className = `result-status ${result.success ? 'success' : 'error'}`;
    }

    if (executionTime) {
      executionTime.textContent = result.execution_time || '';
    }

    if (commandOutput) {
      commandOutput.textContent = result.output || 'No output generated.';
    }

    if (commandErrors) {
      commandErrors.textContent = result.errors || 'No errors occurred.';
    }

    if (executionSummary) {
      executionSummary.innerHTML = `
        <div class="detail-item">
          <span class="detail-label">Commands Executed:</span>
          <span class="detail-value">${result.commands_executed || 0}</span>
        </div>
        <div class="detail-item">
          <span class="detail-label">Files Modified:</span>
          <span class="detail-value">${result.files_modified || 0}</span>
        </div>
        <div class="detail-item">
          <span class="detail-label">Exit Code:</span>
          <span class="detail-value">${result.exit_code || 0}</span>
        </div>
        ${result.backup_created ? `
          <div class="detail-item">
            <span class="detail-label">Backup Created:</span>
            <span class="detail-value">${result.backup_path}</span>
          </div>
        ` : ''}
      `;
    }
  }

  displayExecutionError(error) {
    const resultStatus = document.getElementById('result-status');
    const commandOutput = document.getElementById('command-output');
    const commandErrors = document.getElementById('command-errors');

    if (resultStatus) {
      resultStatus.textContent = 'Error';
      resultStatus.className = 'result-status error';
    }

    if (commandOutput) {
      commandOutput.textContent = 'Execution failed. See errors tab for details.';
    }

    if (commandErrors) {
      commandErrors.textContent = error.response?.data?.message || error.message || 'Unknown error occurred';
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

  switchResultTab(tabName) {
    // Update tabs
    document.querySelectorAll('.result-tab').forEach(tab => {
      tab.classList.remove('active');
    });
    document.querySelector(`[data-tab="${tabName}"]`).classList.add('active');

    // Update panels
    document.querySelectorAll('.result-panel').forEach(panel => {
      panel.classList.remove('active');
    });
    document.getElementById(`${tabName}-panel`).classList.add('active');
  }

  async copyCommands() {
    if (this.currentCommands.length === 0) {
      this.showToast('warning', 'No commands to copy');
      return;
    }

    try {
      const commandsText = this.currentCommands.join('\n');
      await navigator.clipboard.writeText(commandsText);
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
      const filename = `tashi-commands-${timestamp}.sh`;
      const commandsText = this.currentCommands.join('\n');
      
      const blob = new Blob([commandsText], { type: 'text/plain' });
      const url = URL.createObjectURL(blob);
      
      const a = document.createElement('a');
      a.href = url;
      a.download = filename;
      a.click();
      
      URL.revokeObjectURL(url);
      this.showToast('success', 'Commands exported successfully');
    } catch (error) {
      this.log('error', 'Failed to export commands:', error);
      this.showToast('error', 'Failed to export commands');
    }
  }

  async saveTemplate() {
    const description = document.getElementById('task-description').value.trim();
    const complexity = document.getElementById('task-complexity').value;
    const safetyLevel = document.getElementById('safety-level').value;
    const outputFormat = document.getElementById('output-format').value;

    if (!description) {
      this.showToast('warning', 'Please enter a task description first');
      return;
    }

    const templateName = prompt('Enter a name for this template:');
    if (!templateName) return;

    const template = {
      name: templateName,
      description,
      complexity,
      safetyLevel,
      outputFormat,
      timestamp: new Date().toISOString()
    };

    this.templates.set(templateName, template);
    this.store.set('templates', Object.fromEntries(this.templates));
    
    this.showToast('success', 'Template saved successfully');
    this.addActivity('💾', 'Template Saved', `Created template: ${templateName}`);
  }

  useTemplate(templateType) {
    const templates = {
      'file-management': {
        description: 'Organize and manage files in my working directory by type, date, and importance',
        complexity: 'moderate',
        safetyLevel: 'high',
        outputFormat: 'detailed'
      },
      'data-processing': {
        description: 'Process and analyze data files, extract information, and generate reports',
        complexity: 'complex',
        safetyLevel: 'medium',
        outputFormat: 'detailed'
      },
      'system-maintenance': {
        description: 'Perform system maintenance tasks like cleaning temporary files and optimizing performance',
        complexity: 'moderate',
        safetyLevel: 'high',
        outputFormat: 'detailed'
      },
      'development': {
        description: 'Set up development environment, manage dependencies, and organize code structure',
        complexity: 'complex',
        safetyLevel: 'medium',
        outputFormat: 'interactive'
      }
    };

    const template = templates[templateType];
    if (!template) return;

    const taskDescription = document.getElementById('task-description');
    const taskComplexity = document.getElementById('task-complexity');
    const safetyLevel = document.getElementById('safety-level');
    const outputFormat = document.getElementById('output-format');

    if (taskDescription) taskDescription.value = template.description;
    if (taskComplexity) taskComplexity.value = template.complexity;
    if (safetyLevel) safetyLevel.value = template.safetyLevel;
    if (outputFormat) outputFormat.value = template.outputFormat;

    this.updateCharCounter(template.description.length);
    this.showToast('info', `Template "${templateType}" loaded`);
  }

  setExample(exampleType) {
    const examples = {
      'file-backup': 'Create a timestamped backup folder on my desktop and copy all important documents, photos, and configuration files to it',
      'photo-organization': 'Organize all photos in my Pictures folder by year and month, creating a folder structure like 2024/01-January, 2024/02-February, etc.',
      'cleanup-downloads': 'Clean up my Downloads folder by removing files older than 30 days and organizing remaining files by type (documents, images, videos, etc.)',
      'project-setup': 'Create a new project directory structure with folders for source code, documentation, tests, and build output, including basic configuration files'
    };

    const example = examples[exampleType];
    if (!example) return;

    const taskDescription = document.getElementById('task-description');
    if (taskDescription) {
      taskDescription.value = example;
      this.updateCharCounter(example.length);
      
      // Add visual feedback
      taskDescription.style.backgroundColor = '#e7f3ff';
      setTimeout(() => {
        taskDescription.style.backgroundColor = '';
      }, 500);
    }
  }

  async executeOfficeAction() {
    const app = document.getElementById('office-app').value;
    const filePath = document.getElementById('office-file').value;
    const description = document.getElementById('office-description').value.trim();

    if (!app) {
      this.showToast('warning', 'Please select an application');
      return;
    }

    if (!description) {
      this.showToast('warning', 'Please enter an action description');
      return;
    }

    if (this.backendStatus === 'offline') {
      this.showToast('error', 'Backend is offline. Please start the backend server.');
      return;
    }

    this.showLoading('Executing office action...');

    try {
      const response = await axios.post(`${this.backendUrl}/api/office`, {
        application: app,
        file_path: filePath,
        description: description,
        working_dir: this.currentWorkingDir
      });

      const result = response.data.data || response.data;
      this.showToast('success', 'Office action executed successfully!');
      this.addActivity('📊', 'Office Action', `Executed ${app} action: ${description}`);
      
    } catch (error) {
      this.log('error', 'Failed to execute office action:', error);
      this.showToast('error', `Failed to execute office action: ${error.message}`);
    } finally {
      this.hideLoading();
    }
  }

  async checkOfficeCapabilities() {
    try {
      const response = await axios.get(`${this.backendUrl}/api/office/capabilities`);
      const capabilities = response.data.data || response.data;
      
      const statusElements = {
        excel: document.getElementById('excel-status'),
        word: document.getElementById('word-status'),
        powerpoint: document.getElementById('powerpoint-status'),
        outlook: document.getElementById('outlook-status')
      };

      Object.entries(capabilities).forEach(([app, available]) => {
        const element = statusElements[app];
        if (element) {
          element.textContent = available ? 'Available' : 'Not Available';
          element.className = `capability-status ${available ? 'available' : 'unavailable'}`;
        }
      });
      
    } catch (error) {
      this.log('error', 'Failed to check office capabilities:', error);
      Object.values(statusElements).forEach(element => {
        if (element) {
          element.textContent = 'Error checking';
          element.className = 'capability-status unavailable';
        }
      });
    }
  }

  async loadTestSuites() {
    // Implementation for loading test suites
    this.addActivity('🧪', 'Tests Loaded', 'Test suites loaded successfully');
  }

  async runTest(testType) {
    const resultElement = document.getElementById(`${testType}-test-results`);
    if (!resultElement) return;

    resultElement.innerHTML = 'Running tests...';
    this.showLoading(`Running ${testType} tests...`);

    try {
      const response = await axios.post(`${this.backendUrl}/api/test`, {
        type: testType
      });

      const result = response.data.data || response.data;
      resultElement.innerHTML = `
        <div class="test-result">
          <div class="test-summary">
            <span class="test-passed">${result.passed} passed</span>
            <span class="test-failed">${result.failed} failed</span>
            <span class="test-duration">${result.duration}ms</span>
          </div>
          <pre class="test-output">${result.output}</pre>
        </div>
      `;
      
      this.addActivity('🧪', 'Tests Run', `${testType} tests: ${result.passed} passed, ${result.failed} failed`);
      
    } catch (error) {
      this.log('error', `Failed to run ${testType} tests:`, error);
      resultElement.innerHTML = `<div class="test-error">Failed to run tests: ${error.message}</div>`;
    } finally {
      this.hideLoading();
    }
  }

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

  async refreshAll() {
    this.showLoading('Refreshing all data...');
    
    try {
      await Promise.all([
        this.checkBackendStatus(),
        this.updateSystemInfo(),
        this.checkOfficeCapabilities()
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

  async showConfirmation(title, message, details = '') {
    return new Promise((resolve) => {
      const modal = document.getElementById('confirmation-modal');
      const titleElement = document.getElementById('confirmation-title');
      const messageElement = document.getElementById('confirmation-message');
      const detailsElement = document.getElementById('confirmation-details');
      const yesBtn = document.getElementById('confirm-yes');
      const noBtn = document.getElementById('confirm-no');

      if (titleElement) titleElement.textContent = title;
      if (messageElement) messageElement.textContent = message;
      if (detailsElement) detailsElement.textContent = details;

      if (modal) {
        modal.classList.remove('hidden');
      }

      const handleYes = () => {
        modal.classList.add('hidden');
        cleanup();
        resolve(true);
      };

      const handleNo = () => {
        modal.classList.add('hidden');
        cleanup();
        resolve(false);
      };

      const cleanup = () => {
        yesBtn.removeEventListener('click', handleYes);
        noBtn.removeEventListener('click', handleNo);
      };

      yesBtn.addEventListener('click', handleYes);
      noBtn.addEventListener('click', handleNo);
    });
  }

  closeModal(modalId) {
    const modal = document.getElementById(modalId);
    if (modal) {
      modal.classList.add('hidden');
    }
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
  window.app = new TashiApp();
});

// Global functions for onclick handlers
window.switchView = (view) => {
  if (window.app) {
    window.app.switchView(view);
  }
};

window.setWorkingDir = (type) => {
  if (window.app) {
    window.app.setWorkingDirectoryByType(type);
  }
};

window.useTemplate = (template) => {
  if (window.app) {
    window.app.useTemplate(template);
  }
};

window.setExample = (example) => {
  if (window.app) {
    window.app.setExample(example);
  }
};

window.runTest = (testType) => {
  if (window.app) {
    window.app.runTest(testType);
  }
};

window.closeModal = (modalId) => {
  if (window.app) {
    window.app.closeModal(modalId);
  }
};

// Export for testing
if (typeof module !== 'undefined' && module.exports) {
  module.exports = TashiApp;
}
