# NLP-Automation Backend

NLP-Automation backend is a Go-based API server that provides AI-powered command generation and execution capabilities for the NLP-Automation desktop application. It uses OpenAI's GPT models to convert natural language descriptions into system commands that can be safely executed.

## Features

- **AI-Powered Command Generation**: Uses OpenAI's GPT-4o-mini for cost-effective command generation
- **Multi-Platform Support**: Supports Windows, Linux, and macOS commands
- **Safety Validation**: Built-in safety checks to warn about potentially dangerous operations
- **Command Execution**: Secure command execution with proper error handling
- **RESTful API**: Complete REST API for frontend integration
- **Health Monitoring**: Built-in health checks and system information endpoints
- **Docker Support**: Full Docker containerization support

## Quick Start

### Prerequisites

- Go 1.19 or later
- OpenAI API key

### Installation

1. Navigate to the backend directory:
```bash
cd backend
```

2. Install dependencies:
```bash
go mod download
```

3. Set up environment variables:
```bash
cp .env.example .env
# Edit .env with your OpenAI API key and other settings
```

4. Run the server:
```bash
go run .
```

The server will start on `http://localhost:8080`

### Docker Setup

Alternatively, you can use Docker:

```bash
# Build the Docker image
docker build -t nlp-automation-backend .

# Run the container
docker run -p 8080:8080 -e OPENAI_API_KEY=your-key nlp-automation-backend
```

Or use the docker-compose from the root directory:
```bash
cd ..
docker-compose up -d
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `OPENAI_API_KEY` | Your OpenAI API key | Required |
| `DB_PATH` | SQLite database file path | `./nlp-automation.db` |
| `UPLOAD_DIR` | Directory for uploaded files | `./uploads` |
| `OUTPUT_DIR` | Directory for generated scripts | `./outputs` |
| `TEMP_DIR` | Directory for temporary files | `./temp` |
| `MAX_SCRIPT_SIZE` | Maximum script size in bytes | `10485760` |
| `COMPILATION_TIMEOUT` | Compilation timeout in seconds | `300` |

## API Endpoints

### Tasks

- `POST /api/tasks` - Create a new automation task
- `GET /api/tasks` - Get all tasks (with optional status filter)
- `GET /api/tasks/:id` - Get a specific task
- `GET /api/tasks/:id/download` - Download the generated script

### System

- `GET /api/health` - Health check
- `GET /api/system/info` - System information and capabilities
- `GET /api/stats` - Usage statistics

### Templates

- `GET /api/templates` - Get available script templates
- `GET /api/templates?category=:category` - Get templates by category

### Example Usage

#### Create a Task

```bash
curl -X POST http://localhost:8080/api/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Rename all files in a directory by adding the current date as prefix",
    "platform": "windows",
    "language": "powershell"
  }'
```

#### Get Task Status

```bash
curl http://localhost:8080/api/tasks/your-task-id
```

#### Download Generated Script

```bash
curl -O http://localhost:8080/api/tasks/your-task-id/download
```

## Project Structure

```
nlp-automation-backend/
├── main.go                 # Application entry point
├── go.mod                  # Go module definition
├── .env.example            # Environment variables template
├── README.md              # This file
├── internal/
│   ├── config/            # Configuration management
│   │   └── config.go
│   ├── database/          # Database models and operations
│   │   ├── models.go
│   │   └── database.go
│   ├── services/          # Business logic
│   │   ├── ai_service.go      # AI/OpenAI integration
│   │   ├── script_service.go  # Script generation and validation
│   │   ├── compiler_service.go # Script compilation
│   │   └── task_service.go    # Task orchestration
│   └── api/               # HTTP API layer
│       ├── handlers.go
│       └── server.go
├── uploads/               # User uploaded files
├── outputs/               # Generated scripts
└── temp/                  # Temporary files
```

## Supported Languages and Platforms

### Script Languages
- **PowerShell** (.ps1) - Windows automation
- **Bash** (.sh) - Linux/macOS shell scripts
- **Python** (.py) - Cross-platform scripting
- **Batch** (.bat) - Windows batch files

### Compiled Languages
- **Go** - Cross-platform compiled executables
- **C** - Native compiled executables
- **C++** - Native compiled executables
- **Rust** - Safe systems programming

### Platforms
- **Windows** - PowerShell, Batch, compiled executables
- **Linux** - Bash, Python, compiled executables
- **macOS** - Bash, Python, compiled executables
- **Cross-platform** - Languages that work on multiple platforms

## Safety Features

NLP-Automation includes several safety features to protect users:

1. **Script Analysis**: AI-powered analysis to detect potentially dangerous operations
2. **Pattern Detection**: Built-in patterns to identify risky commands
3. **Safety Warnings**: Clear warnings about destructive operations
4. **Input Validation**: Validates all user inputs and script content
5. **Size Limits**: Prevents generation of excessively large scripts

## Development

### Running Tests

```bash
go test ./...
```

### Building for Production

```bash
go build -o nlp-automation-backend main.go
```

### Docker Support

```bash
# Build Docker image
docker build -t nlp-automation-backend .

# Run with Docker
docker run -p 8080:8080 -e OPENAI_API_KEY=your-key nlp-automation-backend
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

For support, please open an issue on GitHub or contact the development team.

## Roadmap

- [ ] WebSocket support for real-time updates
- [ ] User authentication and authorization
- [ ] Script execution sandbox
- [ ] More language support (Java, C#, etc.)
- [ ] Advanced template system
- [ ] Scheduled task execution
- [ ] Script version control
- [ ] Performance optimizations
- [ ] Monitoring and logging improvements

## Security Considerations

- Always review generated scripts before execution
- Use environment variables for sensitive configuration
- Implement proper authentication in production
- Monitor script execution and outputs
- Regular security audits of generated content

## Troubleshooting

### Common Issues

1. **OpenAI API Key Not Set**: Ensure your `.env` file contains a valid OpenAI API key
2. **Compilation Failures**: Ensure required compilers are installed on your system
3. **Permission Errors**: Check file permissions for upload/output directories
4. **Database Issues**: Ensure the database directory is writable

### Debug Mode

Set `GIN_MODE=debug` in your `.env` file for detailed logging.

### Logging

Application logs are written to stdout. In production, consider using a log aggregation service.
