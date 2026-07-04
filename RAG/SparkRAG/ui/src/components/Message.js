import React from 'react';
import ReactMarkdown from 'react-markdown';
import { FiCheckCircle, FiBot, FiUser } from 'react-icons/fi';
import './Message.css';

function Message({ message }) {
  return (
    <div className={`message message-${message.type}`}>
      <div className="message-icon">
        {message.type === 'user' ? (
          <FiUser size={18} />
        ) : (
          <FiBot size={18} />
        )}
      </div>
      <div className="message-content">
        {message.type === 'user' ? (
          <p>{message.content}</p>
        ) : (
          <ReactMarkdown
            components={{
              p: ({ node, ...props }) => <p {...props} />,
              code: ({ node, inline, ...props }) => 
                inline ? (
                  <code className="inline-code" {...props} />
                ) : (
                  <pre><code {...props} /></pre>
                ),
              ul: ({ node, ...props }) => <ul className="message-list" {...props} />,
              ol: ({ node, ...props }) => <ol className="message-list" {...props} />,
              h1: ({ node, ...props }) => <h3 {...props} />,
              h2: ({ node, ...props }) => <h4 {...props} />,
              h3: ({ node, ...props }) => <h5 {...props} />,
              blockquote: ({ node, ...props }) => <blockquote className="message-blockquote" {...props} />,
            }}
          >
            {message.content}
          </ReactMarkdown>
        )}
        {message.confidenceScore !== undefined && (
          <div className="confidence-score">
            <FiCheckCircle size={14} />
            <span>Confidence: {(message.confidenceScore * 100).toFixed(0)}%</span>
          </div>
        )}
      </div>
    </div>
  );
}

export default Message;
