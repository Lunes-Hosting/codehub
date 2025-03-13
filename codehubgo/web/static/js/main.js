/**
 * Main JavaScript file for Lunes Code Hub
 */

document.addEventListener('DOMContentLoaded', function() {
    // Initialize Bootstrap tooltips
    var tooltipTriggerList = [].slice.call(document.querySelectorAll('[data-bs-toggle="tooltip"]'));
    var tooltipList = tooltipTriggerList.map(function (tooltipTriggerEl) {
        return new bootstrap.Tooltip(tooltipTriggerEl);
    });
    
    // Initialize Bootstrap popovers
    var popoverTriggerList = [].slice.call(document.querySelectorAll('[data-bs-toggle="popover"]'));
    var popoverList = popoverTriggerList.map(function (popoverTriggerEl) {
        return new bootstrap.Popover(popoverTriggerEl);
    });
    
    // Auto-dismiss alerts
    setTimeout(function() {
        var alerts = document.querySelectorAll('.alert:not(.alert-permanent)');
        alerts.forEach(function(alert) {
            var bsAlert = new bootstrap.Alert(alert);
            bsAlert.close();
        });
    }, 5000);
    
    // Toggle project visibility
    var togglePublishButtons = document.querySelectorAll('.btn-toggle-publish');
    togglePublishButtons.forEach(function(button) {
        button.addEventListener('click', function(e) {
            e.preventDefault();
            var form = this.closest('form');
            if (form) {
                form.submit();
            }
        });
    });
    
    // Toggle project star
    var toggleStarButtons = document.querySelectorAll('.btn-star');
    toggleStarButtons.forEach(function(button) {
        button.addEventListener('click', function(e) {
            e.preventDefault();
            var form = this.closest('form');
            if (form) {
                form.submit();
            }
        });
    });
    
    // Confirm delete project
    var deleteProjectForms = document.querySelectorAll('.delete-project-form');
    deleteProjectForms.forEach(function(form) {
        form.addEventListener('submit', function(e) {
            e.preventDefault();
            if (confirm('Are you sure you want to delete this project? This action cannot be undone.')) {
                this.submit();
            }
        });
    });
    
    // Handle markdown preview for project creation/editing
    var markdownTextarea = document.getElementById('content');
    var previewButton = document.getElementById('preview-button');
    var editButton = document.getElementById('edit-button');
    var contentEditor = document.getElementById('content-editor');
    var contentPreview = document.getElementById('content-preview');
    
    if (markdownTextarea && previewButton && editButton && contentEditor && contentPreview) {
        previewButton.addEventListener('click', function() {
            // Show preview, hide editor
            contentEditor.classList.add('d-none');
            contentPreview.classList.remove('d-none');
            previewButton.classList.add('active');
            editButton.classList.remove('active');
            
            // Get the markdown content
            var markdown = markdownTextarea.value;
            
            // Make an AJAX request to render the markdown
            fetch('/api/render-markdown', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ markdown: markdown }),
            })
            .then(response => response.json())
            .then(data => {
                contentPreview.innerHTML = data.html;
            })
            .catch(error => {
                console.error('Error rendering markdown:', error);
                contentPreview.innerHTML = '<div class="alert alert-danger">Error rendering markdown preview.</div>';
            });
        });
        
        editButton.addEventListener('click', function() {
            // Show editor, hide preview
            contentEditor.classList.remove('d-none');
            contentPreview.classList.add('d-none');
            previewButton.classList.remove('active');
            editButton.classList.add('active');
        });
    }
});
