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

const CodeWalkthrough = () => {
  const location = useLocation();
  const [url, setUrl] = useState(location.state?.url || "");
  const [branch, setBranch] = useState("");
  const [loading, setLoading] = useState(false);
  const [repoInfo, setRepoInfo] = useState(null);
  const [walkthrough, setWalkthrough] = useState(null);
  const [selectedStep, setSelectedStep] = useState(null);
  const [selectedFile, setSelectedFile] = useState(null);
  const [errors, setErrors] = useState({});
  const [focusPath, setFocusPath] = useState("");
  const [entryPoints, setEntryPoints] = useState([]);
  const [entryPointInput, setEntryPointInput] = useState("");

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

  const handleAddEntryPoint = () => {
    if (entryPointInput.trim()) {
      setEntryPoints([...entryPoints, entryPointInput.trim()]);
      setEntryPointInput("");
    }
  };

  const handleRemoveEntryPoint = (index) => {
    const newEntryPoints = [...entryPoints];
    newEntryPoints.splice(index, 1);
    setEntryPoints(newEntryPoints);
  };

  const handleGenerateWalkthrough = async () => {
    if (!url) {
      setErrors({ url: "Please enter a GitHub repository URL" });
      return;
    }

    setLoading(true);
    setErrors({});
    setWalkthrough(null);
    setSelectedStep(null);
    setSelectedFile(null);

    try {
      const branchToUse =
        branch || (repoInfo ? repoInfo.default_branch : "main");

      toast.info("Generating walkthrough...", { duration: 2000 });

      // Call API to generate walkthrough
      const response = await navigatorService.generateWalkthrough(
        url,
        branchToUse,
        3, // default depth
        focusPath,
        entryPoints
      );

      setWalkthrough(response.data);

      // Select the first step by default if available
      if (response.data.walkthrough && response.data.walkthrough.length > 0) {
        setSelectedStep(response.data.walkthrough[0]);

        // Fetch the content of the selected step
        if (response.data.walkthrough[0].path) {
          const fileResponse = await repositoryService.getFile(
            url,
            response.data.walkthrough[0].path,
            branchToUse
          );
          setSelectedFile({
            path: response.data.walkthrough[0].path,
            content: fileResponse.data.content,
          });
        }
      }

      toast.success("Walkthrough generated successfully!");

      // Fetch repo info if not already loaded
      if (!repoInfo) {
        await fetchRepositoryInfo(url);
      }
    } catch (error) {
      console.error("Error generating walkthrough:", error);

      let errorMessage = "Failed to generate walkthrough";
      if (error.response?.data?.error) {
        errorMessage = error.response.data.error;
      }

      toast.error(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  const handleSelectStep = async (step) => {
    setSelectedStep(step);

    if (step.path) {
      try {
        const branchToUse =
          branch || (repoInfo ? repoInfo.default_branch : "main");
        const response = await repositoryService.getFile(
          url,
          step.path,
          branchToUse
        );

        setSelectedFile({
          path: step.path,
          content: response.data.content,
        });
      } catch (error) {
        console.error("Error fetching file content:", error);
        toast.error("Could not load file content");
      }
    } else {
      setSelectedFile(null);
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
              Code Walkthrough
            </h1>
            <p className="text-gray-600 dark:text-gray-300">
              Get guided tours of unfamiliar codebases to understand structure
              and flow
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
                <path d="M9 4.804A7.968 7.968 0 005.5 4c-1.255 0-2.443.29-3.5.804v10A7.969 7.969 0 015.5 14c1.669 0 3.218.51 4.5 1.385A7.962 7.962 0 0114.5 14c1.255 0 2.443.29 3.5.804v-10A7.968 7.968 0 0014.5 4c-1.255 0-2.443.29-3.5.804V12a1 1 0 11-2 0V4.804z" />
              </svg>
              <span>Like a guided tour of the codebase</span>
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

          <div className="mt-4">
            <Input
              label="Focus Path (Optional)"
              placeholder="src/main.js"
              helperText="Specify a file or directory to focus on"
              value={focusPath}
              onChange={(e) => setFocusPath(e.target.value)}
              dark
            />
          </div>

          <div className="mt-4">
            <label className="block text-sm font-medium text-gray-200 mb-1">
              Entry Points (Optional)
            </label>
            <p className="text-xs text-gray-400 mb-2">
              Specify key files to include in the walkthrough
            </p>

            <div className="flex gap-2 mb-2">
              <Input
                placeholder="src/index.js"
                value={entryPointInput}
                onChange={(e) => setEntryPointInput(e.target.value)}
                className="flex-grow"
                dark
              />
              <Button
                type="button"
                onClick={handleAddEntryPoint}
                variant="outline"
                className="border-secondary-500 text-secondary-400 hover:bg-secondary-500/10"
              >
                Add
              </Button>
            </div>

            {entryPoints.length > 0 && (
              <div className="mt-2 flex flex-wrap gap-2">
                {entryPoints.map((point, index) => (
                  <div
                    key={index}
                    className="flex items-center bg-dark-200 text-gray-300 px-3 py-1 rounded-full text-sm"
                  >
                    <span className="mr-1">{point}</span>
                    <button
                      type="button"
                      onClick={() => handleRemoveEntryPoint(index)}
                      className="text-gray-400 hover:text-gray-200"
                    >
                      <svg
                        xmlns="http://www.w3.org/2000/svg"
                        className="h-4 w-4"
                        viewBox="0 0 20 20"
                        fill="currentColor"
                      >
                        <path
                          fillRule="evenodd"
                          d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
                          clipRule="evenodd"
                        />
                      </svg>
                    </button>
                  </div>
                ))}
              </div>
            )}
          </div>

          <div className="mt-6">
            <Button
              type="button"
              onClick={handleGenerateWalkthrough}
              isLoading={loading}
              disabled={loading}
              className="bg-gradient-to-r from-primary-600 to-secondary-600 hover:from-primary-500 hover:to-secondary-500 shadow-glow"
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
                  d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                />
              </svg>
              Generate Walkthrough
            </Button>
          </div>
        </Card>
      </div>

      {walkthrough && (
        <div className="space-y-6">
          {/* Overview section */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5 }}
          >
            <Card
              title="Codebase Overview"
              className="bg-gradient-to-b from-dark-100/70 to-dark-100 border-none shadow-glow"
            >
              <div className="prose dark:prose-invert max-w-none text-gray-200">
                <div
                  dangerouslySetInnerHTML={{
                    __html: DOMPurify.sanitize(
                      marked.parse(walkthrough.overview)
                    ),
                  }}
                />
              </div>

              {walkthrough.entry_points &&
                walkthrough.entry_points.length > 0 && (
                  <div className="mt-6">
                    <h3 className="text-lg font-medium text-gray-200 mb-2">
                      Main Entry Points
                    </h3>
                    <div className="bg-dark-200 rounded-lg p-4">
                      <ul className="space-y-2">
                        {walkthrough.entry_points.map((entry, index) => (
                          <li key={index} className="flex items-start">
                            <svg
                              xmlns="http://www.w3.org/2000/svg"
                              className="h-5 w-5 text-accent-400 mr-2 mt-0.5"
                              viewBox="0 0 20 20"
                              fill="currentColor"
                            >
                              <path
                                fillRule="evenodd"
                                d="M12.293 5.293a1 1 0 011.414 0l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414-1.414L14.586 11H3a1 1 0 110-2h11.586l-2.293-2.293a1 1 0 010-1.414z"
                                clipRule="evenodd"
                              />
                            </svg>
                            <code className="text-sm bg-dark-300 text-gray-300 px-1.5 py-0.5 rounded">
                              {entry}
                            </code>
                          </li>
                        ))}
                      </ul>
                    </div>
                  </div>
                )}
            </Card>
          </motion.div>

          {/* Walkthrough steps and code view */}
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            <motion.div
              initial={{ opacity: 0, x: -20 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ duration: 0.5, delay: 0.1 }}
            >
              <Card
                title="Walkthrough Steps"
                noPadding
                className="h-full bg-gradient-to-b from-dark-100/70 to-dark-100 border-none shadow-glow"
              >
                <div className="h-full flex flex-col">
                  <div className="p-4 border-b border-gray-800">
                    <p className="text-sm text-gray-400">
                      Follow these steps to understand the codebase:
                    </p>
                  </div>

                  <div
                    className="flex-1 overflow-y-auto scrollbar-thin scrollbar-thumb-gray-700 scrollbar-track-transparent"
                    style={{ maxHeight: "600px" }}
                  >
                    {walkthrough.walkthrough &&
                    walkthrough.walkthrough.length > 0 ? (
                      <ul className="divide-y divide-gray-800">
                        {walkthrough.walkthrough.map((step, index) => (
                          <li
                            key={index}
                            className={`
                              p-4 hover:bg-dark-200 cursor-pointer transition-colors duration-150
                              ${
                                selectedStep && selectedStep.path === step.path
                                  ? "bg-secondary-900/20 border-l-2 border-secondary-500"
                                  : ""
                              }
                            `}
                            onClick={() => handleSelectStep(step)}
                          >
                            <div className="flex items-center">
                              <div className="flex-shrink-0 h-8 w-8 bg-dark-200 rounded-full text-white flex items-center justify-center mr-3">
                                {index + 1}
                              </div>
                              <div>
                                <h3 className="text-sm font-medium text-gray-200">
                                  {step.name ||
                                    (step.path
                                      ? step.path.split("/").pop()
                                      : `Step ${index + 1}`)}
                                </h3>
                                {step.path && (
                                  <p className="mt-1 text-xs text-gray-500">
                                    {step.path}
                                  </p>
                                )}
                              </div>
                            </div>

                            {step.importance && (
                              <div className="mt-2">
                                <span
                                  className={`inline-block px-2 py-0.5 rounded-full text-xs
                                    ${
                                      step.importance > 7
                                        ? "bg-warning-900/30 text-warning-300"
                                        : step.importance > 4
                                        ? "bg-secondary-900/30 text-secondary-300"
                                        : "bg-primary-900/30 text-primary-300"
                                    }
                                  `}
                                >
                                  Importance: {step.importance}/10
                                </span>
                              </div>
                            )}
                          </li>
                        ))}
                      </ul>
                    ) : (
                      <div className="p-6 text-center text-gray-500">
                        No walkthrough steps available.
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
              {selectedStep ? (
                <div className="space-y-4">
                  <Card className="bg-gradient-to-b from-dark-100/70 to-dark-100 border-none shadow-glow">
                    <h3 className="text-xl font-semibold text-white mb-3">
                      {selectedStep.name ||
                        (selectedStep.path
                          ? selectedStep.path.split("/").pop()
                          : "Step Details")}
                    </h3>

                    <div className="prose dark:prose-invert max-w-none text-gray-200">
                      <div
                        dangerouslySetInnerHTML={{
                          __html: DOMPurify.sanitize(
                            marked.parse(
                              selectedStep.description ||
                                "No description available."
                            )
                          ),
                        }}
                      />
                    </div>
                  </Card>

                  {selectedFile && (
                    <Card
                      title={selectedFile.path.split("/").pop()}
                      subtitle={selectedFile.path}
                      className="bg-gradient-to-b from-dark-100/70 to-dark-100 border-none shadow-glow"
                    >
                      <CodeEditor
                        code={selectedFile.content}
                        language={detectLanguage(selectedFile.path)}
                        readOnly={true}
                        lineNumbers={true}
                        showCopyButton={true}
                        style={{ minHeight: "400px" }}
                      />
                    </Card>
                  )}
                </div>
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
                        d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                      />
                    </svg>
                    <p className="text-lg font-medium">
                      Select a step from the list to view details
                    </p>
                    <p className="mt-2 text-sm text-gray-600">
                      The walkthrough is designed to guide you through the
                      codebase in logical order
                    </p>
                  </div>
                </Card>
              )}
            </motion.div>
          </div>

          {/* Dependencies visualization section */}
          {walkthrough.dependencies &&
            Object.keys(walkthrough.dependencies).length > 0 && (
              <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.5, delay: 0.3 }}
              >
                <Card
                  title="Dependencies Map"
                  className="bg-gradient-to-b from-dark-100/70 to-dark-100 border-none shadow-glow"
                >
                  <div className="overflow-x-auto">
                    <table className="min-w-full divide-y divide-gray-800">
                      <thead>
                        <tr>
                          <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                            Module
                          </th>
                          <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                            Dependencies
                          </th>
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-gray-800">
                        {Object.entries(walkthrough.dependencies).map(
                          ([module, deps], index) => (
                            <tr key={index} className="hover:bg-dark-200">
                              <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-300">
                                <code className="bg-dark-300 px-2 py-1 rounded">
                                  {module}
                                </code>
                              </td>
                              <td className="px-6 py-4 text-sm text-gray-300">
                                <div className="flex flex-wrap gap-2">
                                  {deps.map((dep, i) => (
                                    <code
                                      key={i}
                                      className="bg-dark-300 px-2 py-1 rounded"
                                    >
                                      {dep}
                                    </code>
                                  ))}
                                  {deps.length === 0 && (
                                    <span className="text-gray-500 italic">
                                      No dependencies
                                    </span>
                                  )}
                                </div>
                              </td>
                            </tr>
                          )
                        )}
                      </tbody>
                    </table>
                  </div>
                </Card>
              </motion.div>
            )}
        </div>
      )}

      {!walkthrough && !loading && (
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
                d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
              />
            </svg>
            <h2 className="text-2xl font-bold text-white mb-3">
              Generate a Guided Walkthrough
            </h2>
            <p className="text-gray-400 mb-6 max-w-md mx-auto">
              Enter a GitHub URL to generate a step-by-step guided tour of the
              codebase. Perfect for newcomers trying to understand the project
              structure.
            </p>
            <div className="flex flex-col md:flex-row gap-4 justify-center mt-6">
              <div className="flex-1 text-left p-4 bg-dark-200 rounded-lg border border-gray-800">
                <h3 className="font-medium text-gray-200 mb-2 flex items-center">
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    className="h-5 w-5 mr-2 text-accent-400"
                    viewBox="0 0 20 20"
                    fill="currentColor"
                  >
                    <path
                      fillRule="evenodd"
                      d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
                      clipRule="evenodd"
                    />
                  </svg>
                  Understand Flow
                </h3>
                <p className="text-sm text-gray-400">
                  See how data and control flow through the application with
                  logical steps
                </p>
              </div>
              <div className="flex-1 text-left p-4 bg-dark-200 rounded-lg border border-gray-800">
                <h3 className="font-medium text-gray-200 mb-2 flex items-center">
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    className="h-5 w-5 mr-2 text-primary-400"
                    viewBox="0 0 20 20"
                    fill="currentColor"
                  >
                    <path
                      fillRule="evenodd"
                      d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
                      clipRule="evenodd"
                    />
                  </svg>
                  Identify Key Components
                </h3>
                <p className="text-sm text-gray-400">
                  Discover the most important files and modules to focus on
                  first
                </p>
              </div>
              <div className="flex-1 text-left p-4 bg-dark-200 rounded-lg border border-gray-800">
                <h3 className="font-medium text-gray-200 mb-2 flex items-center">
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    className="h-5 w-5 mr-2 text-secondary-400"
                    viewBox="0 0 20 20"
                    fill="currentColor"
                  >
                    <path
                      fillRule="evenodd"
                      d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
                      clipRule="evenodd"
                    />
                  </svg>
                  See Dependencies
                </h3>
                <p className="text-sm text-gray-400">
                  Visualize how components depend on each other in the
                  architecture
                </p>
              </div>
            </div>
          </div>
        </motion.div>
      )}
    </div>
  );
};

export default CodeWalkthrough;
