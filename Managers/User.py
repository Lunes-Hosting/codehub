from flask import session
from database import DatabaseManager as db

class User():
    def __init__(self, session: dict):
        self.session: dict = session

    def get_id(self) -> int:
        if "user_id" in session:
            return session['user_id']
        
        user_id = db.execute_query("SELECT id FROM users WHERE email = %s", (session['email'],))[0]
        session ['user_id'] = user_id
        return user_id