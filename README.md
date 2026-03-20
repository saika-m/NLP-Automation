# NLP-Automation - Personal Automation Command Executor

NLP-Automation is a user-friendly desktop application that helps you automate tasks by generating and executing commands based on natural language descriptions. It uses GPT-4o-mini for cost-effective AI-powered command generation and provides a clean, intuitive interface designed for both technical and non-technical users.

## 🚀 Key Features

- **No-Code Interface**: Simple, user-friendly UI with example tasks and smart suggestions
- **Intelligent Path Inference**: Automatically understands common locations like "desktop", "downloads", "documents"
- **Natural Language Processing**: Describe tasks in plain English with smart context understanding
- **Command Generation**: Uses GPT-4o-mini for cost-effective AI-powered command generation
- **Direct Command Execution**: Execute commands directly from the interface with real-time feedback
- **Safety Features**: Dry-run mode, safety warnings, and command validation
- **Cross-platform**: Works on Windows, macOS, and Linux
- **Beautiful UI**: Clean design with your specified color scheme (#FFFFFF, #202020, #1192DC)

## 📋 Prerequisites

- **Node.js** (v16 or higher)
- **Go** (v1.19 or higher)
- **OpenAI API Key** (for command generation)

## 🛠️ Installation & Setup

### 1. Clone the Repository
```bash
git clone <repository-url>
cd NLP-Automation
```

### 2. Install Dependencies
```bash
# Install frontend dependencies
cd frontend && npm install

# Install backend dependencies (Go modules)
cd ../backend && go mod download
```

### 3. Setup Environment
```bash
# Copy and configure environment file
cp backend/.env.example backend/.env
# Edit backend/.env with your OpenAI API key
```

## 🏃‍♂️ Running the Application

### Method 1: Docker (Recommended)
```bash
# Start the entire stack
docker-compose up -d

# View logs
docker-compose logs -f

# Stop everything
docker-compose down
```

### Method 2: Local Development
```bash
# Terminal 1: Start backend
cd backend && go run .

# Terminal 2: Start frontend (Electron app)
cd frontend && npm start
```

### Method 3: Web Frontend (Alternative)
```bash
# Terminal 1: Start backend
cd backend && go run .

# Terminal 2: Start frontend as web app
cd frontend && npm run dev
```

## 🎯 How to Use

### Getting Started
1. **Launch the App**: Start the Electron application
2. **Check Backend Status**: Ensure the backend shows "Online" in the top-right corner
3. **Select Working Directory**: Choose where commands should be executed (defaults to home directory)

### Using the Interface
1. **Try Quick Examples**: Click any of the example buttons to see how it works:
   - 📁 Backup Documents
   - 🖼️ Organize Photos
   - 🗑️ Clean Downloads
   - 💻 Organize Code

2. **Or Write Your Own Tasks**: Describe what you want to accomplish in plain English:
   - "Create a backup folder on desktop and copy all my documents to it"
   - "Find all photos on desktop and organize them by date"
   - "Clean up downloads folder by removing old files"

3. **Smart Path Understanding**: No need to specify full paths - NLP-Automation understands:
   - "desktop" → ~/Desktop
   - "downloads" → ~/Downloads
   - "documents" → ~/Documents
   - "photos" or "pictures" → ~/Pictures

### Command Execution
1. **Dry Run Mode** (default): Preview commands before execution
2. **Generate Commands**: Click to see what NLP-Automation will do
3. **Review & Execute**: Check the commands and safety warnings, then execute if satisfied

## 🌟 Example Tasks

### File Management
- "Create a backup folder and copy all documents to it"
- "Organize photos by date in separate folders"
- "Move all code files from desktop to a new project folder"

### System Maintenance
- "Find all large files over 100MB in the current directory"
- "Clean up temporary files and empty directories"
- "Remove duplicate files from downloads folder"

### Development Tasks
- "Set up a new project directory structure"
- "Create a backup of all configuration files"
- "Organize downloads by file type"

## 🔧 Key Improvements

### Enhanced User Experience
- **Quick Example Buttons**: Get started instantly with common tasks
- **Smart Path Inference**: No need to type full paths
- **Visual Feedback**: Emojis and icons throughout the interface
- **Helpful Suggestions**: Context-aware tips and guidance

### Improved AI Intelligence
- **Better Path Understanding**: Automatically resolves common folder names
- **Context Awareness**: Understands your working directory and suggests relevant actions
- **Safety First**: Enhanced safety warnings for potentially dangerous operations

### Modern UI Design
- **Clean Color Scheme**: Uses your specified colors (#FFFFFF, #202020, #1192DC)
- **Responsive Design**: Works on different screen sizes
- **Smooth Animations**: Subtle visual feedback for better user experience

## 📊 API Endpoints

- `POST /api/commands` - Generate and execute commands
- `GET /api/health` - Backend health check
- `GET /api/system/info` - System information
- `POST /api/validate` - Command validation
- `GET /api/stats` - Usage statistics

## 🔒 Security Features

- **Dry Run Mode**: Preview commands before execution
- **Safety Warnings**: Automatic detection of potentially dangerous operations
- **Command Validation**: Analysis of commands for security risks
- **Working Directory Isolation**: Commands execute in selected directories only

## 🐛 Troubleshooting

### Backend won't start
- Check if OpenAI API key is set in `backend/.env`
- Verify Go installation: `go version`
- Check port 8080 is available: `lsof -i :8080`

### Frontend shows "Backend Offline"
- Ensure backend is running: `npm run start:backend`
- Check firewall settings
- Verify network connectivity to localhost:8080

### Commands fail to execute
- Check working directory permissions
- Verify file/directory paths exist
- Review safety warnings before execution

## 🔄 Development

### Building for Production
```bash
# Build everything
npm run build

# Build just the backend
npm run build:backend

# Build just the frontend
npm run build:frontend
```

### Platform-specific Builds
```bash
# macOS
npm run package:mac

# Windows
npm run package:win

# Linux
npm run package:linux
```

## 📁 Project Structure

```
NLP-Automation/
├── docker-compose.yml     # Docker orchestration for full stack
├── README.md              # This file
├── backend/               # Go backend server
│   ├── Dockerfile         # Backend Docker configuration
│   ├── go.mod             # Go module dependencies
│   ├── go.sum             # Go module checksums
│   ├── main.go            # Application entry point
│   ├── .env               # Environment variables
│   └── internal/
│       ├── api/           # HTTP handlers and routes
│       ├── config/        # Configuration management
│       ├── database/      # Database operations
│       └── services/      # Business logic
│           ├── ai_service.go        # Enhanced GPT-4o-mini integration
│           ├── command_service.go   # Command execution
│           └── task_service.go      # Task management
└── frontend/             # Electron desktop app
    ├── Dockerfile         # Frontend Docker configuration
    ├── package.json       # Frontend dependencies
    ├── main.js            # Electron main process
    ├── renderer.js        # Enhanced frontend JavaScript
    ├── index.html         # Improved application UI
    └── styles.css         # Updated styles with your color scheme
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## 📄 License

This project is licensed under the ISC License.

## 🆘 Support

For issues and questions:
1. Check the troubleshooting section
2. Review the API documentation
3. Create an issue in the repository

---

**Note**: This application executes commands on your system. Always review generated commands before execution and use dry-run mode when testing. The enhanced AI now provides better path inference and safer command generation.

## 🎨 UI Color Scheme

- **Primary**: #1192DC (Blue)
- **Background**: #FFFFFF (White)
- **Text**: #202020 (Dark Gray)
- **Accent**: Various shades of the primary blue for gradients and highlights
