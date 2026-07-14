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

  // NEW: Add event listeners to enable dynamic interactions
  attachResultInteractions();
  setupSearchFilters();
}

// NEW: Handle result clicks and interactions
function attachResultInteractions() {
  const results = document.querySelectorAll('.result');
  results.forEach(result => {
    result.addEventListener('click', function(e) {
      // Example: Log result selection for future features
      const title = this.querySelector('a')?.textContent || '';
      const url = this.querySelector('a')?.href || '';
      console.log('Result selected:', { title, url });
      // Could POST this to /api/save_note or similar
    });
  });
}

// NEW: Add filtering/sorting capabilities
function setupSearchFilters() {
  // You can add filter UI elements here
  // For example: sort by relevance, date, etc.
  const filterContainer = document.querySelector('body');
  
  // Example: Add a simple filter button
  const filterBtn = document.createElement('button');
  filterBtn.textContent = 'Add Filter';
  filterBtn.id = 'filter-btn';
  filterBtn.style.marginBottom = '1rem';
  filterBtn.style.padding = '0.5rem 1rem';
  filterBtn.style.backgroundColor = 'var(--accent-color)';
  filterBtn.style.color = 'white';
  filterBtn.style.border = 'none';
  filterBtn.style.borderRadius = 'var(--border-radius)';
  filterBtn.style.cursor = 'pointer';
  
  filterBtn.addEventListener('click', () => {
    console.log('Filter clicked - ready for backend API call');
    // Could call /api/search with additional parameters
  });

  filterContainer.insertBefore(filterBtn, filterContainer.querySelector('h2')?.nextElementSibling);
}

// NEW: Helper to make API calls to Go backend
async function callBackendAPI(endpoint, method = 'GET', data = null) {
  try {
    const options = {
      method: method,
      headers: { 'Content-Type': 'application/json' }
    };
    
    if (data) {
      options.body = JSON.stringify(data);
    }

    const response = await fetch(endpoint, options);
    
    if (!response.ok) {
      throw new Error(`API Error: ${response.status}`);
    }

    return await response.json();
  } catch (error) {
    console.error('Backend API call failed:', error);
    return null;
  }
}

// NEW: Example function to fetch additional results
export async function loadMoreResults(query, page = 1) {
  const result = await callBackendAPI(`/api/search?query=${encodeURIComponent(query)}&page=${page}`);
  if (result && result.results) {
    appendResultsToDOM(result.results);
  }
}

// NEW: Helper to append results to the page
function appendResultsToDOM(results) {
  const resultsContainer = document.querySelector('body');
  results.forEach(result => {
    const resultDiv = document.createElement('div');
    resultDiv.className = 'result';
    resultDiv.innerHTML = `
      <a href="${escapeHTML(result.URL)}">${escapeHTML(result.Title)}</a>
      <p>${escapeHTML(result.Snippet)}</p>
    `;
    resultsContainer.appendChild(resultDiv);
  });
}

// HTML escape helper
function escapeHTML(text) {
  const map = {
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    '"': '&quot;',
    "'": '&#39;'
  };
  return text.replace(/[&<>"']/g, m => map[m]);
}

// Automatically initialize the page enhancements when loaded
document.addEventListener('DOMContentLoaded', initResultsPage);