from flask import Blueprint, render_template, request, redirect, url_for, flash, session
from functools import wraps
import bcrypt
from database import DatabaseManager as db
from config import DASHBOARD_URL
from urllib.parse import urlparse, urljoin

auth = Blueprint('auth', __name__)

def is_safe_url(target):
    ref_url = urlparse(request.host_url)
    test_url = urlparse(urljoin(request.host_url, target))
    return test_url.scheme in ('http', 'https') and ref_url.netloc == test_url.netloc

def login_required(f):
    @wraps(f)
    def decorated_function(*args, **kwargs):
        if 'email' not in session:
            # Store the full URL they were trying to access
            next_page = request.url
            if is_safe_url(next_page):
                session['next'] = next_page
            flash('Please log in to access this page', 'error')
            return redirect(url_for('auth.login'))
        return f(*args, **kwargs)
    return decorated_function

@auth.route('/login', methods=['GET', 'POST'])
def login():
    if request.method == 'POST':
        email = request.form['email']
        password = request.form['password']
        
        try:
            user = db.execute_query(
                'SELECT * FROM users WHERE email = %s',
                (email,),
                fetch_all=False
            )
            
            if user:
                stored_hash = user['password'].encode('utf-8') if user['password'] else b''
                try:
                    if bcrypt.checkpw(password.encode('utf-8'), stored_hash):
                        session['email'] = email
                        session['user_id'] = user['id']
                        session['role'] = user['role']
                        flash('Logged in successfully!', 'success')
                        
                        # Redirect to the stored next page if it exists and is safe
                        next_page = session.pop('next', None)
                        if next_page and is_safe_url(next_page):
                            return redirect(next_page)
                        return redirect(url_for('dashboard.index'))
                except ValueError as e:
                    print(f"Password hash error: {str(e)}")
                    flash('Invalid login credentials', 'error')
                    return render_template('login.html')
            
            flash('Invalid username or password', 'error')
            
        except Exception as e:
            print(f"Login error: {str(e)}")
            flash('An error occurred during login. Please try again.', 'error')
    
    return render_template('login.html')

@auth.route('/register', methods=['GET', 'POST'])
def register():
    return redirect(f"{DASHBOARD_URL}/register")

@auth.route('/logout')
def logout():
    session.clear()
    return redirect(url_for('auth.login'))
