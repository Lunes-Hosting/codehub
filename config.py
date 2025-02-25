"""
Configuration Module
==================

Centralizes configuration settings loaded from environment variables.
"""

import os
from dotenv import load_dotenv

# Load environment variables
load_dotenv()

# Database configuration
HOST = os.getenv('MYSQL_HOST')
USER = os.getenv('MYSQL_USER')
PASSWORD = os.getenv('MYSQL_PASSWORD')
DATABASE = os.getenv('MYSQL_DATABASE')

# Application configuration
SECRET_KEY = os.getenv('SECRET_KEY')
DASHBOARD_URL = os.getenv('DASHBOARD_URL')
