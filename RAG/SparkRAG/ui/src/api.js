import axios from 'axios';

const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080/api';

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

export const chat = async (message, conversationId = null, history = []) => {
  try {
    const response = await api.post('/chat', {
      message,
      conversation_id: conversationId,
      history,
    });
    return response.data;
  } catch (error) {
    console.error('Chat error:', error);
    throw error;
  }
};

export const search = async (query, topK = 5) => {
  try {
    const response = await api.post('/search', {
      query,
      top_k: topK,
    });
    return response.data;
  } catch (error) {
    console.error('Search error:', error);
    throw error;
  }
};

export const health = async () => {
  try {
    const response = await api.get('/health');
    return response.data;
  } catch (error) {
    console.error('Health check error:', error);
    throw error;
  }
};

export const getProviders = async () => {
  try {
    const response = await api.get('/providers');
    return response.data;
  } catch (error) {
    console.error('Providers error:', error);
    throw error;
  }
};

export const getCollections = async () => {
  try {
    const response = await api.get('/collections');
    return response.data;
  } catch (error) {
    console.error('Collections error:', error);
    throw error;
  }
};
