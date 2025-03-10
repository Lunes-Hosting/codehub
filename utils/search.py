"""
Search utilities for CodeHub
Provides fuzzy search capabilities for projects
"""
import re
from difflib import SequenceMatcher

def preprocess_text(text):
    """
    Preprocess text for searching:
    - Convert to lowercase
    - Remove special characters
    - Remove extra whitespace
    """
    if not text:
        return ""
    
    # Convert to string if not already
    text = str(text)
    
    # Convert to lowercase
    text = text.lower()
    
    # Remove quotes and special characters
    text = re.sub(r'[\'"\[\](){}<>]', '', text)
    
    # Replace multiple spaces with a single space
    text = re.sub(r'\s+', ' ', text)
    
    return text.strip()

def calculate_similarity(query, text):
    """
    Calculate similarity between query and text using SequenceMatcher
    Returns a score between 0 and 1, where 1 is a perfect match
    """
    if not query or not text:
        return 0
    
    # Preprocess both query and text
    query = preprocess_text(query)
    text = preprocess_text(text)
    
    # If either is empty after preprocessing, return 0
    if not query or not text:
        return 0
    
    # Calculate similarity ratio
    return SequenceMatcher(None, query, text).ratio()

def search_projects(projects, query, min_score=0.3):
    """
    Search projects by query using fuzzy matching
    
    Args:
        projects: List of project dictionaries or tuples
        query: Search query string
        min_score: Minimum similarity score to include in results (0-1)
        
    Returns:
        List of (project, score) tuples sorted by score in descending order
    """
    if not query or not projects:
        return []
    
    # Preprocess query
    processed_query = preprocess_text(query)
    
    # If query is empty after preprocessing, return all projects
    if not processed_query:
        return [(project, 1.0) for project in projects]
    
    results = []
    
    for project in projects:
        # Check if project is a dictionary or a tuple
        if isinstance(project, dict):
            # Dictionary access
            name = project.get('name', '')
            owner_name = project.get('owner_name', '')
            description = project.get('description', '')
            github_link = project.get('github_link', '')
        else:
            # Tuple access - assuming order from the SQL query
            # id, name, description, github_link, create_time, owner, owner_name, enabled
            name = project[1] if len(project) > 1 else ''
            description = project[2] if len(project) > 2 else ''
            github_link = project[3] if len(project) > 3 else ''
            owner_name = project[6] if len(project) > 6 else ''
        
        # Calculate similarity scores for different fields
        name_score = calculate_similarity(processed_query, name) * 1.5  # Weight name higher
        owner_score = calculate_similarity(processed_query, owner_name)
        description_score = calculate_similarity(processed_query, description)
        github_score = calculate_similarity(processed_query, github_link)
        
        # Get the maximum score
        max_score = max(name_score, owner_score, description_score, github_score)
        
        # If any field has a score above the threshold, include the project
        if max_score >= min_score:
            results.append((project, max_score))
    
    # Sort by score in descending order
    return sorted(results, key=lambda x: x[1], reverse=True)
