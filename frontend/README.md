# Tashi Frontend

Tashi frontend is an Electron-based desktop application that provides a user-friendly interface for the Tashi automation command executor. It communicates with the Tashi backend API to generate and execute commands based on natural language descriptions.

## Features

- **Electron Desktop App**: Cross-platform desktop application (Windows, macOS, Linux)
- **Clean User Interface**: Modern, intuitive design with the Tashi color scheme
- **Real-time Backend Communication**: Connects to the Tashi backend API
- **Command Preview**: Shows generated commands before execution
- **Safety Warnings**: Displays safety warnings for potentially dangerous operations
- **Working Directory Selection**: Choose where commands should be executed
- **Quick Example Tasks**: Pre-built example tasks for common automation needs

## Quick Start

### Prerequisites

- **Node.js** (v16 or higher)
- **Tashi Backend** running on `http://localhost:8080`

### Installation

1. Navigate to the frontend directory:
```bash
cd frontend
```

2. Install dependencies:
```bash
npm install
```

3. Start the Electron app:
```bash
npm start
```

### Development Mode

For development with auto-reload:
```bash
npm run dev
```

### Building for Production

Build the application for your platform:

```bash
# Build for current platform
npm run build

# Build for macOS
npm run build-mac

# Build for Windows
npm run build-win

# Build for Linux
npm run build-linux
```

## Docker Setup

The frontend can also be run as a web application using Docker:

```bash
# Build Docker image
docker build -t tashi-frontend .

# Run container
docker run -p 3000:3000 tashi-frontend
```

Or use the docker-compose from the root directory:
```bash
cd ..
docker-compose up -d
```

## Project Structure

```
frontend/
├── package.json           # Dependencies and build scripts
├── Dockerfile            # Docker configuration for web deployment
├── nginx.conf            # Nginx configuration for web deployment
├── main.js               # Electron main process
├── renderer.js           # Frontend JavaScript logic
├── index.html            # Application UI
├── styles.css            # Application styles
└── README.md             # This file
```

## Key Files

### main.js
- Electron main process
- Creates and manages the application window
- Handles app lifecycle events
- Configures security settings

### renderer.js
- Frontend application logic
- Handles user interactions
- Communicates with the backend API
- Manages UI state and updates

### index.html
- Application user interface
- Contains the main UI structure
- Includes example task buttons
- Responsive design elements

### styles.css
- Application styling
- Implements the Tashi color scheme (#1192DC, #FFFFFF, #202020)
- Responsive design rules
- Modern UI components

## API Communication

The frontend communicates with the backend through these endpoints:

- `GET /api/health` - Check backend status
- `POST /api/commands` - Generate and execute commands
- `GET /api/system/info` - Get system information
- `POST /api/validate` - Validate commands

## Configuration

The frontend is configured to connect to the backend at `http://localhost:8080`. This can be modified in the `renderer.js` file if needed.

## Building and Distribution

### Electron Builder Configuration

The `package.json` includes configuration for electron-builder:

```json
{
  "build": {
    "appId": "com.tashi.desktop",
    "productName": "Tashi",
    "directories": {
      "output": "dist"
    },
    "files": [
      "main.js",
      "renderer.js", 
      "index.html",
      "styles.css"
    ]
  }
}
```

### Platform-Specific Builds

- **macOS**: Generates `.dmg` installer
- **Windows**: Generates `.exe` installer using NSIS
- **Linux**: Generates `.AppImage` package

## Development

### Running in Development Mode

```bash
npm run dev
```

This starts Electron with development tools enabled.

### Building for Production

```bash
npm run build
```

This creates optimized builds for the current platform.

### Testing

The frontend connects to the backend API. Make sure the backend is running before testing:

```bash
# Start backend first
cd ../backend && go run .

# Then start frontend
cd ../frontend && npm start
```

## Troubleshooting

### Common Issues

1. **Backend Connection Failed**: Ensure the backend is running on `http://localhost:8080`
2. **Electron Won't Start**: Make sure Node.js v16+ is installed
3. **Build Failures**: Check that all dependencies are properly installed
4. **Permission Errors**: Ensure proper file permissions for the application directory

### Debug Mode

The Electron app includes developer tools. Press `Ctrl+Shift+I` (or `Cmd+Option+I` on macOS) to open them.

### Logs

Application logs are displayed in the Electron developer console.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly with the backend
5. Submit a pull request

## License

This project is licensed under the ISC License.

## Support

For issues and questions:
1. Check the troubleshooting section
2. Ensure the backend is running properly
3. Create an issue in the main repository

---

**Note**: This is an Electron desktop application that requires the Tashi backend to be running for full functionality. Always ensure the backend is accessible before using the frontend.
