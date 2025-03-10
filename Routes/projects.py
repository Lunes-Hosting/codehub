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
import requests
import json


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
    egg_info = requests.get(f"{PTERODACTYL_URL}/api/application/nests/{AUTODEPLOY_NEST_ID}/eggs/{selected_egg_id}?include=variables", 
                                  headers={"Authorization": f"Bearer {PTERODACTYL_ADMIN_KEY}"}).json()
    
    
    if egg_info:
        egg_variables = egg_info['attributes']['relationships']['variables']['data']
        print(egg_variables)
        return jsonify({'egg_variables': egg_variables})
    
    return jsonify({'error': 'Egg not found'}), 404


@projects.route("/create_project", methods=['GET', 'POST'])
def create_project():
    user = User(session)
    user_id = user.get_id()
    username = user.get_email()  # Get the user's email to display in navbar
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
            db.execute_query("INSERT INTO projects (create_time, name, owner, github_link, environment, egg_id, enabled) \
                             VALUES (%s, %s, %s, %s, %s, %s, %s)",
                            (datetime.datetime.now(), str(name), owner_id, github_link, json.dumps(environment), egg_id, 0))
            return redirect(url_for('projects.home'))  # Redirect to clear form after submission

    # For GET requests, fetch the available eggs
    resp = requests.get(f"{PTERODACTYL_URL}/api/application/nests/{AUTODEPLOY_NEST_ID}/eggs?include=variables", 
                                  headers={"Authorization": f"Bearer {PTERODACTYL_ADMIN_KEY}"})
    print(f"{PTERODACTYL_URL}/api/application/nests/{AUTODEPLOY_NEST_ID}/eggs?include=variables")
    print(resp.text)
    available_eggs = resp.json()['data']

    form = ProjectForm()
    return render_template("create_project.html", form=form, available_eggs=available_eggs, egg_variables=egg_variables, username=username, user_id=user_id)

@projects.route('/project/<int:proj_id>', methods=['GET', 'POST'])
@login_required
def project(proj_id):
    user = User(session)
    user_id = user.get_id()
    username = user.get_email()  # Get the user's email to display in navbar
    
    # Get the project details
    project = db.execute_query(
        "SELECT * FROM projects WHERE id = %s", 
        (proj_id,)
    )
    
    if not project:
        flash("Project not found", "error")
        return redirect(url_for('projects.home'))
    
    # Check if the user is the owner of the project
    is_owner = project['owner'] == user_id
    
    # Check if the project is published or if the user is the owner
    if project['enabled'] == 0 and not is_owner:
        flash("You don't have permission to view this project", "danger")
        return redirect(url_for('projects.home'))
    
    # Get the owner's information
    owner_info = db.execute_query(
        "SELECT id, name FROM users WHERE id = %s", 
        (project['owner'],)
    )
    
    # Load environment variables from the project data
    env_variables = {}
    if project.get('environment'):
        try:
            env_variables = json.loads(project['environment'])
        except (json.JSONDecodeError, TypeError):
            # If environment is not valid JSON, use empty dict
            env_variables = {}
    
    # Get star information
    # Check if the user has starred this project
    is_starred = db.execute_query(
        "SELECT * FROM project_stars WHERE project_id = %s AND user_id = %s", 
        (proj_id, user_id)
    ) is not None
    
    # Get the total number of stars for this project
    star_count_result = db.execute_query(
        "SELECT COUNT(*) as count FROM project_stars WHERE project_id = %s", 
        (proj_id,)
    )
    star_count = star_count_result['count'] if star_count_result else 0
    
    # Convert markdown description to HTML
    project_description_html = markdown.markdown(project['description']) if project['description'] else ""
    
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

    return render_template('project.html', 
                          project=project, 
                          is_owner=is_owner, 
                          env_variables=env_variables, 
                          project_description_html=project_description_html,
                          is_starred=is_starred,
                          star_count=star_count,
                          username=username,
                          user_id=user_id,
                          owner_info=owner_info)


@projects.route('/project/<int:proj_id>/toggle_star', methods=['POST'])
@login_required
def toggle_star(proj_id):
    try:
        user = User(session)
        user_id = user.get_id()
        
        print(f"Toggle star for project {proj_id} by user {user_id}")
        
        # Check if the user has already starred this project
        existing_star = db.execute_query(
            "SELECT * FROM project_stars WHERE project_id = %s AND user_id = %s", 
            (proj_id, user_id)
        )
        
        print(f"Existing star: {existing_star}")
        
        if existing_star:
            # User has already starred this project, so remove the star
            print(f"Removing star for project {proj_id} by user {user_id}")
            db.execute_query(
                "DELETE FROM project_stars WHERE project_id = %s AND user_id = %s", 
                (proj_id, user_id)
            )
            is_starred = False
        else:
            # User has not starred this project, so add a star
            try:
                print(f"Adding star for project {proj_id} by user {user_id}")
                db.execute_query(
                    "INSERT INTO project_stars (project_id, user_id) VALUES (%s, %s)", 
                    (proj_id, user_id)
                )
                is_starred = True
            except Exception as e:
                print(f"Error adding star: {str(e)}")
                # Handle any database errors (like duplicate entries)
                return jsonify({
                    'status': 'error',
                    'message': str(e)
                }), 500
        
        # Get the updated star count
        print(f"Getting star count for project {proj_id}")
        star_count_result = db.execute_query(
            "SELECT COUNT(*) as count FROM project_stars WHERE project_id = %s", 
            (proj_id,)
        )
        star_count = star_count_result['count'] if star_count_result else 0
        
        print(f"Star count: {star_count}")
        
        return jsonify({
            'status': 'success',
            'is_starred': is_starred,
            'star_count': star_count
        })
    except Exception as e:
        print(f"Unexpected error in toggle_star: {str(e)}")
        import traceback
        traceback.print_exc()
        return jsonify({
            'status': 'error',
            'message': f"An unexpected error occurred: {str(e)}"
        }), 500


@projects.route('/project/<int:proj_id>/star_status', methods=['GET'])
@login_required
def star_status(proj_id):
    try:
        user = User(session)
        user_id = user.get_id()
        
        print(f"Checking star status for project {proj_id} by user {user_id}")
        
        # Check if the user has already starred this project
        existing_star = db.execute_query(
            "SELECT * FROM project_stars WHERE project_id = %s AND user_id = %s", 
            (proj_id, user_id)
        )
        
        print(f"Existing star: {existing_star}")
        
        # Get the star count for this project
        star_count_result = db.execute_query(
            "SELECT COUNT(*) as count FROM project_stars WHERE project_id = %s",
            (proj_id,)
        )
        star_count = star_count_result['count'] if star_count_result else 0
        
        print(f"Star count: {star_count}")
        
        return jsonify({
            "status": "success",
            "is_starred": existing_star is not None,
            "star_count": star_count
        })
    except Exception as e:
        print(f"Unexpected error in star_status: {str(e)}")
        import traceback
        traceback.print_exc()
        return jsonify({
            'status': 'error',
            'message': f"An unexpected error occurred: {str(e)}"
        }), 500


@projects.route('/project/<int:proj_id>/delete', methods=['POST'])
@login_required
def delete_project(proj_id):
    try:
        user = User(session)
        user_id = user.get_id()
        
        # Get the project details to check ownership
        project = db.execute_query(
            "SELECT * FROM projects WHERE id = %s", 
            (proj_id,)
        )
        
        if not project:
            flash("Project not found", "danger")
            return redirect(url_for('projects.home'))
        
        # Check if the user is the owner of the project
        if project['owner'] != user_id:
            flash("You don't have permission to delete this project", "danger")
            return redirect(url_for('projects.project', proj_id=proj_id))
        
        # Delete associated stars first (foreign key constraint)
        db.execute_query(
            "DELETE FROM project_stars WHERE project_id = %s",
            (proj_id,)
        )
        
        # Delete the project
        db.execute_query(
            "DELETE FROM projects WHERE id = %s AND owner = %s",
            (proj_id, user_id)
        )
        
        flash("Project successfully deleted", "success")
        return redirect(url_for('projects.home'))
    
    except Exception as e:
        print(f"Error deleting project: {str(e)}")
        import traceback
        traceback.print_exc()
        flash(f"An error occurred while deleting the project: {str(e)}", "danger")
        return redirect(url_for('projects.project', proj_id=proj_id))


@projects.route('/project/<int:proj_id>/toggle_publish', methods=['POST'])
@login_required
def toggle_publish(proj_id):
    try:
        user = User(session)
        user_id = user.get_id()
        
        print(f"Toggle publish for project {proj_id} by user {user_id}")
        
        # Get the project details to check ownership
        project = db.execute_query(
            "SELECT * FROM projects WHERE id = %s", 
            (proj_id,)
        )
        
        if not project:
            flash("Project not found", "danger")
            return redirect(url_for('projects.home'))
        
        # Check if the user is the owner of the project
        if project['owner'] != user_id:
            flash("You don't have permission to publish/unpublish this project", "danger")
            return redirect(url_for('projects.project', proj_id=proj_id))
        
        # Toggle the enabled status
        new_status = 0 if project['enabled'] == 1 else 1
        status_text = "published" if new_status == 1 else "unpublished"
        
        db.execute_query(
            "UPDATE projects SET enabled = %s WHERE id = %s AND owner = %s",
            (new_status, proj_id, user_id)
        )
        
        flash(f"Project successfully {status_text}", "success")
        return redirect(url_for('projects.project', proj_id=proj_id))
    
    except Exception as e:
        print(f"Error toggling publish status: {str(e)}")
        import traceback
        traceback.print_exc()
        flash(f"An error occurred while updating the project: {str(e)}", "danger")
        return redirect(url_for('projects.project', proj_id=proj_id))


@projects.route('/projects')
@login_required
def home():
    user = User(session)
    user_id = user.get_id()
    username = user.get_email()  # Get the current user's email to display in navbar
    
    # Get all projects owned by the current user (both published and unpublished)
    user_projects = db.execute_query(
        "SELECT * FROM projects WHERE owner = %s", 
        (user_id,), 
        fetch_all=True
    )
    
    # Get all published projects (including those owned by the current user)
    public_projects = db.execute_query(
        "SELECT * FROM projects WHERE enabled = %s", 
        (1,), 
        fetch_all=True
    )
    
    # Get star counts for all projects
    star_counts = {}
    starred_projects = set()
    
    # Get all project IDs
    project_ids = []
    for project in user_projects + public_projects:
        if project['id'] not in project_ids:
            project_ids.append(project['id'])
    
    if project_ids:
        # Get star counts for all projects
        star_count_results = db.execute_query(
            "SELECT project_id, COUNT(*) as count FROM project_stars WHERE project_id IN ({}) GROUP BY project_id".format(
                ','.join(['%s'] * len(project_ids))
            ), 
            project_ids,
            fetch_all=True
        )
        
        if star_count_results:
            for result in star_count_results:
                star_counts[result['project_id']] = result['count']
        
        # Get projects starred by the current user
        starred_by_user = db.execute_query(
            "SELECT project_id FROM project_stars WHERE user_id = %s AND project_id IN ({})".format(
                ','.join(['%s'] * len(project_ids))
            ), 
            [user_id] + project_ids,
            fetch_all=True
        )
        
        if starred_by_user:
            for result in starred_by_user:
                starred_projects.add(result['project_id'])
    
    return render_template("projects.html", user_projects=user_projects, public_projects=public_projects, 
                           star_counts=star_counts, starred_projects=starred_projects, username=username, user_id=user_id)


@projects.route('/user/<int:user_id>')
@login_required
def user_profile(user_id):
    user = User(session)
    current_user_id = user.get_id()
    username = user.get_email()  # Get the current user's email to display in navbar
    
    # Get the profile user's details
    profile_user = db.execute_query(
        "SELECT id, name, email, avatar, created_at, role, show_email FROM users WHERE id = %s", 
        (user_id,)
    )
    
    if not profile_user:
        flash("User not found", "danger")
        return redirect(url_for('projects.home'))
    
    # Get all published projects owned by this user
    user_projects = db.execute_query(
        "SELECT * FROM projects WHERE owner = %s AND (enabled = 1 OR owner = %s)", 
        (user_id, current_user_id), 
        fetch_all=True
    )
    
    # Get star counts for all projects
    star_counts = {}
    starred_projects = set()
    
    # Get all project IDs
    project_ids = []
    for project in user_projects:
        if project['id'] not in project_ids:
            project_ids.append(project['id'])
    
    if project_ids:
        # Get star counts for all projects
        star_count_results = db.execute_query(
            "SELECT project_id, COUNT(*) as count FROM project_stars WHERE project_id IN ({}) GROUP BY project_id".format(
                ','.join(['%s'] * len(project_ids))
            ), 
            project_ids,
            fetch_all=True
        )
        
        if star_count_results:
            for result in star_count_results:
                star_counts[result['project_id']] = result['count']
        
        # Get projects starred by the current user
        starred_by_user = db.execute_query(
            "SELECT project_id FROM project_stars WHERE user_id = %s AND project_id IN ({})".format(
                ','.join(['%s'] * len(project_ids))
            ), 
            [current_user_id] + project_ids,
            fetch_all=True
        )
        
        if starred_by_user:
            for result in starred_by_user:
                starred_projects.add(result['project_id'])
    
    return render_template('user_profile.html', 
                          profile_user=profile_user, 
                          user_projects=user_projects,
                          star_counts=star_counts,
                          starred_projects=starred_projects,
                          is_own_profile=(current_user_id == user_id),
                          username=username,
                          user_id=current_user_id)


@projects.route('/toggle_email_visibility', methods=['POST'])
@login_required
def toggle_email_visibility():
    user = User(session)
    user_id = user.get_id()
    
    # Get the form data
    show_email = 1 if request.form.get('show_email') else 0
    
    # Update the user's email visibility setting
    db.execute_query(
        "UPDATE users SET show_email = %s WHERE id = %s",
        (show_email, user_id)
    )
    
    flash("Email visibility settings updated successfully", "success")
    return redirect(url_for('projects.user_profile', user_id=user_id))