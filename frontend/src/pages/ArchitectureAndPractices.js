import React, { useState, useEffect } from "react";
import { useLocation } from "react-router-dom";
import Card from "../components/ui/Card";
import Button from "../components/ui/Button";
import Input from "../components/ui/Input";
import { toast } from "sonner";
import { motion } from "framer-motion";
import { repositoryService, navigatorService } from "../services/api";
import CodebaseGraph from "../components/graph/CodebaseGraph";
import CodebaseGraph3D from "../components/graph/ForceGraph3D";
import ReactMarkdown from "react-markdown";

const ArchitectureAndPractices = () => {
  const location = useLocation();
  const [url, setUrl] = useState(location.state?.url || "");
  const [branch, setBranch] = useState("");
  const [loading, setLoading] = useState(false);
  const [graphLoading, setGraphLoading] = useState(false);
  const [repoInfo, setRepoInfo] = useState(null);
  const [architecture, setArchitecture] = useState(null);
  const [graphData, setGraphData] = useState(null);
  const [errors, setErrors] = useState({});
  const [detailLevel, setDetailLevel] = useState("medium");
  const [focusPaths, setFocusPaths] = useState([]);
  const [focusPathInput, setFocusPathInput] = useState("");
  const [visualizationType, setVisualizationType] = useState("2d");
  const [explanation, setExplanation] = useState("");

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

      // Fetch repo info if not already loaded
      if (!repoInfo) {
        await fetchRepositoryInfo(url);
      }

      // Fetch graph data after architecture is generated
      if (response.data) {
        fetchGraphData(url, branchToUse);
      }

      toast.success("Architecture visualization complete!");
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

  const fetchGraphData = async (repoUrl, branchToUse) => {
    if (!repoUrl) return;

    setGraphLoading(true);
    setGraphData(null);
    setExplanation("");
    try {
      toast.info("Fetching graph data and generating explanation...", {
        duration: 3000,
      });

      const response =
        await navigatorService.getArchitectureGraphAndExplanation(
          repoUrl,
          branchToUse
        );
      console.log(
        "Fetched graph data and explanation from API:",
        response.data
      );

      if (
        response.data &&
        response.data.graph &&
        response.data.graph.nodes &&
        (response.data.graph.edges || response.data.graph.relationships) &&
        response.data.explanation
      ) {
        setGraphData(response.data.graph);
        setExplanation(response.data.explanation);
        toast.success("Graph data and explanation loaded successfully");
      } else {
        console.error(
          "API returned invalid graph data or explanation structure",
          response.data
        );
        throw new Error("Invalid graph data or explanation structure from API");
      }
    } catch (error) {
      console.error("Error processing graph data:", error);
      let errorMessage = "Failed to load graph data and explanation";
      if (error.response?.data?.error) {
        errorMessage = error.response.data.error;
      } else if (error.message) {
        errorMessage = error.message;
      }
      toast.error(errorMessage);
      setGraphData(null);
      setExplanation("");
    } finally {
      setGraphLoading(false);
    }
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
              Codebase Structure & Architecture
            </h1>
            <p className="text-gray-600 dark:text-gray-300">
              Visualize architecture in the repository
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
              <span>Understand architecture</span>
            </div>
          </div>
        </div>
      </motion.div>

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
            Higher detail levels provide more components but may be more complex
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
            className="bg-gradient-to-r from-primary-600 to-secondary-600 hover:from-primary-500 hover:to-secondary-500 shadow-glow w-full"
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

      {/* Loading indicators */}
      {loading && (
        <div className="flex justify-center my-6">
          <div className="text-center">
            <div className="inline-block animate-spin rounded-full h-8 w-8 border-t-2 border-b-2 border-primary-500 mb-2"></div>
            <p className="text-gray-400">
              Generating architecture visualization...
            </p>
          </div>
        </div>
      )}

      {/* Results Section */}
      {architecture && (
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5, delay: 0.2 }}
        >
          {/* Explanation Section */}
          {explanation && (
            <Card
              title="Architecture Explanation"
              className="bg-gradient-to-b from-dark-100/70 to-dark-100 border-none shadow-glow mb-6"
            >
              <div className="prose prose-sm max-w-none prose-dark prose-invert text-white">
                <ReactMarkdown>{explanation}</ReactMarkdown>
              </div>
            </Card>
          )}

          {/* Architecture Visualization Section */}
          <Card
            title="Architecture Diagram"
            className="bg-gradient-to-b from-dark-100/70 to-dark-100 border-none shadow-glow relative"
          >
            {/* Loading overlay */}
            {graphLoading && (
              <div className="absolute inset-0 bg-gray-800 bg-opacity-70 flex items-center justify-center z-10 rounded-lg">
                <div className="text-white text-center">
                  <svg
                    className="animate-spin h-8 w-8 text-white mx-auto mb-2"
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
                  Loading Graph Data...
                </div>
              </div>
            )}

            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-200 mb-1">
                Visualization Type
              </label>
              <div className="flex space-x-4">
                <label className="flex items-center">
                  <input
                    type="radio"
                    name="visualizationType"
                    value="2d"
                    checked={visualizationType === "2d"}
                    onChange={() => setVisualizationType("2d")}
                    className="h-4 w-4 text-primary-600 focus:ring-primary-500 border-gray-700 bg-dark-200"
                  />
                  <span className="ml-2 text-gray-300">2D Graph</span>
                </label>
                <label className="flex items-center">
                  <input
                    type="radio"
                    name="visualizationType"
                    value="3d"
                    checked={visualizationType === "3d"}
                    onChange={() => setVisualizationType("3d")}
                    className="h-4 w-4 text-primary-600 focus:ring-primary-500 border-gray-700 bg-dark-200"
                  />
                  <span className="ml-2 text-gray-300">3D Graph</span>
                </label>
              </div>
            </div>

            {!graphLoading &&
            (!graphData || !graphData.nodes || graphData.nodes.length === 0) ? (
              <div className="flex flex-col items-center justify-center h-96">
                <p className="text-gray-400 mb-4">
                  No graph data available for visualization
                </p>
                <Button
                  type="button"
                  onClick={() =>
                    fetchGraphData(
                      url,
                      branch || (repoInfo ? repoInfo.default_branch : "main")
                    )
                  }
                  disabled={!url}
                  className="bg-gradient-to-r from-primary-600 to-secondary-600 hover:from-primary-500 hover:to-secondary-500"
                  size="sm"
                >
                  Reload Graph Data
                </Button>
              </div>
            ) : (
              <div
                className="bg-gray-800 rounded-lg overflow-hidden"
                style={{ height: "600px" }}
              >
                {visualizationType === "2d" ? (
                  <CodebaseGraph
                    graphData={
                      graphData || {
                        nodes: architecture?.diagram_data?.nodes || [],
                        edges: architecture?.diagram_data?.edges || [],
                        relationships:
                          architecture?.diagram_data?.relationships,
                      }
                    }
                  />
                ) : (
                  <CodebaseGraph3D
                    graphData={
                      graphData || {
                        nodes: architecture?.diagram_data?.nodes || [],
                        edges: architecture?.diagram_data?.edges || [],
                        relationships:
                          architecture?.diagram_data?.relationships,
                      }
                    }
                  />
                )}
              </div>
            )}

            <p className="text-sm text-gray-400 mt-4">
              Note: The diagram shows how components in the codebase relate to
              each other. Hover over nodes to see details.
              {visualizationType === "3d" &&
                " Use mouse to rotate the view and scroll to zoom."}
            </p>
          </Card>
        </motion.div>
      )}
    </div>
  );
};

export default ArchitectureAndPractices;
