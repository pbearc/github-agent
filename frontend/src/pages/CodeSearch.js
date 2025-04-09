import React, { useState, useEffect } from "react";
import { useLocation } from "react-router-dom";
import Card from "../components/ui/Card";
import Button from "../components/ui/Button";
import Input from "../components/ui/Input";
import CodeEditor from "../components/ui/CodeEditor";
import { toast } from "sonner";
import { motion } from "framer-motion";
import { repositoryService, navigatorService } from "../services/api";
import DOMPurify from "dompurify";
import { marked } from "marked";

const CodeSearch = () => {
  const location = useLocation();
  const [url, setUrl] = useState(location.state?.url || "");
  const [branch, setBranch] = useState("");
  const [question, setQuestion] = useState("");
  const [loading, setLoading] = useState(false);
  const [indexing, setIndexing] = useState(false);
  const [searchResults, setSearchResults] = useState(null);
  const [repoInfo, setRepoInfo] = useState(null);
  const [selectedResult, setSelectedResult] = useState(null);
  const [errors, setErrors] = useState({});
  const [isIndexed, setIsIndexed] = useState(false);
  const [showSuggestions, setShowSuggestions] = useState(false);

  useEffect(() => {
    // If URL is passed from another page, set it
    if (location.state?.url) {
      setUrl(location.state.url);
      fetchRepositoryInfo(location.state.url);
    }
  }, [location.state]);

  const fetchRepositoryInfo = async (repoUrl) => {
    if (!repoUrl) return;

    try {
      const response = await repositoryService.getInfo(repoUrl);
      setRepoInfo(response.data);
      setBranch(response.data.default_branch);

      // Add to recent repositories in localStorage
      const recentRepos = JSON.parse(
        localStorage.getItem("recentRepositories") || "[]"
      );

      // Check if repo already exists and remove it
      const filteredRepos = recentRepos.filter((repo) => repo.url !== repoUrl);

      // Add the repo to the beginning of the array
      filteredRepos.unshift({
        url: repoUrl,
        name: response.data.name,
        owner: response.data.owner,
      });

      // Keep only the 5 most recent repos
      const updatedRepos = filteredRepos.slice(0, 5);

      localStorage.setItem("recentRepositories", JSON.stringify(updatedRepos));
    } catch (error) {
      console.error("Error fetching repository info:", error);
      toast.error(
        "Could not fetch repository information. Please check the URL."
      );
    }
  };

  const handleIndex = async () => {
    if (!url) {
      setErrors({
        url: !url ? "Please enter a GitHub repository URL" : "",
      });
      return;
    }

    setIndexing(true);
    setErrors({});

    try {
      const branchToUse =
        branch || (repoInfo ? repoInfo.default_branch : "main");

      // Use navigatorService to index the repository
      const response = await navigatorService.index(url, branchToUse);

      toast.success("Repository indexed successfully");
      setIsIndexed(true);

      // Fetch repo info if not already loaded
      if (!repoInfo) {
        await fetchRepositoryInfo(url);
      }
    } catch (error) {
      console.error("Error indexing repository:", error);

      let errorMessage = "Failed to index repository";
      if (error.response?.data?.error) {
        errorMessage = error.response.data.error;
      }

      toast.error(errorMessage);
    } finally {
      setIndexing(false);
    }
  };

  const handleSearch = async (e) => {
    if (e) e.preventDefault();

    if (!url || !question) {
      setErrors({
        url: !url ? "Please enter a GitHub repository URL" : "",
        question: !question ? "Please enter a question" : "",
      });
      return;
    }

    setLoading(true);
    setErrors({});
    setSearchResults(null);
    setSelectedResult(null);

    try {
      const branchToUse =
        branch || (repoInfo ? repoInfo.default_branch : "main");

      // Show indexing toast
      toast.info("Analyzing codebase...");

      // Use navigatorService to ask a question
      const response = await navigatorService.question(
        url,
        question,
        branchToUse,
        5
      );

      setSearchResults(response.data);
      setShowSuggestions(false);

      if (
        response.data.relevant_files &&
        response.data.relevant_files.length > 0
      ) {
        // Select the first result by default
        setSelectedResult(response.data.relevant_files[0]);
        toast.success(
          `Found ${response.data.relevant_files.length} relevant ${
            response.data.relevant_files.length === 1 ? "file" : "files"
          }`
        );
      } else {
        toast.info("No relevant files found for your question");
      }

      // Fetch repo info if not already loaded
      if (!repoInfo) {
        await fetchRepositoryInfo(url);
      }
    } catch (error) {
      console.error("Error searching codebase:", error);

      let errorMessage = "Failed to search codebase";
      if (error.response?.data?.error) {
        errorMessage = error.response.data.error;
      }

      toast.error(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  const detectLanguage = (filePath) => {
    const fileExt = filePath.split(".").pop().toLowerCase();

    const extMap = {
      "js": "javascript",
      "jsx": "jsx",
      "ts": "typescript",
      "tsx": "typescript",
      "py": "python",
      "rb": "ruby",
      "php": "php",
      "java": "java",
      "go": "go",
      "c": "c",
      "cpp": "cpp",
      "cs": "csharp",
      "html": "html",
      "css": "css",
      "json": "json",
      "md": "markdown",
      "sh": "bash",
      "rs": "rust",
      "swift": "swift",
      "kt": "kotlin",
    };

    return extMap[fileExt] || "text";
  };

  // Sample questions
  const sampleQuestions = [
    "How is authentication implemented?",
    "What's the main entry point of the application?",
    "How does data flow through the codebase?",
    "Explain the directory structure of this project",
    "What design patterns are used in this codebase?",
    "How are API calls handled?",
    "How does the error handling work?",
    "What database models exist in this project?",
    "How is state management implemented?",
    "How does the routing system work?",
  ];

  return (
    <div>
      <motion.div
        initial={{ opacity: 0, y: -20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.3 }}
        className="mb-8"
      >
        <div className="flex flex-col md:flex-row md:items-center md:justify-between">
          <div>
            <h1 className="text-3xl font-bold bg-clip-text text-transparent bg-gradient-to-r from-primary-500 to-secondary-500 mb-2">
              Code Q&A
            </h1>
            <p className="text-gray-600 dark:text-gray-300">
              Ask questions about any GitHub repository and get AI-powered
              answers
            </p>
          </div>
          <div className="mt-4 md:mt-0">
            <div className="text-sm inline-flex items-center rounded-full bg-warning-50 dark:bg-warning-900/20 px-3 py-1 text-warning-700 dark:text-warning-300 border border-warning-200 dark:border-warning-800">
              <svg
                xmlns="http://www.w3.org/2000/svg"
                className="h-4 w-4 mr-1"
                viewBox="0 0 20 20"
                fill="currentColor"
              >
                <path
                  fillRule="evenodd"
                  d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z"
                  clipRule="evenodd"
                />
              </svg>
              <span>
                Ask questions in natural language to understand any codebase
              </span>
            </div>
          </div>
        </div>
      </motion.div>

      <div className="mb-6">
        <Card className="bg-gradient-to-b from-dark-100/70 to-dark-100 border-none shadow-glow">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div className="md:col-span-2">
              <Input
                label="GitHub Repository URL"
                placeholder="https://github.com/username/repository"
                value={url}
                onChange={(e) => setUrl(e.target.value)}
                error={errors.url}
                required
                dark
              />
            </div>
            <div>
              <Input
                label="Branch (Optional)"
                placeholder="main"
                value={branch}
                onChange={(e) => setBranch(e.target.value)}
                dark
              />
            </div>
          </div>

          <div className="mt-4 flex flex-col sm:flex-row gap-4 items-start sm:items-end">
            <Button
              type="button"
              variant="outline"
              onClick={handleIndex}
              isLoading={indexing}
              disabled={indexing}
              className="border-secondary-500 text-secondary-400 hover:bg-secondary-500/10"
            >
              {indexing ? "Indexing..." : "Index Repository"}
            </Button>
            <div className="text-sm text-gray-400 dark:text-gray-400 flex items-center">
              <svg
                xmlns="http://www.w3.org/2000/svg"
                className="h-5 w-5 mr-2 text-secondary-400"
                viewBox="0 0 20 20"
                fill="currentColor"
              >
                <path
                  fillRule="evenodd"
                  d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z"
                  clipRule="evenodd"
                />
              </svg>
              Indexing allows for more accurate code analysis
            </div>
          </div>

          <div className="mt-6">
            <form onSubmit={handleSearch}>
              <div className="relative">
                <Input
                  label="Ask a Question About the Codebase"
                  placeholder="How is authentication implemented?"
                  value={question}
                  onChange={(e) => {
                    setQuestion(e.target.value);
                    // Show suggestions when user starts typing
                    if (e.target.value && !showSuggestions) {
                      setShowSuggestions(true);
                    } else if (!e.target.value) {
                      setShowSuggestions(false);
                    }
                  }}
                  error={errors.question}
                  onFocus={() => question && setShowSuggestions(true)}
                  onBlur={() =>
                    setTimeout(() => setShowSuggestions(false), 200)
                  }
                  required
                  dark
                />

                {showSuggestions && (
                  <div className="absolute z-10 w-full mt-1 bg-dark-100 border border-gray-700 rounded-md shadow-lg overflow-hidden">
                    <div className="max-h-60 overflow-y-auto">
                      {sampleQuestions
                        .filter(
                          (q) =>
                            q.toLowerCase().includes(question.toLowerCase()) ||
                            question === ""
                        )
                        .map((q, idx) => (
                          <div
                            key={idx}
                            className="px-4 py-2 hover:bg-dark-200 cursor-pointer text-gray-200"
                            onClick={() => {
                              setQuestion(q);
                              setShowSuggestions(false);
                            }}
                          >
                            {q}
                          </div>
                        ))}
                    </div>
                  </div>
                )}
              </div>

              <div className="mt-6 flex flex-col sm:flex-row gap-4">
                <Button
                  type="submit"
                  isLoading={loading}
                  disabled={loading}
                  className="bg-gradient-to-r from-primary-600 to-secondary-600 hover:from-primary-500 hover:to-secondary-500 shadow-glow"
                >
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    className="h-5 w-5 mr-2"
                    viewBox="0 0 20 20"
                    fill="currentColor"
                  >
                    <path
                      fillRule="evenodd"
                      d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-8-3a1 1 0 00-.867.5 1 1 0 11-1.731-1A3 3 0 0113 8a3.001 3.001 0 01-2 2.83V11a1 1 0 11-2 0v-1a1 1 0 011-1 1 1 0 100-2zm0 8a1 1 0 100-2 1 1 0 000 2z"
                      clipRule="evenodd"
                    />
                  </svg>
                  Ask Question
                </Button>

                <Button
                  type="button"
                  variant="outline"
                  onClick={() => setShowSuggestions(!showSuggestions)}
                  className="border-accent-500 text-accent-400 hover:bg-accent-500/10"
                >
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    className="h-5 w-5 mr-2"
                    viewBox="0 0 20 20"
                    fill="currentColor"
                  >
                    <path
                      fillRule="evenodd"
                      d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-8-3a1 1 0 00-.867.5 1 1 0 11-1.731-1A3 3 0 0113 8a3.001 3.001 0 01-2 2.83V11a1 1 0 11-2 0v-1a1 1 0 011-1 1 1 0 100-2zm0 8a1 1 0 100-2 1 1 0 000 2z"
                      clipRule="evenodd"
                    />
                  </svg>
                  Suggested Questions
                </Button>
              </div>
            </form>
          </div>
        </Card>
      </div>

      {searchResults && (
        <div className="space-y-6">
          {/* Answer first */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5 }}
          >
            <Card
              title="Understanding the Code"
              className="bg-gradient-to-b from-dark-100/70 to-dark-100 border-none shadow-glow"
            >
              <div className="prose dark:prose-invert max-w-none text-gray-200">
                <div
                  dangerouslySetInnerHTML={{
                    __html: DOMPurify.sanitize(
                      marked.parse(searchResults.answer)
                    ),
                  }}
                />
              </div>
            </Card>
          </motion.div>

          {/* Relevant files and code snippets */}
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            <motion.div
              initial={{ opacity: 0, x: -20 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ duration: 0.5, delay: 0.1 }}
            >
              <Card
                title={`Relevant Files (${
                  searchResults.relevant_files?.length || 0
                })`}
                noPadding
                className="h-full bg-gradient-to-b from-dark-100/70 to-dark-100 border-none shadow-glow"
              >
                <div className="h-full flex flex-col">
                  <div className="p-4 border-b border-gray-800">
                    <p className="text-sm text-gray-400">
                      Question:{" "}
                      <span className="font-medium text-white">{question}</span>
                    </p>
                  </div>

                  <div
                    className="flex-1 overflow-y-auto scrollbar-thin scrollbar-thumb-gray-700 scrollbar-track-transparent"
                    style={{ maxHeight: "500px" }}
                  >
                    {searchResults.relevant_files &&
                    searchResults.relevant_files.length > 0 ? (
                      <ul className="divide-y divide-gray-800">
                        {searchResults.relevant_files.map((file, index) => (
                          <li
                            key={index}
                            className={`
                              p-4 hover:bg-dark-200 cursor-pointer transition-colors duration-150
                              ${
                                selectedResult &&
                                selectedResult.path === file.path
                                  ? "bg-secondary-900/20 border-l-2 border-secondary-500"
                                  : ""
                              }
                            `}
                            onClick={() => setSelectedResult(file)}
                          >
                            <h3 className="text-sm font-medium text-gray-200 truncate">
                              {file.path.split("/").pop()}
                            </h3>
                            <p className="mt-1 text-xs text-gray-500 truncate">
                              {file.path}
                            </p>
                            <div className="mt-1 text-xs">
                              <span
                                className={`inline-block px-2 py-1 rounded-full text-xs
                                  ${
                                    file.relevance > 0.8
                                      ? "bg-accent-900/30 text-accent-300"
                                      : file.relevance > 0.5
                                      ? "bg-secondary-900/30 text-secondary-300"
                                      : "bg-primary-900/30 text-primary-300"
                                  }
                                `}
                              >
                                Relevance: {(file.relevance * 100).toFixed(1)}%
                              </span>
                            </div>
                          </li>
                        ))}
                      </ul>
                    ) : (
                      <div className="p-6 text-center text-gray-500">
                        No relevant files found for your question.
                      </div>
                    )}
                  </div>
                </div>
              </Card>
            </motion.div>

            <motion.div
              initial={{ opacity: 0, x: 20 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ duration: 0.5, delay: 0.2 }}
              className="lg:col-span-2"
            >
              {selectedResult ? (
                <Card
                  title={selectedResult.path.split("/").pop()}
                  subtitle={selectedResult.path}
                  className="bg-gradient-to-b from-dark-100/70 to-dark-100 border-none shadow-glow"
                >
                  <CodeEditor
                    code={selectedResult.snippet}
                    language={detectLanguage(selectedResult.path)}
                    readOnly={true}
                    lineNumbers={true}
                    showCopyButton={true}
                    style={{ minHeight: "500px" }}
                  />
                </Card>
              ) : (
                <Card className="bg-gradient-to-b from-dark-100/70 to-dark-100 border-none shadow-glow">
                  <div className="text-center py-12 text-gray-500">
                    <svg
                      xmlns="http://www.w3.org/2000/svg"
                      className="mx-auto h-12 w-12 mb-4 text-gray-400"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={1}
                        d="M8 16l2.879-2.879m0 0a3 3 0 104.243-4.242 3 3 0 00-4.243 4.242zM21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                      />
                    </svg>
                    <p className="text-lg font-medium">
                      Select a file from the list to view the code
                    </p>
                    <p className="mt-2 text-sm text-gray-600">
                      The AI has identified these files as most relevant to your
                      question
                    </p>
                  </div>
                </Card>
              )}
            </motion.div>
          </div>
        </div>
      )}

      {!searchResults && !loading && (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ duration: 0.5, delay: 0.3 }}
          className="text-center py-12"
        >
          <div className="bg-dark-100 rounded-2xl p-8 max-w-2xl mx-auto border border-gray-800">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              className="mx-auto h-16 w-16 mb-4 text-gray-600"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={1}
                d="M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
            <h2 className="text-2xl font-bold text-white mb-3">
              Ask Me Anything About the Code
            </h2>
            <p className="text-gray-400 mb-6 max-w-md mx-auto">
              Enter a GitHub URL, then ask questions to understand the codebase
              - from architecture to specific functionality.
            </p>
            <div className="flex flex-col gap-4 mt-6">
              <p className="text-sm font-medium text-gray-400">
                Example Questions:
              </p>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                {sampleQuestions.slice(0, 4).map((q, idx) => (
                  <div
                    key={idx}
                    className="text-left p-3 bg-dark-200 rounded-lg cursor-pointer hover:bg-dark-300 border border-gray-800 text-gray-300"
                    onClick={() => {
                      setQuestion(q);
                      // Auto-scroll to question input
                      document
                        .querySelector(
                          'input[placeholder="How is authentication implemented?"]'
                        )
                        ?.scrollIntoView({
                          behavior: "smooth",
                          block: "center",
                        });
                    }}
                  >
                    {q}
                  </div>
                ))}
              </div>
            </div>
          </div>
        </motion.div>
      )}
    </div>
  );
};

export default CodeSearch;
