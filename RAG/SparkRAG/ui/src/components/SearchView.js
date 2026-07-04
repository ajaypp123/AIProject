import React, { useState } from 'react';
import { search } from '../api';
import './SearchView.css';
import { FiSearch, FiLoader } from 'react-icons/fi';

function SearchView() {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const handleSearch = async (e) => {
    e.preventDefault();
    if (!query.trim()) return;

    setLoading(true);
    setError(null);

    try {
      const response = await search(query);
      setResults(response.results || []);
    } catch (err) {
      setError(err.message || 'Search failed');
      setResults([]);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="search-view">
      <div className="search-container">
        <form onSubmit={handleSearch} className="search-form">
          <div className="search-input-group">
            <FiSearch className="search-icon" size={20} />
            <input
              type="text"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="Search documents..."
              className="search-input"
            />
          </div>
          <button type="submit" disabled={loading || !query.trim()} className="search-button">
            {loading ? <FiLoader className="spinning" size={18} /> : 'Search'}
          </button>
        </form>
      </div>

      {error && (
        <div className="search-error">
          <p>{error}</p>
        </div>
      )}

      <div className="results-container">
        {results.length === 0 && !loading && query && (
          <div className="no-results">
            <p>No results found for "{query}"</p>
          </div>
        )}

        {results.map((result, idx) => (
          <div key={idx} className="result-card">
            <h3>{result.document}</h3>
            <p className="result-content">{result.content.substring(0, 200)}...</p>
            <div className="result-meta">
              <span className="result-score">Match: {(result.score * 100).toFixed(0)}%</span>
              {result.citation && (
                <span className="result-source">{result.citation.source}</span>
              )}
            </div>
          </div>
        ))}

        {results.length > 0 && (
          <div className="results-summary">
            Found {results.length} result{results.length !== 1 ? 's' : ''}
          </div>
        )}
      </div>
    </div>
  );
}

export default SearchView;
