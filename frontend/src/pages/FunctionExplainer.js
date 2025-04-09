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

const FunctionExplainer = () => {
  const location = useLocation();
  const [url, setUrl] = useState(location.state?.url || "");
  const [branch, setBranch] = useState("");
  const [filePath, setFilePath] = useState("");
  const [functionName, setFunctionName] = useState("");
  const [loading, setLoading] = useState(false);
  const [fileLoading, setFileLoading] = useState(false);
  const [repoInfo, setRepoInfo] = useState(null);
  const [fileContent, setFileContent] = useState(null);
  const [explanation, setExplanation] = useState(null);
  const [errors, setErrors] = useState({});
  const [lineStart, setLineStart] = useState(0);
  const [lineEnd, setLineEnd] = useState(0);
  const [fileStructure, setFileStructure] = useState([]);
  const [selectedDirectory, setSelectedDirectory] = useState("");

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

      // Fetch the root directory structure
      fetchDirectoryStructure(repoUrl, response.data.default_branch, "");

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

  const fetchDirectoryStructure = async (repoUrl, branch, path) => {
    try {
      const response = await repositoryService.listFiles(repoUrl, path, branch);
      setFileStructure(response.data.files || []);
      setSelectedDirectory(path);
    } catch (error) {
      console.error("Error fetching directory structure:", error);
      toast.error("Could not fetch directory structure.");
    }
  };

  const navigateToDirectory = (path) => {
    const branchToUse = branch || (repoInfo ? repoInfo.default_branch : "main");
    fetchDirectoryStructure(url, branchToUse, path);
  };

  const navigateUp = () => {
    if (!selectedDirectory) return;

    const parts = selectedDirectory.split("/");
    parts.pop();
    const parentPath = parts.join("/");

    const branchToUse = branch || (repoInfo ? repoInfo.default_branch : "main");
    fetchDirectoryStructure(url, branchToUse, parentPath);
  };

  const selectFile = (file) => {
    setFilePath(file.path);
    loadFileContent(file.path);
  };

  const loadFileContent = async (path) => {
    if (!url || !path) return;

    setFileLoading(true);
    try {
      const branchToUse =
        branch || (repoInfo ? repoInfo.default_branch : "main");
      const response = await repositoryService.getFile(url, path, branchToUse);
      setFileContent(response.data.content);
      setFileLoading(false);
    } catch (error) {
      console.error("Error loading file content:", error);
      toast.error("Could not load file content.");
      setFileLoading(false);
    }
  };

  const handleExplainFunction = async () => {
    if (!url || !filePath || !functionName) {
      setErrors({
        url: !url ? "Please enter a GitHub repository URL" : "",
        filePath: !filePath ? "Please select a file" : "",
        functionName: !functionName ? "Please enter a function name" : "",
      });
      return;
    }

    setLoading(true);
    setErrors({});
    setExplanation(null);

    try {
      const branchToUse =
        branch || (repoInfo ? repoInfo.default_branch : "main");

      toast.info("Analyzing function...");

      // Call API to explain function
      const response = await navigatorService.explainFunction(
        url,
        branchToUse,
        filePath,
        functionName,
        lineStart > 0 ? lineStart : undefined,
        lineEnd > 0 ? lineEnd : undefined
      );

      setExplanation(response.data);
      toast.success("Function explanation generated!");

      // Fetch repo info if not already loaded
      if (!repoInfo) {
        await fetchRepositoryInfo(url);
      }
    } catch (error) {
      console.error("Error explaining function:", error);

      let errorMessage = "Failed to explain function";
      if (error.response?.data?.error) {
        errorMessage = error.response.data.error;
      }

      toast.error(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  const detectLanguage = (filePath) => {
    if (!filePath) return "text";

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
              Function Explainer
            </h1>
            <p className="text-gray-600 dark:text-gray-300">
              Get detailed explanations of any function or component in the
              codebase
            </p>
          </div>
          <div className="mt-4 md:mt-0">
            <div className="text-sm inline-flex items-center rounded-full bg-secondary-50 dark:bg-secondary-900/20 px-3 py-1 text-secondary-700 dark:text-secondary-300 border border-secondary-200 dark:border-secondary-800">
              <svg
                xmlns="http://www.w3.org/2000/svg"
                className="h-4 w-4 mr-1"
                viewBox="0 0 20 20"
                fill="currentColor"
              >
                <path
                  fillRule="evenodd"
                  d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-8-3a1 1 0 00-.867.5 1 1 0 11-1.731-1A3 3 0 0113 8a3.001 3.001 0 01-2 2.83V11a1 1 0 11-2 0v-1a1 1 0 011-1 1 1 0 100-2zm0 8a1 1 0 100-2 1 1 0 000 2z"
                  clipRule="evenodd"
                />
              </svg>
              <span>Understand functions in plain English</span>
            </div>
          </div>
        </div>
      </motion.div>

      <div className="grid grid-cols-1 lg:grid-cols-5 gap-6 mb-6">
        {/* Repository and function selection */}
        <div className="lg:col-span-2">
          <Card className="bg-gradient-to-b from-dark-100/70 to-dark-100 border-none shadow-glow h-full">
            <div className="space-y-4">
              <div>
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

              <div>
                <label className="block text-sm font-medium text-gray-200 mb-1">
                  File Browser
                </label>
                {url ? (
                  <div className="border border-gray-800 rounded-md bg-dark-200 overflow-hidden">
                    {/* Directory navigation */}
                    <div className="bg-dark-300 p-2 flex items-center border-b border-gray-800">
                      <button
                        onClick={navigateUp}
                        disabled={!selectedDirectory}
                        className={`p-1 rounded mr-1 ${
                          !selectedDirectory
                            ? "text-gray-600"
                            : "text-gray-400 hover:text-white hover:bg-dark-200"
                        }`}
                      >
                        <svg
                          xmlns="http://www.w3.org/2000/svg"
                          className="h-5 w-5"
                          viewBox="0 0 20 20"
                          fill="currentColor"
                        >
                          <path
                            fillRule="evenodd"
                            d="M12.707 5.293a1 1 0 010 1.414L9.414 10l3.293 3.293a1 1 0 01-1.414 1.414l-4-4a1 1 0 010-1.414l4-4a1 1 0 011.414 0z"
                            clipRule="evenodd"
                          />
                        </svg>
                      </button>
                      <div className="text-sm text-gray-300 truncate flex-1">
                        /{selectedDirectory}
                      </div>
                    </div>

                    {/* File list */}
                    <div className="max-h-60 overflow-y-auto scrollbar-thin scrollbar-thumb-gray-700 scrollbar-track-transparent">
                      {fileStructure.length > 0 ? (
                        <ul className="divide-y divide-gray-800">
                          {fileStructure.map((file, index) => (
                            <li
                              key={index}
                              className={`p-2 hover:bg-dark-300 flex items-center ${
                                filePath === file.path ? "bg-dark-300" : ""
                              }`}
                            >
                              {file.type === "dir" ? (
                                <button
                                  className="w-full text-left flex items-center"
                                  onClick={() => navigateToDirectory(file.path)}
                                >
                                  <svg
                                    xmlns="http://www.w3.org/2000/svg"
                                    className="h-5 w-5 mr-2 text-primary-400"
                                    viewBox="0 0 20 20"
                                    fill="currentColor"
                                  >
                                    <path d="M2 6a2 2 0 012-2h5l2 2h5a2 2 0 012 2v6a2 2 0 01-2 2H4a2 2 0 01-2-2V6z" />
                                  </svg>
                                  <span className="text-sm text-gray-300">
                                    {file.name}
                                  </span>
                                </button>
                              ) : (
                                <button
                                  className="w-full text-left flex items-center"
                                  onClick={() => selectFile(file)}
                                >
                                  <svg
                                    xmlns="http://www.w3.org/2000/svg"
                                    className="h-5 w-5 mr-2 text-gray-400"
                                    viewBox="0 0 20 20"
                                    fill="currentColor"
                                  >
                                    <path
                                      fillRule="evenodd"
                                      d="M4 4a2 2 0 012-2h4.586A2 2 0 0112 2.586L15.414 6A2 2 0 0116 7.414V16a2 2 0 01-2 2H6a2 2 0 01-2-2V4z"
                                      clipRule="evenodd"
                                    />
                                  </svg>
                                  <span className="text-sm text-gray-300">
                                    {file.name}
                                  </span>
                                </button>
                              )}
                            </li>
                          ))}
                        </ul>
                      ) : (
                        <div className="p-4 text-center text-gray-500">
                          {url
                            ? "Loading files..."
                            : "Enter a repository URL to browse files"}
                        </div>
                      )}
                    </div>
                  </div>
                ) : (
                  <div className="p-4 text-center text-gray-500 border border-gray-800 rounded-md">
                    Enter a repository URL to browse files
                  </div>
                )}
              </div>

              <div>
                <Input
                  label="Selected File"
                  value={filePath}
                  onChange={(e) => setFilePath(e.target.value)}
                  error={errors.filePath}
                  readOnly
                  dark
                />
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <Input
                    label="Function Name"
                    placeholder="main"
                    value={functionName}
                    onChange={(e) => setFunctionName(e.target.value)}
                    error={errors.functionName}
                    required
                    dark
                  />
                  <p className="text-xs text-gray-400 mt-1">
                    Name of the function/method to explain
                  </p>
                </div>

                <div className="grid grid-cols-2 gap-2">
                  <div>
                    <Input
                      label="Line Start (Optional)"
                      type="number"
                      placeholder="0"
                      value={lineStart}
                      onChange={(e) =>
                        setLineStart(parseInt(e.target.value) || 0)
                      }
                      dark
                    />
                  </div>
                  <div>
                    <Input
                      label="Line End (Optional)"
                      type="number"
                      placeholder="0"
                      value={lineEnd}
                      onChange={(e) =>
                        setLineEnd(parseInt(e.target.value) || 0)
                      }
                      dark
                    />
                  </div>
                </div>
              </div>

              <div className="mt-6">
                <Button
                  type="button"
                  onClick={handleExplainFunction}
                  isLoading={loading}
                  disabled={loading || !url || !filePath || !functionName}
                  className="bg-gradient-to-r from-primary-600 to-secondary-600 hover:from-primary-500 hover:to-secondary-500 shadow-glow w-full"
                >
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    className="h-5 w-5 mr-2"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                    />
                  </svg>
                  Explain Function
                </Button>
              </div>
            </div>
          </Card>
        </div>

        {/* Code preview */}
        <div className="lg:col-span-3">
          <Card
            title={filePath ? filePath.split("/").pop() : "Selected File"}
            subtitle={filePath || "No file selected"}
            className="bg-gradient-to-b from-dark-100/70 to-dark-100 border-none shadow-glow h-full"
          >
            {fileLoading ? (
              <div className="flex justify-center items-center py-12">
                <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-primary-500"></div>
              </div>
            ) : fileContent ? (
              <CodeEditor
                code={fileContent}
                language={detectLanguage(filePath)}
                readOnly={true}
                lineNumbers={true}
                showCopyButton={true}
                style={{ minHeight: "300px" }}
              />
            ) : (
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
                    d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                  />
                </svg>
                <p className="text-lg font-medium">
                  Select a file to view the code
                </p>
                <p className="mt-2 text-sm text-gray-600">
                  Browse the repository file structure and select a file
                </p>
              </div>
            )}
          </Card>
        </div>
      </div>

      {/* Function explanation */}
      {explanation && (
        <div className="space-y-6">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5 }}
          >
            <Card
              title={`${explanation.function_name} Explanation`}
              className="bg-gradient-to-b from-dark-100/70 to-dark-100 border-none shadow-glow"
            >
              <div className="prose dark:prose-invert max-w-none text-gray-200">
                <h2 className="text-xl text-accent-400 font-semibold mb-4">
                  Description
                </h2>
                <div
                  dangerouslySetInnerHTML={{
                    __html: DOMPurify.sanitize(
                      marked.parse(explanation.description)
                    ),
                  }}
                />
              </div>
            </Card>
          </motion.div>

          {/* Parameters */}
          {explanation.parameters && explanation.parameters.length > 0 && (
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5, delay: 0.1 }}
            >
              <Card
                title="Parameters"
                className="bg-gradient-to-b from-dark-100/70 to-dark-100 border-none shadow-glow"
              >
                <div className="overflow-x-auto">
                  <table className="min-w-full divide-y divide-gray-800">
                    <thead>
                      <tr>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                          Name
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                          Type
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                          Description
                        </th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-800">
                      {explanation.parameters.map((param, index) => (
                        <tr key={index} className="hover:bg-dark-200">
                          <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-primary-400">
                            {param.name}
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-secondary-300">
                            <code className="bg-dark-300 px-2 py-1 rounded">
                              {param.type || "unknown"}
                            </code>
                          </td>
                          <td className="px-6 py-4 text-sm text-gray-300">
                            {param.description}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </Card>
            </motion.div>
          )}

          {/* Return values */}
          {explanation.return_values &&
            explanation.return_values.length > 0 && (
              <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.5, delay: 0.2 }}
              >
                <Card
                  title="Return Values"
                  className="bg-gradient-to-b from-dark-100/70 to-dark-100 border-none shadow-glow"
                >
                  <div className="overflow-x-auto">
                    <table className="min-w-full divide-y divide-gray-800">
                      <thead>
                        <tr>
                          <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                            Name
                          </th>
                          <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                            Type
                          </th>
                          <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                            Description
                          </th>
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-gray-800">
                        {explanation.return_values.map((param, index) => (
                          <tr key={index} className="hover:bg-dark-200">
                            <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-accent-400">
                              {param.name || "return value"}
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap text-sm text-secondary-300">
                              <code className="bg-dark-300 px-2 py-1 rounded">
                                {param.type || "unknown"}
                              </code>
                            </td>
                            <td className="px-6 py-4 text-sm text-gray-300">
                              {param.description}
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </Card>
              </motion.div>
            )}

          {/* Usage examples */}
          {explanation.usage && explanation.usage.length > 0 && (
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5, delay: 0.3 }}
            >
              <Card
                title="Usage Examples"
                className="bg-gradient-to-b from-dark-100/70 to-dark-100 border-none shadow-glow"
              >
                <div className="space-y-4">
                  {explanation.usage.map((example, index) => (
                    <div key={index} className="bg-dark-200 rounded-md p-4">
                      <CodeEditor
                        code={example}
                        language={detectLanguage(filePath)}
                        readOnly={true}
                        lineNumbers={false}
                        showCopyButton={true}
                      />
                    </div>
                  ))}
                </div>
              </Card>
            </motion.div>
          )}

          {/* Related functions */}
          {explanation.related_functions &&
            explanation.related_functions.length > 0 && (
              <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.5, delay: 0.4 }}
              >
                <Card
                  title="Related Functions"
                  className="bg-gradient-to-b from-dark-100/70 to-dark-100 border-none shadow-glow"
                >
                  <div className="flex flex-wrap gap-2">
                    {explanation.related_functions.map((func, index) => (
                      <div
                        key={index}
                        className="bg-dark-200 rounded-full px-4 py-2 text-sm text-gray-300 cursor-pointer hover:bg-dark-300 border border-gray-800"
                        onClick={() => setFunctionName(func)}
                      >
                        <code>{func}</code>
                      </div>
                    ))}
                  </div>
                </Card>
              </motion.div>
            )}
        </div>
      )}

      {!explanation && !loading && !fileContent && (
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
              Explain Any Function or Component
            </h2>
            <p className="text-gray-400 mb-6 max-w-md mx-auto">
              Select a file, enter a function name, and get a detailed
              explanation of how it works, including parameters, return values,
              and usage examples.
            </p>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mt-6 text-left">
              <div className="p-4 bg-dark-200 rounded-lg border border-gray-800">
                <h3 className="font-medium text-gray-200 mb-2 flex items-center">
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    className="h-5 w-5 mr-2 text-primary-400"
                    viewBox="0 0 20 20"
                    fill="currentColor"
                  >
                    <path d="M13.586 3.586a2 2 0 112.828 2.828l-.793.793-2.828-2.828.793-.793zM11.379 5.793L3 14.172V17h2.828l8.38-8.379-2.83-2.828z" />
                  </svg>
                  Parameters
                </h3>
                <p className="text-sm text-gray-400">
                  Understand what each input parameter does
                </p>
              </div>
              <div className="p-4 bg-dark-200 rounded-lg border border-gray-800">
                <h3 className="font-medium text-gray-200 mb-2 flex items-center">
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    className="h-5 w-5 mr-2 text-secondary-400"
                    viewBox="0 0 20 20"
                    fill="currentColor"
                  >
                    <path
                      fillRule="evenodd"
                      d="M12.316 3.051a1 1 0 01.633 1.265l-4 12a1 1 0 11-1.898-.632l4-12a1 1 0 011.265-.633zM5.707 6.293a1 1 0 010 1.414L3.414 10l2.293 2.293a1 1 0 11-1.414 1.414l-3-3a1 1 0 010-1.414l3-3a1 1 0 011.414 0zm8.586 0a1 1 0 011.414 0l3 3a1 1 0 010 1.414l-3 3a1 1 0 01-1.414-1.414L16.586 10l-2.293-2.293a1 1 0 010-1.414z"
                      clipRule="evenodd"
                    />
                  </svg>
                  Logic Flow
                </h3>
                <p className="text-sm text-gray-400">
                  Understand how the function works internally
                </p>
              </div>
              <div className="p-4 bg-dark-200 rounded-lg border border-gray-800">
                <h3 className="font-medium text-gray-200 mb-2 flex items-center">
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    className="h-5 w-5 mr-2 text-accent-400"
                    viewBox="0 0 20 20"
                    fill="currentColor"
                  >
                    <path d="M4 4a2 2 0 00-2 2v1h16V6a2 2 0 00-2-2H4z" />
                    <path
                      fillRule="evenodd"
                      d="M18 9H2v5a2 2 0 002 2h12a2 2 0 002-2V9zM4 13a1 1 0 011-1h1a1 1 0 110 2H5a1 1 0 01-1-1zm5-1a1 1 0 100 2h1a1 1 0 100-2H9z"
                      clipRule="evenodd"
                    />
                  </svg>
                  Return Values
                </h3>
                <p className="text-sm text-gray-400">
                  See what the function returns and when
                </p>
              </div>
            </div>
          </div>
        </motion.div>
      )}
    </div>
  );
};

export default FunctionExplainer;
