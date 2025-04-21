/**
 * Global utility functions and event handlers for the application
 */

// Prevent context menu on all images throughout the application
document.addEventListener('DOMContentLoaded', () => {
  // Select all images in the document and prevent context menu
  document.addEventListener('contextmenu', (e) => {
    // Check if the target is an image or a parent element contains an image
    if (e.target.tagName === 'IMG' || 
        e.target.querySelector('img') || 
        e.target.classList.contains('chore-image') ||
        e.target.classList.contains('routine-image')) {
      e.preventDefault();
      return false;
    }
  }, false);
});