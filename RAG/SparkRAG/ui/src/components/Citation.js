import React from 'react';
import { FiExternalLink, FiFileText } from 'react-icons/fi';
import './Citation.css';

function Citation({ citation }) {
  return (
    <div className="citation">
      <div className="citation-icon">
        <FiFileText size={14} />
      </div>
      <div className="citation-content">
        <div className="citation-source">
          <strong>{citation.source || 'Unknown Source'}</strong>
          {citation.section && <span className="citation-section">{citation.section}</span>}
        </div>
        {citation.document && (
          <div className="citation-document">{citation.document}</div>
        )}
        <div className="citation-meta">
          {citation.page && <span>Page {citation.page}</span>}
          {citation.score && (
            <span className="citation-score">
              Match: {(citation.score * 100).toFixed(0)}%
            </span>
          )}
        </div>
      </div>
      {citation.url && (
        <a href={citation.url} target="_blank" rel="noopener noreferrer" className="citation-link">
          <FiExternalLink size={14} />
        </a>
      )}
    </div>
  );
}

export default Citation;
