import React, { useState, useEffect, useRef } from 'react';
import './App.css';
import ChatView from './components/ChatView';
import SearchView from './components/SearchView';
import SettingsView from './components/SettingsView';
import { health } from './api';
import { FiMessageSquare, FiSearch, FiSettings, FiMoon, FiSun } from 'react-icons/fi';

function App() {
  const [activeTab, setActiveTab] = useState('chat');
  const [darkMode, setDarkMode] = useState(false);
  const [apiHealth, setApiHealth] = useState(null);
  const healthCheckInterval = useRef(null);

  useEffect(() => {
    document.documentElement.classList.toggle('dark-mode', darkMode);
    localStorage.setItem('darkMode', darkMode);
  }, [darkMode]);

  useEffect(() => {
    const savedDarkMode = localStorage.getItem('darkMode') === 'true';
    setDarkMode(savedDarkMode);
  }, []);

  useEffect(() => {
    const checkHealth = async () => {
      try {
        const data = await health();
        setApiHealth(data);
      } catch (error) {
        setApiHealth({ status: 'error', error: error.message });
      }
    };

    checkHealth();
    healthCheckInterval.current = setInterval(checkHealth, 30000);

    return () => clearInterval(healthCheckInterval.current);
  }, []);

  return (
    <div className="app-container">
      <header className="app-header">
        <div className="header-content">
          <div className="logo-section">
            <h1>Spark RAG</h1>
            <p>Enterprise Knowledge Assistant</p>
          </div>
          <div className="header-controls">
            <button
              className="theme-toggle"
              onClick={() => setDarkMode(!darkMode)}
              title="Toggle dark mode"
            >
              {darkMode ? <FiSun size={20} /> : <FiMoon size={20} />}
            </button>
            {apiHealth && (
              <div className={`health-indicator ${apiHealth.status}`}>
                <span className="status-dot"></span>
                <span className="status-text">
                  {apiHealth.status === 'healthy' ? 'Online' : 'Offline'}
                </span>
              </div>
            )}
          </div>
        </div>
      </header>

      <div className="app-main">
        <nav className="app-nav">
          <button
            className={`nav-item ${activeTab === 'chat' ? 'active' : ''}`}
            onClick={() => setActiveTab('chat')}
          >
            <FiMessageSquare size={18} />
            <span>Chat</span>
          </button>
          <button
            className={`nav-item ${activeTab === 'search' ? 'active' : ''}`}
            onClick={() => setActiveTab('search')}
          >
            <FiSearch size={18} />
            <span>Search</span>
          </button>
          <button
            className={`nav-item ${activeTab === 'settings' ? 'active' : ''}`}
            onClick={() => setActiveTab('settings')}
          >
            <FiSettings size={18} />
            <span>Settings</span>
          </button>
        </nav>

        <main className="app-content">
          {activeTab === 'chat' && <ChatView />}
          {activeTab === 'search' && <SearchView />}
          {activeTab === 'settings' && <SettingsView />}
        </main>
      </div>

      <footer className="app-footer">
        <p>Spark RAG Platform v1.0.0 | Apache Spark Knowledge Assistant</p>
      </footer>
    </div>
  );
}

export default App;
