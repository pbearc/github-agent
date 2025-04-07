import React, { useState, useEffect } from "react";
import { useLocation } from "react-router-dom";
import Card from "../components/ui/Card";
import Button from "../components/ui/Button";
import Input from "../components/ui/Input";
import CodeEditor from "../components/ui/CodeEditor";
import { toast } from "sonner";
import { motion } from "framer-motion";
import { repositoryService, navigatorService } from "../services/api";
import ReactMarkdown from "react-markdown";
import { marked } from "marked";
import DOMPurify from "dompurify";

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
    } catch (error) {
      console.error("Error fetching repository info:", error);
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
    e.preventDefault();

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

      // Use navigatorService to ask a question
      const response = await navigatorService.question(
        url,
        question,
        branchToUse,
        5
      );

      setSearchResults(response.data);

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
    "Where is the database connection configured?",
    "How does the routing system work?",
    "How are errors handled?",
    "What security measures are in place?",
    "How is the application structured?",
    "Where is the main entry point?",
  ];

  return (
    <div>
      <motion.div
        initial={{ opacity: 0, y: -20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.3 }}
      >
        <h1 className="text-2xl font-bold text-gray-900 dark:text-white mb-2">
          Code Search
        </h1>
        <p className="text-gray-600 dark:text-gray-300 mb-6">
          AI-powered search for understanding any GitHub repository
        </p>
      </motion.div>

      <div className="mb-6">
        <Card title="Repository Information">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div className="md:col-span-2">
              <Input
                label="GitHub Repository URL"
                placeholder="https://github.com/username/repository"
                value={url}
                onChange={(e) => setUrl(e.target.value)}
                error={errors.url}
                required
              />
            </div>
            <div>
              <Input
                label="Branch (Optional)"
                placeholder="main"
                value={branch}
                onChange={(e) => setBranch(e.target.value)}
              />
            </div>
          </div>

          <div className="mt-4 flex gap-4">
            <Button
              type="button"
              variant="outline"
              onClick={handleIndex}
              isLoading={indexing}
              disabled={indexing}
            >
              {indexing ? "Indexing..." : "Index Repository"}
            </Button>
            <div className="text-sm text-gray-500 dark:text-gray-400 flex items-center">
              Index the repository to enable AI code navigation
            </div>
          </div>

          {isIndexed && (
            <>
              <div className="border-t border-gray-200 dark:border-gray-700 my-6"></div>

              <form onSubmit={handleSearch}>
                <div>
                  <Input
                    label="Ask a Question About the Codebase"
                    placeholder="How is authentication implemented?"
                    value={question}
                    onChange={(e) => setQuestion(e.target.value)}
                    error={errors.question}
                    required
                  />
                </div>

                <div className="mt-4">
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    Sample Questions
                  </label>
                  <div className="flex flex-wrap gap-2">
                    {sampleQuestions.map((sample, index) => (
                      <Button
                        key={index}
                        variant="outline"
                        size="xs"
                        onClick={() => setQuestion(sample)}
                      >
                        {sample}
                      </Button>
                    ))}
                  </div>
                </div>

                <div className="mt-6">
                  <Button type="submit" isLoading={loading} disabled={loading}>
                    Ask Question
                  </Button>
                </div>
              </form>
            </>
          )}
        </Card>
      </div>

      {searchResults && (
        <div className="space-y-6">
          {/* Answer first */}
          <Card title="Answer">
            <div className="prose dark:prose-invert max-w-none text-gray-800 dark:text-gray-200">
              {/* If using dangerouslySetInnerHTML approach */}
              {/* <div
                dangerouslySetInnerHTML={{
                  __html: searchResults.answer
                    .replace(/\n/g, "<br />")
                    .replace(/`{3}([^`]+)`{3}/g, "<pre><code>$1</code></pre>")
                    .replace(/`([^`]+)`/g, "<code>$1</code>")
                    .replace(/#{3}([^#\n]+)/g, "<h3>$1</h3>")
                    .replace(/#{2}([^#\n]+)/g, "<h2>$1</h2>")
                    .replace(/#{1}([^#\n]+)/g, "<h1>$1</h1>"),
                }}
              /> */}

              {/* Or if using marked library */}

              <div
                dangerouslySetInnerHTML={{
                  __html: DOMPurify.sanitize(
                    marked.parse(searchResults.answer)
                  ),
                }}
              />
            </div>
          </Card>

          {/* Relevant files and code snippets */}
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            <div>
              <Card
                title={`Relevant Files (${
                  searchResults.relevant_files?.length || 0
                })`}
                noPadding
                className="h-full"
              >
                <div className="h-full flex flex-col">
                  <div className="p-4 border-b border-gray-200 dark:border-gray-700">
                    <p className="text-sm text-gray-600 dark:text-gray-400">
                      Question:{" "}
                      <span className="font-medium text-gray-900 dark:text-white">
                        {question}
                      </span>
                    </p>
                  </div>

                  <div
                    className="flex-1 overflow-y-auto"
                    style={{ maxHeight: "500px" }}
                  >
                    {searchResults.relevant_files &&
                    searchResults.relevant_files.length > 0 ? (
                      <ul className="divide-y divide-gray-200 dark:divide-gray-700">
                        {searchResults.relevant_files.map((file, index) => (
                          <li
                            key={index}
                            className={`
                              p-4 hover:bg-gray-50 dark:hover:bg-gray-800 cursor-pointer
                              ${
                                selectedResult &&
                                selectedResult.path === file.path
                                  ? "bg-blue-50 dark:bg-blue-900/20"
                                  : ""
                              }
                            `}
                            onClick={() => setSelectedResult(file)}
                          >
                            <h3 className="text-sm font-medium text-gray-900 dark:text-white truncate">
                              {file.path.split("/").pop()}
                            </h3>
                            <p className="mt-1 text-xs text-gray-500 dark:text-gray-400 truncate">
                              {file.path}
                            </p>
                            <div className="mt-1 text-xs text-blue-600 dark:text-blue-400">
                              Relevance: {(file.relevance * 100).toFixed(1)}%
                            </div>
                          </li>
                        ))}
                      </ul>
                    ) : (
                      <div className="p-6 text-center text-gray-500 dark:text-gray-400">
                        No relevant files found for your question.
                      </div>
                    )}
                  </div>
                </div>
              </Card>
            </div>

            <div className="lg:col-span-2">
              {selectedResult ? (
                <Card
                  title={selectedResult.path.split("/").pop()}
                  subtitle={selectedResult.path}
                >
                  <CodeEditor
                    code={selectedResult.snippet}
                    language={detectLanguage(selectedResult.path)}
                    readOnly={true}
                    lineNumbers={true}
                    showCopyButton={true}
                    style={{ minHeight: "400px" }}
                  />
                </Card>
              ) : (
                <Card>
                  <div className="text-center py-12 text-gray-500 dark:text-gray-400">
                    <svg
                      xmlns="http://www.w3.org/2000/svg"
                      className="mx-auto h-12 w-12 mb-4"
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
                    <p>Select a file from the list to view the code</p>
                  </div>
                </Card>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default CodeSearch;
