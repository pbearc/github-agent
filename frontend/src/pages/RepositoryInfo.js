import React, { useState, useEffect } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import Card from "../components/ui/Card";
import Button from "../components/ui/Button";
import Input from "../components/ui/Input";
import { toast } from "sonner";
import { motion } from "framer-motion";
import { repositoryService } from "../services/api";
import CodeEditor from "../components/ui/CodeEditor";

const RepositoryInfo = () => {
  const location = useLocation();
  const navigate = useNavigate();
  const [url, setUrl] = useState(location.state?.url || "");
  const [branch, setBranch] = useState("");
  const [loading, setLoading] = useState(false);
  const [repoInfo, setRepoInfo] = useState(null);
  const [files, setFiles] = useState([]);
  const [currentPath, setCurrentPath] = useState("/");
  const [selectedFile, setSelectedFile] = useState(null);
  const [fileContent, setFileContent] = useState("");
  const [fileLoading, setFileLoading] = useState(false);
  const [errors, setErrors] = useState({});

  useEffect(() => {
    // If URL is passed from another page, fetch info automatically
    if (location.state?.url) {
      setUrl(location.state.url);
      fetchRepositoryInfo(location.state.url);
    }
  }, [location.state]);

  const fetchRepositoryInfo = async (repoUrl) => {
    if (!repoUrl) {
      setErrors({ url: "Please enter a GitHub repository URL" });
      return;
    }

    setLoading(true);
    setErrors({});

    try {
      const response = await repositoryService.getInfo(repoUrl, branch);
      setRepoInfo(response.data);

      // Add to recently used repositories
      addToRecentRepositories(response.data);

      // List files in the root directory
      await fetchFiles(repoUrl, "/", response.data.default_branch);

      toast.success("Repository information loaded successfully");
    } catch (error) {
      console.error("Error fetching repository info:", error);

      let errorMessage = "Failed to fetch repository information";
      if (error.response?.data?.error) {
        errorMessage = error.response.data.error;
      }

      setErrors({ url: errorMessage });
      toast.error(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  const fetchFiles = async (repoUrl, path, defaultBranch) => {
    setFiles([]);
    try {
      const branchToUse = branch || defaultBranch;
      const response = await repositoryService.listFiles(
        repoUrl,
        path,
        branchToUse
      );
      setFiles(response.data.files);
      setCurrentPath(response.data.path);
    } catch (error) {
      console.error("Error fetching files:", error);
      toast.error("Failed to fetch repository files");
    }
  };

  const fetchFileContent = async (file) => {
    if (file.type !== "file") return;

    setSelectedFile(file);
    setFileLoading(true);

    try {
      const response = await repositoryService.getFile(
        url,
        file.path,
        branch || repoInfo.default_branch
      );
      setFileContent(response.data.content);
    } catch (error) {
      console.error("Error fetching file content:", error);
      toast.error("Failed to fetch file content");
      setFileContent("");
    } finally {
      setFileLoading(false);
    }
  };

  const addToRecentRepositories = (repo) => {
    const recentRepos = JSON.parse(
      localStorage.getItem("recentRepositories") || "[]"
    );

    // Remove if already exists
    const filteredRepos = recentRepos.filter((r) => r.url !== repo.url);

    // Add to the beginning
    const updatedRepos = [
      {
        url: repo.url,
        name: repo.name,
        owner: repo.owner,
      },
      ...filteredRepos,
    ].slice(0, 5); // Keep only the last 5

    localStorage.setItem("recentRepositories", JSON.stringify(updatedRepos));
  };

  const handleNavigateToDirectory = (file) => {
    if (file.type === "dir") {
      fetchFiles(url, file.path, branch || repoInfo.default_branch);
    } else {
      fetchFileContent(file);
    }
  };

  const handleNavigateUp = () => {
    if (currentPath === "/" || currentPath === "") return;

    const pathParts = currentPath.split("/").filter(Boolean);
    pathParts.pop();
    const newPath = pathParts.length === 0 ? "/" : `/${pathParts.join("/")}`;

    fetchFiles(url, newPath, branch || repoInfo.default_branch);
  };

  const detectLanguage = (fileName) => {
    const extension = fileName.split(".").pop().toLowerCase();

    const languageMap = {
      js: "javascript",
      jsx: "javascript",
      ts: "typescript",
      tsx: "typescript",
      py: "python",
      rb: "ruby",
      php: "php",
      java: "java",
      c: "c",
      cpp: "cpp",
      h: "c",
      hpp: "cpp",
      cs: "csharp",
      go: "go",
      rs: "rust",
      swift: "swift",
      kt: "kotlin",
      html: "html",
      css: "css",
      scss: "scss",
      json: "json",
      md: "markdown",
      xml: "xml",
      yml: "yaml",
      yaml: "yaml",
      sh: "bash",
    };

    return languageMap[extension] || "text";
  };

  const handleSubmit = (e) => {
    e.preventDefault();
    fetchRepositoryInfo(url);
  };

  return (
    <div>
      <motion.div
        initial={{ opacity: 0, y: -20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.3 }}
      >
        <h1 className="text-2xl font-bold text-gray-900 dark:text-white mb-6">
          Repository Information
        </h1>
      </motion.div>

      <Card className="mb-6">
        <form onSubmit={handleSubmit}>
          <div className="flex flex-col md:flex-row gap-4">
            <div className="flex-1">
              <Input
                label="GitHub Repository URL"
                placeholder="https://github.com/username/repository"
                value={url}
                onChange={(e) => setUrl(e.target.value)}
                error={errors.url}
                required
              />
            </div>
            <div className="w-full md:w-1/4">
              <Input
                label="Branch (Optional)"
                placeholder="main"
                value={branch}
                onChange={(e) => setBranch(e.target.value)}
              />
            </div>
            <div className="flex items-end">
              <Button type="submit" isLoading={loading} disabled={loading}>
                Fetch Info
              </Button>
            </div>
          </div>
        </form>
      </Card>

      {repoInfo && (
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Repository details */}
          <motion.div
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ duration: 0.3 }}
            className="lg:col-span-1"
          >
            <Card title="Repository Details">
              <div className="space-y-4">
                <div>
                  <h3 className="text-sm font-medium text-gray-500 dark:text-gray-400">
                    Repository
                  </h3>
                  <p className="mt-1 text-lg font-semibold text-gray-900 dark:text-white">
                    {repoInfo.name}
                  </p>
                </div>

                <div>
                  <h3 className="text-sm font-medium text-gray-500 dark:text-gray-400">
                    Owner
                  </h3>
                  <p className="mt-1 text-md text-gray-800 dark:text-gray-200">
                    {repoInfo.owner}
                  </p>
                </div>

                {repoInfo.description && (
                  <div>
                    <h3 className="text-sm font-medium text-gray-500 dark:text-gray-400">
                      Description
                    </h3>
                    <p className="mt-1 text-md text-gray-800 dark:text-gray-200">
                      {repoInfo.description}
                    </p>
                  </div>
                )}

                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <h3 className="text-sm font-medium text-gray-500 dark:text-gray-400">
                      Default Branch
                    </h3>
                    <p className="mt-1 text-md text-gray-800 dark:text-gray-200">
                      {repoInfo.default_branch}
                    </p>
                  </div>

                  <div>
                    <h3 className="text-sm font-medium text-gray-500 dark:text-gray-400">
                      Language
                    </h3>
                    <p className="mt-1 text-md text-gray-800 dark:text-gray-200">
                      {repoInfo.language || "Not specified"}
                    </p>
                  </div>

                  <div>
                    <h3 className="text-sm font-medium text-gray-500 dark:text-gray-400">
                      Stars
                    </h3>
                    <p className="mt-1 text-md text-gray-800 dark:text-gray-200">
                      {repoInfo.stars.toLocaleString()}
                    </p>
                  </div>

                  <div>
                    <h3 className="text-sm font-medium text-gray-500 dark:text-gray-400">
                      Forks
                    </h3>
                    <p className="mt-1 text-md text-gray-800 dark:text-gray-200">
                      {repoInfo.forks.toLocaleString()}
                    </p>
                  </div>
                </div>

                {repoInfo.languages && repoInfo.languages.length > 0 && (
                  <div>
                    <h3 className="text-sm font-medium text-gray-500 dark:text-gray-400">
                      Languages
                    </h3>
                    <div className="mt-2 flex flex-wrap gap-2">
                      {repoInfo.languages.map((lang) => (
                        <span
                          key={lang}
                          className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200"
                        >
                          {lang}
                        </span>
                      ))}
                    </div>
                  </div>
                )}

                <div className="pt-2">
                  <a
                    href={repoInfo.url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="inline-flex items-center text-primary-600 dark:text-primary-500 hover:underline"
                  >
                    <span>View on GitHub</span>
                    <svg
                      xmlns="http://www.w3.org/2000/svg"
                      className="ml-1 h-4 w-4"
                      viewBox="0 0 20 20"
                      fill="currentColor"
                    >
                      <path d="M11 3a1 1 0 100 2h2.586l-6.293 6.293a1 1 0 101.414 1.414L15 6.414V9a1 1 0 102 0V4a1 1 0 00-1-1h-5z" />
                      <path d="M5 5a2 2 0 00-2 2v8a2 2 0 002 2h8a2 2 0 002-2v-3a1 1 0 10-2 0v3H5V7h3a1 1 0 000-2H5z" />
                    </svg>
                  </a>
                </div>

                <div className="space-y-2 pt-4">
                  <h3 className="text-sm font-medium text-gray-500 dark:text-gray-400">
                    Quick Actions
                  </h3>
                  <div className="grid grid-cols-2 gap-2">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() =>
                        navigate("/generate/readme", {
                          state: { url: repoInfo.url },
                        })
                      }
                    >
                      Generate README
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() =>
                        navigate("/generate/dockerfile", {
                          state: { url: repoInfo.url },
                        })
                      }
                    >
                      Generate Dockerfile
                    </Button>
                  </div>
                </div>
              </div>
            </Card>
          </motion.div>

          {/* Files and Content */}
          <motion.div
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ duration: 0.3, delay: 0.1 }}
            className="lg:col-span-2"
          >
            <Card title="Repository Files" noPadding>
              <div className="flex flex-col h-full">
                {/* File browser */}
                <div className="p-4 border-b border-gray-200 dark:border-gray-700">
                  <div className="flex items-center mb-4">
                    <button
                      onClick={handleNavigateUp}
                      disabled={currentPath === "/"}
                      className={`
                        inline-flex items-center mr-4 text-sm text-gray-500 dark:text-gray-400
                        ${
                          currentPath === "/"
                            ? "opacity-50 cursor-not-allowed"
                            : "hover:text-gray-700 dark:hover:text-gray-300"
                        }
                      `}
                    >
                      <svg
                        xmlns="http://www.w3.org/2000/svg"
                        className="h-5 w-5 mr-1"
                        viewBox="0 0 20 20"
                        fill="currentColor"
                      >
                        <path
                          fillRule="evenodd"
                          d="M12.707 5.293a1 1 0 010 1.414L9.414 10l3.293 3.293a1 1 0 01-1.414 1.414l-4-4a1 1 0 010-1.414l4-4a1 1 0 011.414 0z"
                          clipRule="evenodd"
                        />
                      </svg>
                      Back
                    </button>
                    <div className="flex-1 overflow-hidden">
                      <div className="text-sm font-medium text-gray-900 dark:text-white truncate">
                        {currentPath === "/" ? "/" : currentPath}
                      </div>
                    </div>
                  </div>

                  <div className="overflow-x-auto">
                    <table className="min-w-full">
                      <thead className="bg-gray-50 dark:bg-gray-800">
                        <tr>
                          <th
                            scope="col"
                            className="px-4 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider"
                          >
                            Name
                          </th>
                          <th
                            scope="col"
                            className="px-4 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider"
                          >
                            Type
                          </th>
                          <th
                            scope="col"
                            className="px-4 py-2 text-right text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider"
                          >
                            Size
                          </th>
                        </tr>
                      </thead>
                      <tbody className="bg-white dark:bg-dark-100 divide-y divide-gray-200 dark:divide-gray-700">
                        {files.length > 0 ? (
                          files.map((file) => (
                            <tr
                              key={file.path}
                              className="hover:bg-gray-50 dark:hover:bg-gray-800 cursor-pointer"
                              onClick={() => handleNavigateToDirectory(file)}
                            >
                              <td className="px-4 py-2 whitespace-nowrap">
                                <div className="flex items-center">
                                  {file.type === "dir" ? (
                                    <svg
                                      xmlns="http://www.w3.org/2000/svg"
                                      className="h-5 w-5 text-yellow-500 mr-2"
                                      viewBox="0 0 20 20"
                                      fill="currentColor"
                                    >
                                      <path d="M2 6a2 2 0 012-2h5l2 2h5a2 2 0 012 2v6a2 2 0 01-2 2H4a2 2 0 01-2-2V6z" />
                                    </svg>
                                  ) : (
                                    <svg
                                      xmlns="http://www.w3.org/2000/svg"
                                      className="h-5 w-5 text-gray-400 mr-2"
                                      viewBox="0 0 20 20"
                                      fill="currentColor"
                                    >
                                      <path
                                        fillRule="evenodd"
                                        d="M4 4a2 2 0 012-2h4.586A2 2 0 0112 2.586L15.414 6A2 2 0 0116 7.414V16a2 2 0 01-2 2H6a2 2 0 01-2-2V4z"
                                        clipRule="evenodd"
                                      />
                                    </svg>
                                  )}
                                  <span className="text-sm font-medium text-gray-900 dark:text-white">
                                    {file.name}
                                  </span>
                                </div>
                              </td>
                              <td className="px-4 py-2 whitespace-nowrap">
                                <span className="text-sm text-gray-500 dark:text-gray-400">
                                  {file.type === "dir" ? "Directory" : "File"}
                                </span>
                              </td>
                              <td className="px-4 py-2 whitespace-nowrap text-right text-sm text-gray-500 dark:text-gray-400">
                                {file.type === "dir"
                                  ? "-"
                                  : `${(file.size / 1024).toFixed(1)} KB`}
                              </td>
                            </tr>
                          ))
                        ) : (
                          <tr>
                            <td
                              colSpan="3"
                              className="px-4 py-4 text-center text-sm text-gray-500 dark:text-gray-400"
                            >
                              {loading
                                ? "Loading files..."
                                : "No files found in this directory"}
                            </td>
                          </tr>
                        )}
                      </tbody>
                    </table>
                  </div>
                </div>

                {/* File content viewer */}
                {selectedFile && (
                  <div className="p-4">
                    <div className="flex items-center justify-between mb-4">
                      <h3 className="text-md font-medium text-gray-900 dark:text-white">
                        {selectedFile.name}
                      </h3>
                      <div className="flex space-x-2">
                        <a
                          href={selectedFile.html_url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="inline-flex items-center text-sm text-primary-600 dark:text-primary-500 hover:underline"
                        >
                          View on GitHub
                        </a>
                      </div>
                    </div>

                    {fileLoading ? (
                      <div className="flex justify-center py-8">
                        <svg
                          className="animate-spin h-8 w-8 text-primary-600"
                          xmlns="http://www.w3.org/2000/svg"
                          fill="none"
                          viewBox="0 0 24 24"
                        >
                          <circle
                            className="opacity-25"
                            cx="12"
                            cy="12"
                            r="10"
                            stroke="currentColor"
                            strokeWidth="4"
                          ></circle>
                          <path
                            className="opacity-75"
                            fill="currentColor"
                            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                          ></path>
                        </svg>
                      </div>
                    ) : (
                      <CodeEditor
                        code={fileContent}
                        language={detectLanguage(selectedFile.name)}
                        readOnly={true}
                        lineNumbers={true}
                        showCopyButton={true}
                        maxHeight="500px"
                      />
                    )}
                  </div>
                )}
              </div>
            </Card>
          </motion.div>
        </div>
      )}
    </div>
  );
};

export default RepositoryInfo;
