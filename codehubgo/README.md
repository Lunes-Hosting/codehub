# CodeHub Go

A Go implementation of the CodeHub project management application, originally built with Flask.

## Project Structure

```
codehubgo/
├── cmd/
│   └── server/          # Application entry point
├── internal/
│   ├── config/          # Configuration settings
│   ├── database/        # Database connection and queries
│   ├── handlers/        # HTTP request handlers
│   ├── middleware/      # HTTP middleware
│   ├── models/          # Data models
│   └── utils/           # Utility functions
├── web/
│   ├── static/          # Static assets (CSS, JS)
│   └── templates/       # HTML templates
├── .env                 # Environment variables (not in version control)
├── .env.example         # Example environment variables
├── go.mod               # Go module definition
└── go.sum               # Go module checksums
```

## Prerequisites

- Go 1.16 or higher
- MySQL database

## Setup

1. Clone the repository:
   ```
   git clone https://github.com/DWAA1660/codehubgo.git
   cd codehubgo
   ```

2. Copy the example environment file and edit it with your settings:
   ```
   cp .env.example .env
   ```

3. Edit the `.env` file with your database credentials and other settings.

4. Download dependencies:
   ```
   go mod tidy
   ```

## Building the Application

To build the application, run:

```
go build -o codehubgo.exe ./cmd/server/main.go
```

## Running the Application

After building, you can run the application:

```
./codehubgo.exe
```

Or directly with Go:

```
go run ./cmd/server/main.go
```

The server will start on the port specified in your `.env` file (default: 8080).

## Features

- User authentication and session management
- Project creation and management
- Project visibility controls (public/private)
- Project starring
- User profiles
- Search functionality

## API Endpoints

- `/auth/login` - User login
- `/auth/logout` - User logout
- `/auth/register` - User registration (redirects to dashboard)
- `/` - Dashboard
- `/projects` - List all projects
- `/project/{id}` - View/edit project
- `/project/create` - Create new project
- `/project/{id}/delete` - Delete project
- `/project/{id}/toggle-publish` - Toggle project visibility
- `/project/{id}/toggle-star` - Star/unstar project
- `/user/{id}` - User profile
- `/search` - Search projects
- `/api/egg-variables` - Get variables for a Pterodactyl egg

## License

[MIT License](LICENSE)
