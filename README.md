# Large Scale Flask Application

This is a production-ready Flask application boilerplate designed for large-scale applications with MySQL integration.

## Project Structure

```
.
├── app/
│   ├── __init__.py          # Application factory
│   ├── api/                 # API blueprint
│   ├── auth/                # Authentication blueprint
│   ├── models/              # Database models
│   └── errors.py            # Error handlers
├── migrations/              # Database migrations
├── logs/                    # Application logs
├── .env                     # Environment variables
├── .env.example            # Environment variables template
├── config.py               # Configuration settings
├── requirements.txt        # Python dependencies
├── app.py                  # Application entry point
└── README.md              # Project documentation
```

## Features

- Factory pattern for application creation
- Blueprint-based structure for modularity
- SQLAlchemy integration with MySQL
- JWT authentication
- CORS support
- Error handling
- Logging configuration
- Environment-based configuration
- Database migrations support

## Setup

1. Create a virtual environment:
   ```bash
   python -m venv venv
   source venv/bin/activate  # On Windows: venv\Scripts\activate
   ```

2. Install dependencies:
   ```bash
   pip install -r requirements.txt
   ```

3. Copy `.env.example` to `.env` and configure your environment variables:
   ```bash
   cp .env.example .env
   ```

4. Initialize the database:
   ```bash
   flask db init
   flask db migrate
   flask db upgrade
   ```

5. Run the application:
   ```bash
   flask run
   ```

## Development

- Add new models in `app/models/`
- Create new API endpoints in `app/api/`
- Add authentication routes in `app/auth/`
- Configure database in `.env`

## Production Deployment

1. Set environment variables:
   ```bash
   export FLASK_ENV=production
   ```

2. Use a production WSGI server:
   ```bash
   gunicorn app:app
   ```

## Contributing

1. Create a new branch
2. Make your changes
3. Submit a pull request

## License

This project is licensed under the MIT License.
