import React, { useState, useEffect } from "react";
import { useLocation } from "react-router-dom";
import Card from "../components/ui/Card";
import Button from "../components/ui/Button";
import Input from "../components/ui/Input";
import CodeEditor from "../components/ui/CodeEditor";
import { toast } from "sonner";
import { motion } from "framer-motion";
import { searchService, repositoryService } from "../services/api";

const CodeSearch = () => {
  const location = useLocation();
  const [url, setUrl] = useState(location.state?.url || "");
  const [branch, setBranch] = useState("");
  const [query, setQuery] = useState("");
  const [loading, setLoading] = useState(false);
  const [searchResults, setSearchResults] = useState(null);
  const [repoInfo, setRepoInfo] = useState(null);
  const [selectedResult, setSelectedResult] = useState(null);
  const [errors, setErrors] = useState({});

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

  const handleSearch = async (e) => {
    e.preventDefault();

    if (!url || !query) {
      setErrors({
        url: !url ? "Please enter a GitHub repository URL" : "",
        query: !query ? "Please enter a search query" : "",
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
      const response = await searchService.code(url, query, branchToUse);
      setSearchResults(response.data);

      if (response.data.results.length === 0) {
        toast.info("No results found for your query");
      } else {
        toast.success(
          `Found ${response.data.results.length} result${
            response.data.results.length === 1 ? "" : "s"
          }`
        );
        // Select the first result by default
        setSelectedResult(response.data.results[0]);
      }

      // Fetch repo info if not already loaded
      if (!repoInfo) {
        await fetchRepositoryInfo(url);
      }
    } catch (error) {
      console.error("Error searching code:", error);

      let errorMessage = "Failed to search code";
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

  // Sample search queries
  const sampleQueries = [
    "function authenticate",
    "database connection",
    "api endpoint",
    "error handling",
    "TODO:",
    "security vulnerability",
    "performance issue",
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
          Search for code patterns and snippets in your GitHub repository
        </p>
      </motion.div>

      <div className="mb-6">
        <Card title="Search Parameters">
          <form onSubmit={handleSearch}>
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

            <div className="mt-4">
              <Input
                label="Search Query"
                placeholder="Enter code pattern, function name, or keyword to search for"
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                error={errors.query}
                required
              />
            </div>

            <div className="mt-4">
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                Sample Queries
              </label>
              <div className="flex flex-wrap gap-2">
                {sampleQueries.map((sample, index) => (
                  <Button
                    key={index}
                    variant="outline"
                    size="xs"
                    onClick={() => setQuery(sample)}
                  >
                    {sample}
                  </Button>
                ))}
              </div>
            </div>

            <div className="mt-6">
              <Button type="submit" isLoading={loading} disabled={loading}>
                Search Code
              </Button>
            </div>
          </form>
        </Card>
      </div>

      {searchResults && (
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Results list */}
          <div>
            <Card
              title={`Search Results (${searchResults.results.length})`}
              noPadding
              className="h-full"
            >
              <div className="h-full flex flex-col">
                <div className="p-4 border-b border-gray-200 dark:border-gray-700">
                  <p className="text-sm text-gray-600 dark:text-gray-400">
                    Search query:{" "}
                    <span className="font-medium text-gray-900 dark:text-white">
                      {searchResults.query}
                    </span>
                  </p>
                </div>

                <div
                  className="flex-1 overflow-y-auto"
                  style={{ maxHeight: "500px" }}
                >
                  {searchResults.results.length > 0 ? (
                    <ul className="divide-y divide-gray-200 dark:divide-gray-700">
                      {searchResults.results.map((result, index) => (
                        <li
                          key={index}
                          className={`
                            p-4 hover:bg-gray-50 dark:hover:bg-gray-800 cursor-pointer
                            ${
                              selectedResult &&
                              selectedResult.path === result.path
                                ? "bg-blue-50 dark:bg-blue-900/20"
                                : ""
                            }
                          `}
                          onClick={() => setSelectedResult(result)}
                        >
                          <h3 className="text-sm font-medium text-gray-900 dark:text-white truncate">
                            {result.path.split("/").pop()}
                          </h3>
                          <p className="mt-1 text-xs text-gray-500 dark:text-gray-400 truncate">
                            {result.path}
                          </p>
                        </li>
                      ))}
                    </ul>
                  ) : (
                    <div className="p-6 text-center text-gray-500 dark:text-gray-400">
                      No results found for your query.
                    </div>
                  )}
                </div>
              </div>
            </Card>
          </div>

          {/* Result content */}
          <div className="lg:col-span-2">
            {selectedResult ? (
              <Card
                title={selectedResult.path.split("/").pop()}
                subtitle={selectedResult.path}
                headerAction={
                  <a
                    href={selectedResult.html_link}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-sm text-primary-600 dark:text-primary-500 hover:underline"
                  >
                    View on GitHub
                  </a>
                }
              >
                <CodeEditor
                  code={selectedResult.content}
                  language={detectLanguage(selectedResult.path)}
                  readOnly={true}
                  lineNumbers={true}
                  showCopyButton={true}
                  style={{ minHeight: "400px" }}
                />
              </Card>
            ) : searchResults.analysis ? (
              <Card title="Analysis">
                <div className="prose dark:prose-invert max-w-none">
                  <div
                    dangerouslySetInnerHTML={{
                      __html: searchResults.analysis.replace(/\n/g, "<br />"),
                    }}
                  />
                </div>
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
                  <p>Select a result from the list to view the code</p>
                </div>
              </Card>
            )}

            {searchResults.analysis && selectedResult && (
              <Card title="AI Analysis" className="mt-6">
                <div className="prose dark:prose-invert max-w-none">
                  <div
                    dangerouslySetInnerHTML={{
                      __html: searchResults.analysis.replace(/\n/g, "<br />"),
                    }}
                  />
                </div>
              </Card>
            )}
          </div>
        </div>
      )}
    </div>
  );
};

export default CodeSearch;
