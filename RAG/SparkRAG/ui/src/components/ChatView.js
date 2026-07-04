import React, { useState, useRef, useEffect } from 'react';
import { chat } from '../api';
import MessageComponent from './Message';
import CitationComponent from './Citation';
import './ChatView.css';
import { FiSend, FiX } from 'react-icons/fi';

function ChatView() {
  const [messages, setMessages] = useState([
    {
      id: 'initial',
      type: 'assistant',
      content: 'Hello! I\'m Spark RAG, your Apache Spark knowledge assistant. Ask me anything about Spark, Hadoop, Delta Lake, and more!',
      citations: [],
    },
  ]);
  const [input, setInput] = useState('');
  const [loading, setLoading] = useState(false);
  const [conversationId] = useState(null);
  const [error, setError] = useState(null);
  const messagesEndRef = useRef(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!input.trim() || loading) return;

    const userMessage = {
      id: Date.now().toString(),
      type: 'user',
      content: input,
    };

    setMessages((prev) => [...prev, userMessage]);
    setInput('');
    setLoading(true);
    setError(null);

    try {
      const history = messages
        .filter((m) => m.type !== 'system')
        .map((m) => ({
          role: m.type === 'user' ? 'user' : 'assistant',
          content: m.content,
        }));

      const response = await chat(input, conversationId, history);

      const assistantMessage = {
        id: response.conversation_id,
        type: 'assistant',
        content: response.answer,
        citations: response.citations || [],
        confidenceScore: response.confidence_score,
      };

      setMessages((prev) => [...prev, assistantMessage]);
    } catch (err) {
      setError(err.message || 'Failed to get response');
      console.error('Chat error:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleClearChat = () => {
    setMessages([
      {
        id: 'initial',
        type: 'assistant',
        content: 'Chat cleared. What would you like to know?',
        citations: [],
      },
    ]);
    setError(null);
  };

  return (
    <div className="chat-view">
      <div className="chat-header">
        <h2>Spark RAG Chat</h2>
        <button className="clear-btn" onClick={handleClearChat} title="Clear chat">
          <FiX size={18} />
        </button>
      </div>

      <div className="messages-container">
        {messages.map((msg) => (
          <div key={msg.id} className={`message-wrapper message-${msg.type}`}>
            <MessageComponent message={msg} />
            {msg.citations && msg.citations.length > 0 && (
              <div className="citations-section">
                <h4>Sources</h4>
                {msg.citations.map((citation, idx) => (
                  <CitationComponent key={idx} citation={citation} />
                ))}
              </div>
            )}
          </div>
        ))}
        {loading && (
          <div className="message-wrapper message-loading">
            <div className="typing-indicator">
              <span></span>
              <span></span>
              <span></span>
            </div>
          </div>
        )}
        <div ref={messagesEndRef} />
      </div>

      {error && (
        <div className="error-banner">
          <p>{error}</p>
          <button onClick={() => setError(null)}>Dismiss</button>
        </div>
      )}

      <form className="input-form" onSubmit={handleSubmit}>
        <input
          type="text"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          placeholder="Ask about Apache Spark, Hadoop, Delta Lake, Iceberg..."
          disabled={loading}
          className="input-field"
        />
        <button type="submit" disabled={loading || !input.trim()} className="send-btn">
          <FiSend size={18} />
        </button>
      </form>
    </div>
  );
}

export default ChatView;
