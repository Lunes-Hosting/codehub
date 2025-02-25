from flask import Flask, render_template, redirect, url_for, session
from Routes.auth import auth, login_required
from Routes.dashboard import dashboard
from config import SECRET_KEY

app = Flask(__name__)
app.secret_key = SECRET_KEY

# Register blueprints
app.register_blueprint(auth, url_prefix='/auth')
app.register_blueprint(dashboard)

@app.route('/')
@login_required
def index():
    return render_template('index.html', username=session['username'])

if __name__ == '__main__':
    app.run(debug=True)
