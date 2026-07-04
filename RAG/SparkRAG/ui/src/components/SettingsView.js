import React, { useState, useEffect } from 'react';
import { getProviders, getCollections } from '../api';
import './SettingsView.css';
import { FiLoader } from 'react-icons/fi';

function SettingsView() {
  const [providers, setProviders] = useState(null);
  const [collections, setCollections] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    const loadSettings = async () => {
      try {
        const [providersData, collectionsData] = await Promise.all([
          getProviders(),
          getCollections(),
        ]);
        setProviders(providersData);
        setCollections(collectionsData);
      } catch (err) {
        setError(err.message || 'Failed to load settings');
      } finally {
        setLoading(false);
      }
    };

    loadSettings();
  }, []);

  if (loading) {
    return (
      <div className="settings-view loading">
        <FiLoader className="spinning" size={32} />
        <p>Loading settings...</p>
      </div>
    );
  }

  return (
    <div className="settings-view">
      <h2>System Settings</h2>

      {error && (
        <div className="settings-error">
          <p>{error}</p>
        </div>
      )}

      {providers && (
        <section className="settings-section">
          <h3>Providers</h3>
          <div className="provider-grid">
            <div className="provider-card">
              <h4>Embedding Providers</h4>
              <ul>
                {providers.embedding_providers?.map((p) => (
                  <li key={p}>{p}</li>
                ))}
              </ul>
            </div>
            <div className="provider-card">
              <h4>LLM Providers</h4>
              <ul>
                {providers.llm_providers?.map((p) => (
                  <li key={p}>{p}</li>
                ))}
              </ul>
            </div>
            <div className="provider-card">
              <h4>Vector DB Providers</h4>
              <ul>
                {providers.vectordb_providers?.map((p) => (
                  <li key={p}>{p}</li>
                ))}
              </ul>
            </div>
          </div>
        </section>
      )}

      {collections && (
        <section className="settings-section">
          <h3>Collections</h3>
          <div className="collections-info">
            <p><strong>Active Collection:</strong> {collections.active}</p>
            <p><strong>Total Collections:</strong> {collections.collections?.length || 0}</p>
            {collections.collections && collections.collections.length > 0 && (
              <div>
                <p><strong>Available Collections:</strong></p>
                <ul className="collections-list">
                  {collections.collections.map((c) => (
                    <li key={c} className={c === collections.active ? 'active' : ''}>
                      {c}
                    </li>
                  ))}
                </ul>
              </div>
            )}
          </div>
        </section>
      )}

      <section className="settings-section">
        <h3>Application Info</h3>
        <div className="info-box">
          <p><strong>Application:</strong> Spark RAG Platform</p>
          <p><strong>Version:</strong> 1.0.0</p>
          <p><strong>Mode:</strong> Production</p>
          <p><strong>API Endpoint:</strong> {process.env.REACT_APP_API_URL || 'http://localhost:8080/api'}</p>
        </div>
      </section>
    </div>
  );
}

export default SettingsView;
