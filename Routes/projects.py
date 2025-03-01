from flask import Blueprint, render_template, request, redirect, url_for, flash, session, jsonify
import bcrypt
from database import DatabaseManager as db
from config import DASHBOARD_URL, PTERODACTYL_URL, PTERODACTYL_ADMIN_KEY, AUTODEPLOY_NEST_ID
from Routes.auth import login_required
from flask_wtf import FlaskForm
from wtforms import StringField, PasswordField, SubmitField
from wtforms.validators import DataRequired, Length
import datetime
from Managers.User import User
import markdown
import json
from security import safe_requests


projects = Blueprint('projects', __name__)

class ProjectForm(FlaskForm):
    name = StringField('Name', validators=[DataRequired(), Length(min=3, max=25)])
    github_link = StringField('Github link', validators=[DataRequired()])
    submit = SubmitField('Submit')


@projects.route("/get_egg_variables", methods=['POST'])
def get_egg_variables():
    selected_egg_id = request.form.get('egg')  # Get the selected egg id from the form
    if not selected_egg_id:
        return jsonify({'error': 'No egg selected'}), 400
    
    # Fetch the available eggs and their variables
    egg_info = safe_requests.get(f"{PTERODACTYL_URL}/api/application/nests/{AUTODEPLOY_NEST_ID}/eggs/{selected_egg_id}?include=variables", 
                                  headers={"Authorization": f"Bearer {PTERODACTYL_ADMIN_KEY}"}).json()
    
    
    if egg_info:
        egg_variables = egg_info['attributes']['relationships']['variables']['data']
        print(egg_variables)
        return jsonify({'egg_variables': egg_variables})
    
    return jsonify({'error': 'Egg not found'}), 404


@projects.route("/create_project", methods=['GET', 'POST'])
def create_project():
    user = User(session)
    egg_variables = []  # This will hold the variables for the selected egg

    if request.method == 'POST':
        formig = request.form
        print(formig)
        environment = {}
        for variable in formig:
            if variable.startswith("variable_"):

                environment[variable.replace("variable_", "")] = request.form.get(variable)

        name = request.form.get('name')
        egg_id = request.form.get("egg")
        print(egg_id)
        github_link = request.form.get('github_link')

        if not name or not github_link:
            flash("Both fields are required!", "error")
        else:
            flash(f"Project '{name}' created successfully!", "success")
            owner_id = user.get_id()
            db.execute_query("INSERT INTO projects (create_time, name, owner, github_link, environment, egg_id) \
                             VALUES (%s, %s, %s, %s, %s, %s)",
                            (datetime.datetime.now(), str(name), owner_id, github_link, json.dumps(environment), egg_id))
            return redirect(url_for('projects.home'))  # Redirect to clear form after submission

    # For GET requests, fetch the available eggs
    available_eggs = safe_requests.get(f"{PTERODACTYL_URL}/api/application/nests/{AUTODEPLOY_NEST_ID}/eggs?include=variables", 
                                  headers={"Authorization": f"Bearer {PTERODACTYL_ADMIN_KEY}"}).json()['data']

    form = ProjectForm()
    return render_template("create_project.html", form=form, available_eggs=available_eggs, egg_variables=egg_variables)

@projects.route('/project/<int:proj_id>', methods=['GET', 'POST'])
@login_required
def project(proj_id):
    user = User(session)
    project = db.execute_query("SELECT * FROM projects WHERE id = %s", (proj_id,))

    if not project:
        flash("Project not found!", "danger")
        return redirect(url_for("dashboard.dashboard"))

    user_id = user.get_id()
    is_owner = str(project['owner']) == str(user_id)

    # Load environment variables from the JSON field
    env_variables = json.loads(project['environment']) if project['environment'] else {}

    if request.method == 'POST' and is_owner:
        new_name = request.form.get("name")
        new_github_link = request.form.get("github_link")
        new_description = request.form.get("description")

        # Fetch and update environment variables from the form
        updated_env = {}
        for key in request.form:
            if key.startswith("variable_"):
                updated_env[key.replace("variable_", "")] = request.form.get(key)

        # Update project details in the database
        db.execute_query("""
            UPDATE projects 
            SET name = %s, github_link = %s, description = %s, environment = %s
            WHERE id = %s
        """, (new_name, new_github_link, new_description, json.dumps(updated_env), proj_id))

        flash("Project updated successfully!", "success")
        return redirect(url_for('projects.project', proj_id=proj_id))  # Refresh page

    project_description_html = markdown.markdown(project['description'])
    return render_template('project.html', project=project, env_variables=env_variables, is_owner=is_owner, project_description_html=project_description_html)


@projects.route('/projects')
@login_required
def home():
    projects = db.execute_query("SELECT * FROM projects WHERE enabled = %s", (1,), fetch_all=True)
    return render_template('projects.html', projects=projects)
