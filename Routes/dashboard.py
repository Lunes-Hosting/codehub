from flask import Blueprint, render_template, request, redirect, url_for, flash, session
import bcrypt
from database import DatabaseManager as db
from config import DASHBOARD_URL
from Routes.auth import login_required

dashboard = Blueprint('dashboard', __name__)

@dashboard.route('/')
@login_required
def index():
    return render_template('index.html', username=session['email'], user_id=session['user_id'])