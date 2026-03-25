import axios from "axios";

const instance = axios.create({
	baseURL: 'http://localhost:3000', // Assicurati che sia l'URL giusto del tuo backend
	timeout: 1000 * 5
});

// Questo intercettore aggiunge il "Bearer" a ogni singola chiamata
instance.interceptors.request.use((config) => {
	const token = localStorage.getItem('token');
	if (token) {
		config.headers.Authorization = `Bearer ${token}`;
	}
	return config;
}, (error) => {
	return Promise.reject(error);
});

export default instance;