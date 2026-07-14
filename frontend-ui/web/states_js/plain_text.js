export function initResultsPage() {
  // Update the document title based on the generated search query
  const heading = document.querySelector('h2');
  if (heading) {
    const queryText = heading.textContent;
    document.title = queryText ? queryText : "Search Results";
  }

  // Ensure all result links open in a new tab with secure openers
  const resultLinks = document.querySelectorAll('.result a');
  resultLinks.forEach(link => {
    link.setAttribute('target', '_blank');
    link.setAttribute('rel', 'noopener noreferrer');
  });
}

// Automatically initialize the page enhancements when loaded
document.addEventListener('DOMContentLoaded', initResultsPage);
