import axios from "axios";

// Create an axios instance with baseURL
const api = axios.create({
  baseURL: process.env.REACT_APP_API_URL || "http://localhost:8080/api",
  headers: {
    "Content-Type": "application/json",
  },
});

// Add interceptors for error handling
api.interceptors.response.use(
  (response) => response,
  (error) => {
    // Handle common error cases
    if (error.response) {
      // Server responded with a status code outside of 2xx
      console.error("API Error:", error.response.data);
    } else if (error.request) {
      // The request was made but no response was received
      console.error("Network Error:", error.request);
    } else {
      // Something happened in setting up the request
      console.error("Request Error:", error.message);
    }
    return Promise.reject(error);
  }
);

// Repository endpoints
const repositoryService = {
  getInfo: (url, branch = "") => api.post("/repo/info", { url, branch }),
  getFile: (url, path, branch = "") =>
    api.post("/repo/file", { url, path, branch }),
  listFiles: (url, path = "", branch = "") =>
    api.post("/repo/files", { url, path, branch }),
};

// Generate endpoints
const generateService = {
  readme: (url, includeFiles = true, branch = "") =>
    api.post("/generate/readme", { url, include_files: includeFiles, branch }),

  dockerfile: (url, language = "", branch = "") =>
    api.post("/generate/dockerfile", { url, language, branch }),

  comments: (url, filePath, branch = "") =>
    api.post("/generate/comments", { url, file_path: filePath, branch }),

  refactor: (url, filePath, instructions, branch = "") =>
    api.post("/generate/refactor", {
      url,
      file_path: filePath,
      instructions,
      branch,
    }),
};

// Search endpoints
const searchService = {
  code: (url, query, branch = "") =>
    api.post("/search/code", { url, query, branch }),
};

// Push endpoints
const pushService = {
  file: (url, path, content, message = "", branch = "") =>
    api.post("/push/file", { url, path, content, message, branch }),
};

// LLM operations
const llmService = {
  operation: (operation) => api.post("/llm/operation", { operation }),
};

// Health check
const healthService = {
  check: () => axios.get("/health"),
};

export {
  api,
  repositoryService,
  generateService,
  searchService,
  pushService,
  llmService,
  healthService,
};
