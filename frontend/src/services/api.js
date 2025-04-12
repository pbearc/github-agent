import axios from "axios";

// Create an axios instance with baseURL
const api = axios.create({
  baseURL:
    process.env.REACT_APP_API_URL ||
    "http://github-agent-load-balancer-1708352436.us-east-1.elb.amazonaws.com:8080/api",
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

const navigatorService = {
  index: (url, branch = "") => api.post("/navigate/index", { url, branch }),

  question: (url, question, branch = "", topK = 5) =>
    api.post("/navigate/question", { url, question, branch, top_k: topK }),

  // Add smart navigate endpoint
  smartNavigate: (url, question, branch = "") =>
    api.post("/smart-navigate", { url, question, branch }),

  // Add these new methods
  visualizeArchitecture: (url, branch, detailLevel, focusPaths) =>
    api.post("/navigate/architecture", {
      url,
      branch,
      detail: detailLevel, // Note: parameter name is "detail" not "detail_level"
      focus_paths: focusPaths,
    }),

  generateWalkthrough: (
    url,
    branch,
    depth = 3,
    entryPoint = "",
    focusPaths = []
  ) =>
    api.post("/navigate/walkthrough", {
      url,
      branch,
      depth,
      entry_point: entryPoint,
      focus_paths: focusPaths,
    }),

  getBestPractices: (url, branch, scope, scopePath) =>
    api.post("/navigate/practices", {
      url,
      branch,
      scope,
      path: scopePath, // Note: parameter name is "path" not "scope_path"
    }),

  getArchitectureGraph: (url, branch) =>
    api.post("/navigate/architecture-graph", { url, branch }),

  getArchitectureGraphAndExplanation: (url, branch) =>
    api.post("/navigate/explain-architecture", { url, branch }),
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

const prSummaryService = {
  getSummary: (url) => api.post("/pr/summary", { url }),
};

export {
  api,
  repositoryService,
  generateService,
  searchService,
  pushService,
  llmService,
  healthService,
  navigatorService,
  prSummaryService,
};
