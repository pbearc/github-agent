import React, { useState, useEffect } from "react";
import { useLocation } from "react-router-dom";
import Card from "../components/ui/Card";
import Button from "../components/ui/Button";
import Input from "../components/ui/Input";
import { toast } from "sonner";
import { motion } from "framer-motion";
import { repositoryService, navigatorService } from "../services/api";
import DOMPurify from "dompurify";
import { marked } from "marked";

const ArchitectureAndPractices = () => {
  const location = useLocation();
  const [url, setUrl] = useState(location.state?.url || "");
  const [branch, setBranch] = useState("");
  const [loading, setLoading] = useState(false);
  const [practicesLoading, setPracticesLoading] = useState(false);
  const [repoInfo, setRepoInfo] = useState(null);
  const [architecture, setArchitecture] = useState(null);
  const [bestPractices, setBestPractices] = useState(null);
  const [errors, setErrors] = useState({});
  const [detailLevel, setDetailLevel] = useState("medium");
  const [focusPaths, setFocusPaths] = useState([]);
  const [focusPathInput, setFocusPathInput] = useState("");
  const [activeTab, setActiveTab] = useState("architecture");
  const [scope, setScope] = useState("full");
  const [scopePath, setScopePath] = useState("");

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

  const handleAddFocusPath = () => {
    if (focusPathInput.trim()) {
      setFocusPaths([...focusPaths, focusPathInput.trim()]);
      setFocusPathInput("");
    }
  };

  const handleRemoveFocusPath = (index) => {
    const newFocusPaths = [...focusPaths];
    newFocusPaths.splice(index, 1);
    setFocusPaths(newFocusPaths);
  };

  const handleGenerateArchitecture = async () => {
    if (!url) {
      setErrors({ url: "Please enter a GitHub repository URL" });
      return;
    }

    setLoading(true);
    setErrors({});
    setArchitecture(null);

    try {
      const branchToUse =
        branch || (repoInfo ? repoInfo.default_branch : "main");

      toast.info("Generating architecture visualization...", {
        duration: 3000,
      });

      // Call API to generate architecture visualization
      const response = await navigatorService.visualizeArchitecture(
        url,
        branchToUse,
        detailLevel,
        focusPaths
      );

      setArchitecture(response.data);
      toast.success("Architecture visualization generated!");

      // Fetch repo info if not already loaded
      if (!repoInfo) {
        await fetchRepositoryInfo(url);
      }
    } catch (error) {
      console.error("Error generating architecture visualization:", error);

      let errorMessage = "Failed to generate architecture visualization";
      if (error.response?.data?.error) {
        errorMessage = error.response.data.error;
      }

      toast.error(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  const handleGeneratePractices = async () => {
    if (!url) {
      setErrors({ url: "Please enter a GitHub repository URL" });
      return;
    }

    setPracticesLoading(true);
    setErrors({});
    setBestPractices(null);

    try {
      const branchToUse =
        branch || (repoInfo ? repoInfo.default_branch : "main");

      toast.info("Analyzing code practices...", { duration: 3000 });

      // Call API to generate best practices guide
      const response = await navigatorService.getBestPractices(
        url,
        branchToUse,
        scope,
        scopePath
      );

      setBestPractices(response.data);
      toast.success("Best practices guide generated!");

      // Fetch repo info if not already loaded
      if (!repoInfo) {
        await fetchRepositoryInfo(url);
      }
    } catch (error) {
      console.error("Error generating best practices guide:", error);

      let errorMessage = "Failed to generate best practices guide";
      if (error.response?.data?.error) {
        errorMessage = error.response.data.error;
      }

      toast.error(errorMessage);
    } finally {
      setPracticesLoading(false);
    }
  };

  // Generate a simple Mermaid diagram from the architecture data
  const generateMermaidDiagram = () => {
    if (!architecture || !architecture.diagram_data) return "";

    let mermaidCode = "graph TD\n";

    // Add nodes
    architecture.diagram_data.nodes.forEach((node) => {
      mermaidCode += `  ${node.id}["${node.label}"]\n`;
    });

    // Add edges
    architecture.diagram_data.edges.forEach((edge) => {
      let edgeStyle = "";
      switch (edge.type) {
        case "imports":
          edgeStyle = "-->|imports|";
          break;
        case "extends":
          edgeStyle = "==>|extends|";
          break;
        case "calls":
          edgeStyle = "-.->|calls|";
          break;
        default:
          edgeStyle = "-->";
      }

      mermaidCode += `  ${edge.source}${edgeStyle}${edge.target}\n`;
    });

    return mermaidCode;
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
              Codebase Structure & Patterns
            </h1>
            <p className="text-gray-600 dark:text-gray-300">
              Visualize architecture and understand coding patterns used in the
              repository
            </p>
          </div>
          <div className="mt-4 md:mt-0">
            <div className="text-sm inline-flex items-center rounded-full bg-accent-50 dark:bg-accent-900/20 px-3 py-1 text-accent-700 dark:text-accent-300 border border-accent-200 dark:border-accent-800">
              <svg
                xmlns="http://www.w3.org/2000/svg"
                className="h-4 w-4 mr-1"
                viewBox="0 0 20 20"
                fill="currentColor"
              >
                <path d="M5 3a2 2 0 00-2 2v2a2 2 0 002 2h2a2 2 0 002-2V5a2 2 0 00-2-2H5zM5 11a2 2 0 00-2 2v2a2 2 0 002 2h2a2 2 0 002-2v-2a2 2 0 00-2-2H5zM11 5a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V5zM11 13a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" />
              </svg>
              <span>Understand architecture and code patterns</span>
            </div>
          </div>
        </div>
      </motion.div>

      {/* Tab Navigation */}
      <div className="flex mb-6">
        <button
          onClick={() => setActiveTab("architecture")}
          className={`px-4 py-2 rounded-t-lg flex items-center ${
            activeTab === "architecture"
              ? "bg-dark-100 border-t border-l border-r border-gray-800 text-white font-medium"
              : "bg-gray-900 text-gray-400 hover:text-gray-200"
          }`}
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            className="h-5 w-5 mr-2"
            viewBox="0 0 20 20"
            fill="currentColor"
          >
            <path d="M5 3a2 2 0 00-2 2v2a2 2 0 002 2h2a2 2 0 002-2V5a2 2 0 00-2-2H5zM5 11a2 2 0 00-2 2v2a2 2 0 002 2h2a2 2 0 002-2v-2a2 2 0 00-2-2H5zM11 5a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V5zM11 13a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" />
          </svg>
          Architecture Visualizer
        </button>
        <button
          onClick={() => setActiveTab("practices")}
          className={`px-4 py-2 rounded-t-lg flex items-center ${
            activeTab === "practices"
              ? "bg-dark-100 border-t border-l border-r border-gray-800 text-white font-medium"
              : "bg-gray-900 text-gray-400 hover:text-gray-200"
          }`}
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            className="h-5 w-5 mr-2"
            viewBox="0 0 20 20"
            fill="currentColor"
          >
            <path
              fillRule="evenodd"
              d="M12.316 3.051a1 1 0 01.633 1.265l-4 12a1 1 0 11-1.898-.632l4-12a1 1 0 011.265-.633zM5.707 6.293a1 1 0 010 1.414L3.414 10l2.293 2.293a1 1 0 11-1.414 1.414l-3-3a1 1 0 010-1.414l3-3a1 1 0 011.414 0zm8.586 0a1 1 0 011.414 0l3 3a1 1 0 010 1.414l-3 3a1 1 0 01-1.414-1.414L16.586 10l-2.293-2.293a1 1 0 010-1.414z"
              clipRule="evenodd"
            />
          </svg>
          Best Practices Guide
        </button>
      </div>

      {activeTab === "architecture" ? (
        <div>
          <Card className="bg-gradient-to-b from-dark-100/70 to-dark-100 border-none shadow-glow mb-6">
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
              <label className="block text-sm font-medium text-gray-200 mb-1">
                Detail Level
              </label>
              <div className="flex space-x-4">
                <label className="flex items-center">
                  <input
                    type="radio"
                    name="detailLevel"
                    value="low"
                    checked={detailLevel === "low"}
                    onChange={() => setDetailLevel("low")}
                    className="h-4 w-4 text-primary-600 focus:ring-primary-500 border-gray-700 bg-dark-200"
                  />
                  <span className="ml-2 text-gray-300">Low</span>
                </label>
                <label className="flex items-center">
                  <input
                    type="radio"
                    name="detailLevel"
                    value="medium"
                    checked={detailLevel === "medium"}
                    onChange={() => setDetailLevel("medium")}
                    className="h-4 w-4 text-primary-600 focus:ring-primary-500 border-gray-700 bg-dark-200"
                  />
                  <span className="ml-2 text-gray-300">Medium</span>
                </label>
                <label className="flex items-center">
                  <input
                    type="radio"
                    name="detailLevel"
                    value="high"
                    checked={detailLevel === "high"}
                    onChange={() => setDetailLevel("high")}
                    className="h-4 w-4 text-primary-600 focus:ring-primary-500 border-gray-700 bg-dark-200"
                  />
                  <span className="ml-2 text-gray-300">High</span>
                </label>
              </div>
              <p className="text-xs text-gray-400 mt-1">
                Higher detail levels provide more components but may be more
                complex
              </p>
            </div>

            <div className="mt-4">
              <label className="block text-sm font-medium text-gray-200 mb-1">
                Focus Paths (Optional)
              </label>
              <p className="text-xs text-gray-400 mb-2">
                Specify directories or files to focus on in the architecture
              </p>

              <div className="flex gap-2 mb-2">
                <Input
                  placeholder="src/components"
                  value={focusPathInput}
                  onChange={(e) => setFocusPathInput(e.target.value)}
                  className="flex-grow"
                  dark
                />
                <Button
                  type="button"
                  onClick={handleAddFocusPath}
                  variant="outline"
                  className="border-secondary-500 text-secondary-400 hover:bg-secondary-500/10"
                >
                  Add
                </Button>
              </div>

              {focusPaths.length > 0 && (
                <div className="mt-2 flex flex-wrap gap-2">
                  {focusPaths.map((path, index) => (
                    <div
                      key={index}
                      className="flex items-center bg-dark-200 text-gray-300 px-3 py-1 rounded-full text-sm"
                    >
                      <span className="mr-1">{path}</span>
                      <button
                        type="button"
                        onClick={() => handleRemoveFocusPath(index)}
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
                onClick={handleGenerateArchitecture}
                isLoading={loading}
                disabled={loading || !url}
                className="bg-gradient-to-r from-primary-600 to-secondary-600 hover:from-primary-500 hover:to-secondary-500 shadow-glow"
              >
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  className="h-5 w-5 mr-2"
                  viewBox="0 0 20 20"
                  fill="currentColor"
                >
                  <path d="M5 3a2 2 0 00-2 2v2a2 2 0 002 2h2a2 2 0 002-2V5a2 2 0 00-2-2H5zM5 11a2 2 0 00-2 2v2a2 2 0 002 2h2a2 2 0 002-2v-2a2 2 0 00-2-2H5zM11 5a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V5zM11 13a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" />
                </svg>
                Visualize Architecture
              </Button>
            </div>
          </Card>

          {/* Architecture Visualization Results */}
          {architecture && (
            <div className="space-y-6">
              <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.5 }}
              >
                <Card
                  title="Architecture Overview"
                  className="bg-gradient-to-b from-dark-100/70 to-dark-100 border-none shadow-glow"
                >
                  <div className="prose dark:prose-invert max-w-none text-gray-200">
                    <div
                      dangerouslySetInnerHTML={{
                        __html: DOMPurify.sanitize(
                          marked.parse(architecture.overview)
                        ),
                      }}
                    />
                  </div>
                </Card>
              </motion.div>

              {/* Architecture Diagram */}
              <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.5, delay: 0.1 }}
              >
                <Card
                  title="Architecture Diagram"
                  className="bg-gradient-to-b from-dark-100/70 to-dark-100 border-none shadow-glow"
                >
                  <div
                    className="bg-dark-200 p-4 rounded-lg overflow-auto"
                    style={{ maxHeight: "600px" }}
                  >
                    {/* TODO: Add a proper diagram visualization here */}
                    <pre className="text-xs text-gray-300 overflow-auto p-4 bg-dark-300 rounded-md">
                      {generateMermaidDiagram()}
                    </pre>
                  </div>
                  <p className="text-sm text-gray-400 mt-4">
                    Note: The diagram is a simplified representation of the
                    codebase architecture. Component sizes represent their
                    relative importance in the system.
                  </p>
                </Card>
              </motion.div>

              {/* Component Descriptions */}
              {architecture.component_descriptions &&
                Object.keys(architecture.component_descriptions).length > 0 && (
                  <motion.div
                    initial={{ opacity: 0, y: 20 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ duration: 0.5, delay: 0.2 }}
                  >
                    <Card
                      title="Component Descriptions"
                      className="bg-gradient-to-b from-dark-100/70 to-dark-100 border-none shadow-glow"
                    >
                      <div className="space-y-6">
                        {Object.entries(
                          architecture.component_descriptions
                        ).map(([component, description], index) => (
                          <div
                            key={index}
                            className="p-4 bg-dark-200 rounded-lg"
                          >
                            <h3 className="text-lg font-medium text-primary-400 mb-2">
                              {component}
                            </h3>
                            <div className="prose dark:prose-invert max-w-none text-gray-300 text-sm">
                              <div
                                dangerouslySetInnerHTML={{
                                  __html: DOMPurify.sanitize(
                                    marked.parse(description)
                                  ),
                                }}
                              />
                            </div>
                          </div>
                        ))}
                      </div>
                    </Card>
                  </motion.div>
                )}
            </div>
          )}

          {!architecture && !loading && (
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
                    d="M5 3a2 2 0 00-2 2v2a2 2 0 002 2h2a2 2 0 002-2V5a2 2 0 00-2-2H5zM5 11a2 2 0 00-2 2v2a2 2 0 002 2h2a2 2 0 002-2v-2a2 2 0 00-2-2H5zM11 5a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V5zM11 13a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z"
                  />
                </svg>
                <h2 className="text-2xl font-bold text-white mb-3">
                  Visualize Codebase Architecture
                </h2>
                <p className="text-gray-400 mb-6 max-w-md mx-auto">
                  Generate a visual representation of the codebase architecture
                  to understand how components relate to each other.
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
                        <path d="M5 3a2 2 0 00-2 2v2a2 2 0 002 2h2a2 2 0 002-2V5a2 2 0 00-2-2H5z" />
                      </svg>
                      Component Relationships
                    </h3>
                    <p className="text-sm text-gray-400">
                      See how different parts of the codebase connect
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
                          d="M3 5a2 2 0 012-2h10a2 2 0 012 2v10a2 2 0 01-2 2H5a2 2 0 01-2-2V5zm11 1H6v8l4-2 4 2V6z"
                          clipRule="evenodd"
                        />
                      </svg>
                      Architectural Patterns
                    </h3>
                    <p className="text-sm text-gray-400">
                      Identify design patterns used in the project
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
                        <path d="M9 4.804A7.968 7.968 0 005.5 4c-1.255 0-2.443.29-3.5.804v10A7.969 7.969 0 015.5 14c1.669 0 3.218.51 4.5 1.385A7.962 7.962 0 0114.5 14c1.255 0 2.443.29 3.5.804v-10A7.968 7.968 0 0014.5 4c-1.255 0-2.443.29-3.5.804V12a1 1 0 11-2 0V4.804z" />
                      </svg>
                      Dependency Flow
                    </h3>
                    <p className="text-sm text-gray-400">
                      Understand how data flows through the system
                    </p>
                  </div>
                </div>
              </div>
            </motion.div>
          )}
        </div>
      ) : (
        <div>
          <Card className="bg-gradient-to-b from-dark-100/70 to-dark-100 border-none shadow-glow mb-6">
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
              <label className="block text-sm font-medium text-gray-200 mb-1">
                Analysis Scope
              </label>
              <div className="flex space-x-4">
                <label className="flex items-center">
                  <input
                    type="radio"
                    name="scope"
                    value="full"
                    checked={scope === "full"}
                    onChange={() => {
                      setScope("full");
                      setScopePath("");
                    }}
                    className="h-4 w-4 text-primary-600 focus:ring-primary-500 border-gray-700 bg-dark-200"
                  />
                  <span className="ml-2 text-gray-300">Full Repository</span>
                </label>
                <label className="flex items-center">
                  <input
                    type="radio"
                    name="scope"
                    value="directory"
                    checked={scope === "directory"}
                    onChange={() => setScope("directory")}
                    className="h-4 w-4 text-primary-600 focus:ring-primary-500 border-gray-700 bg-dark-200"
                  />
                  <span className="ml-2 text-gray-300">Directory</span>
                </label>
                <label className="flex items-center">
                  <input
                    type="radio"
                    name="scope"
                    value="file"
                    checked={scope === "file"}
                    onChange={() => setScope("file")}
                    className="h-4 w-4 text-primary-600 focus:ring-primary-500 border-gray-700 bg-dark-200"
                  />
                  <span className="ml-2 text-gray-300">Single File</span>
                </label>
              </div>
            </div>

            {scope !== "full" && (
              <div className="mt-4">
                <Input
                  label={scope === "directory" ? "Directory Path" : "File Path"}
                  placeholder={
                    scope === "directory" ? "src/components" : "src/index.js"
                  }
                  value={scopePath}
                  onChange={(e) => setScopePath(e.target.value)}
                  error={
                    scope !== "full" && !scopePath
                      ? `Please enter a ${scope} path`
                      : ""
                  }
                  required
                  dark
                />
              </div>
            )}

            <div className="mt-6">
              <Button
                type="button"
                onClick={handleGeneratePractices}
                isLoading={practicesLoading}
                disabled={
                  practicesLoading || !url || (scope !== "full" && !scopePath)
                }
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
                    d="M12.316 3.051a1 1 0 01.633 1.265l-4 12a1 1 0 11-1.898-.632l4-12a1 1 0 011.265-.633zM5.707 6.293a1 1 0 010 1.414L3.414 10l2.293 2.293a1 1 0 11-1.414 1.414l-3-3a1 1 0 010-1.414l3-3a1 1 0 011.414 0zm8.586 0a1 1 0 011.414 0l3 3a1 1 0 010 1.414l-3 3a1 1 0 01-1.414-1.414L16.586 10l-2.293-2.293a1 1 0 010-1.414z"
                    clipRule="evenodd"
                  />
                </svg>
                Generate Best Practices Guide
              </Button>
            </div>
          </Card>

          {/* Best Practices Guide Results */}
          {bestPractices && (
            <div className="space-y-6">
              {/* Style Guide */}
              <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.5 }}
              >
                <Card
                  title="Style Guide"
                  className="bg-gradient-to-b from-dark-100/70 to-dark-100 border-none shadow-glow"
                >
                  <div className="prose dark:prose-invert max-w-none text-gray-200">
                    <div
                      dangerouslySetInnerHTML={{
                        __html: DOMPurify.sanitize(
                          marked.parse(
                            bestPractices.style_guide ||
                              "No style guide information available."
                          )
                        ),
                      }}
                    />
                  </div>
                </Card>
              </motion.div>

              {/* Best Practices */}
              {bestPractices.practices && bestPractices.practices.length > 0 ? (
                <motion.div
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.5, delay: 0.1 }}
                >
                  <Card
                    title="Best Practices"
                    className="bg-gradient-to-b from-dark-100/70 to-dark-100 border-none shadow-glow"
                  >
                    <div className="space-y-6">
                      {bestPractices.practices.map((practice, index) => (
                        <div key={index} className="p-4 bg-dark-200 rounded-lg">
                          <h3 className="text-lg font-medium text-primary-400 mb-2">
                            {practice.title}
                          </h3>
                          <div className="prose dark:prose-invert max-w-none text-gray-300 text-sm">
                            <p>{practice.description}</p>

                            {practice.examples &&
                              practice.examples.length > 0 && (
                                <div className="mt-3">
                                  <h4 className="text-sm font-medium text-secondary-400 mb-2">
                                    Examples:
                                  </h4>
                                  <div className="space-y-2">
                                    {practice.examples.map((example, i) => (
                                      <div
                                        key={i}
                                        className="bg-dark-300 p-3 rounded-md text-gray-300 text-xs overflow-auto"
                                      >
                                        <pre>{example}</pre>
                                      </div>
                                    ))}
                                  </div>
                                </div>
                              )}
                          </div>
                        </div>
                      ))}
                    </div>
                  </Card>
                </motion.div>
              ) : null}

              {/* Issues */}
              {bestPractices.issues && bestPractices.issues.length > 0 ? (
                <motion.div
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.5, delay: 0.2 }}
                >
                  <Card
                    title="Potential Issues & Improvements"
                    className="bg-gradient-to-b from-dark-100/70 to-dark-100 border-none shadow-glow"
                  >
                    <div className="overflow-x-auto">
                      <table className="min-w-full divide-y divide-gray-800">
                        <thead>
                          <tr>
                            <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                              Issue
                            </th>
                            <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                              Location
                            </th>
                            <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                              Severity
                            </th>
                            <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
                              Suggestion
                            </th>
                          </tr>
                        </thead>
                        <tbody className="divide-y divide-gray-800">
                          {bestPractices.issues.map((issue, index) => (
                            <tr key={index} className="hover:bg-dark-200">
                              <td className="px-6 py-4 text-sm text-gray-300">
                                {issue.description}
                              </td>
                              <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-400">
                                {issue.path ? (
                                  <>
                                    <code className="bg-dark-300 px-2 py-1 rounded">
                                      {issue.path}
                                    </code>
                                    {issue.line > 0 && (
                                      <span className="ml-2">
                                        line {issue.line}
                                      </span>
                                    )}
                                  </>
                                ) : (
                                  <span className="italic">Global</span>
                                )}
                              </td>
                              <td className="px-6 py-4 whitespace-nowrap">
                                <span
                                  className={`px-2 py-1 text-xs rounded-full 
                                    ${
                                      issue.severity === "high"
                                        ? "bg-warning-900/30 text-warning-300"
                                        : issue.severity === "medium"
                                        ? "bg-accent-900/30 text-accent-300"
                                        : "bg-primary-900/30 text-primary-300"
                                    }`}
                                >
                                  {issue.severity || "low"}
                                </span>
                              </td>
                              <td className="px-6 py-4 text-sm text-gray-300">
                                {issue.suggestion || "No suggestion provided"}
                              </td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                  </Card>
                </motion.div>
              ) : null}
            </div>
          )}

          {!bestPractices && !practicesLoading && (
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
                    d="M12 6V4m0 2a2 2 0 100 4m0-4a2 2 0 110 4m-6 8a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4m6 6v10m6-2a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4"
                  />
                </svg>
                <h2 className="text-2xl font-bold text-white mb-3">
                  Discover Coding Patterns & Conventions
                </h2>
                <p className="text-gray-400 mb-6 max-w-md mx-auto">
                  Generate a guide to the coding patterns, conventions, and best
                  practices used in the repository. Perfect for newcomers to
                  follow team standards.
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
                        <path
                          fillRule="evenodd"
                          d="M3 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1z"
                          clipRule="evenodd"
                        />
                      </svg>
                      Style Guides
                    </h3>
                    <p className="text-sm text-gray-400">
                      Understand naming conventions and code formatting
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
                      Coding Patterns
                    </h3>
                    <p className="text-sm text-gray-400">
                      Learn the design patterns and approaches used in the
                      project
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
                        <path
                          fillRule="evenodd"
                          d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z"
                          clipRule="evenodd"
                        />
                      </svg>
                      Potential Issues
                    </h3>
                    <p className="text-sm text-gray-400">
                      Identify inconsistencies and areas for improvement
                    </p>
                  </div>
                </div>
              </div>
            </motion.div>
          )}
        </div>
      )}
    </div>
  );
};

export default ArchitectureAndPractices;
