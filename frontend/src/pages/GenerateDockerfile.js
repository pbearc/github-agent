import React, { useState, useEffect } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import Card from "../components/ui/Card";
import Button from "../components/ui/Button";
import Input from "../components/ui/Input";
import Select from "../components/ui/Select";
import CodeEditor from "../components/ui/CodeEditor";
import { toast } from "sonner";
import { motion } from "framer-motion";
import {
  generateService,
  pushService,
  repositoryService,
} from "../services/api";

const GenerateDockerfile = () => {
  const location = useLocation();
  const navigate = useNavigate();
  const [url, setUrl] = useState(location.state?.url || "");
  const [branch, setBranch] = useState("");
  const [language, setLanguage] = useState("");
  const [loading, setLoading] = useState(false);
  const [dockerfileContent, setDockerfileContent] = useState("");
  const [pushing, setPushing] = useState(false);
  const [repoInfo, setRepoInfo] = useState(null);
  const [errors, setErrors] = useState({});

  // Language options for the dropdown
  const languageOptions = [
    { value: "", label: "Auto-detect" },
    { value: "javascript", label: "JavaScript" },
    { value: "typescript", label: "TypeScript" },
    { value: "python", label: "Python" },
    { value: "go", label: "Go" },
    { value: "java", label: "Java" },
    { value: "ruby", label: "Ruby" },
    { value: "php", label: "PHP" },
    { value: "rust", label: "Rust" },
    { value: "csharp", label: "C#" },
  ];

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

      // Auto-detect language if not specified
      if (!language && response.data.language) {
        setLanguage(response.data.language.toLowerCase());
      }
    } catch (error) {
      console.error("Error fetching repository info:", error);
    }
  };

  const generateDockerfile = async (e) => {
    e.preventDefault();

    if (!url) {
      setErrors({ url: "Please enter a GitHub repository URL" });
      return;
    }

    setLoading(true);
    setErrors({});

    try {
      const response = await generateService.dockerfile(url, language, branch);
      setDockerfileContent(response.data.content);
      toast.success("Dockerfile generated successfully");

      // Fetch repo info if not already loaded
      if (!repoInfo) {
        await fetchRepositoryInfo(url);
      }
    } catch (error) {
      console.error("Error generating Dockerfile:", error);

      let errorMessage = "Failed to generate Dockerfile";
      if (error.response?.data?.error) {
        errorMessage = error.response.data.error;
      }

      setErrors({ generate: errorMessage });
      toast.error(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  const pushDockerfile = async () => {
    if (!dockerfileContent || !url || !branch) {
      toast.error("Please generate Dockerfile content first");
      return;
    }

    setPushing(true);

    try {
      await pushService.file(
        url,
        "Dockerfile",
        dockerfileContent,
        "Add Dockerfile via GitHub Agent",
        branch
      );
      toast.success("Dockerfile pushed to repository successfully");
    } catch (error) {
      console.error("Error pushing Dockerfile:", error);

      let errorMessage = "Failed to push Dockerfile to repository";
      if (error.response?.data?.error) {
        errorMessage = error.response.data.error;
      }

      toast.error(errorMessage);
    } finally {
      setPushing(false);
    }
  };

  return (
    <div>
      <motion.div
        initial={{ opacity: 0, y: -20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.3 }}
      >
        <h1 className="text-2xl font-bold text-gray-900 dark:text-white mb-2">
          Generate Dockerfile
        </h1>
        <p className="text-gray-600 dark:text-gray-300 mb-6">
          Create a Dockerfile optimized for your GitHub repository
        </p>
      </motion.div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-1">
          <Card title="Repository Information">
            <form onSubmit={generateDockerfile}>
              <Input
                label="GitHub Repository URL"
                placeholder="https://github.com/username/repository"
                value={url}
                onChange={(e) => setUrl(e.target.value)}
                error={errors.url}
                required
                className="mb-4"
              />

              <Input
                label="Branch"
                placeholder="main"
                value={branch}
                onChange={(e) => setBranch(e.target.value)}
                className="mb-4"
              />

              <Select
                label="Language"
                value={language}
                onChange={(e) => setLanguage(e.target.value)}
                options={languageOptions}
                placeholder="Select language or auto-detect"
                className="mb-6"
                helpText="If not specified, language will be auto-detected from repository"
              />

              <Button
                type="submit"
                fullWidth
                isLoading={loading}
                disabled={loading}
              >
                Generate Dockerfile
              </Button>

              {errors.generate && (
                <p className="mt-2 text-sm text-red-600 dark:text-red-500">
                  {errors.generate}
                </p>
              )}
            </form>
          </Card>

          {repoInfo && (
            <Card title="Repository Details" className="mt-6">
              <div className="space-y-3">
                <div>
                  <h3 className="text-sm font-medium text-gray-500 dark:text-gray-400">
                    Repository
                  </h3>
                  <p className="mt-1 text-md text-gray-800 dark:text-gray-200">
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

                <div>
                  <h3 className="text-sm font-medium text-gray-500 dark:text-gray-400">
                    Primary Language
                  </h3>
                  <p className="mt-1 text-md text-gray-800 dark:text-gray-200">
                    {repoInfo.language || "Not specified"}
                  </p>
                </div>
              </div>
            </Card>
          )}
        </div>

        <div className="lg:col-span-2">
          <Card
            title="Dockerfile Preview"
            headerAction={
              dockerfileContent && (
                <Button
                  variant="primary"
                  size="sm"
                  onClick={pushDockerfile}
                  isLoading={pushing}
                  disabled={pushing}
                >
                  Push to Repository
                </Button>
              )
            }
          >
            {dockerfileContent ? (
              <CodeEditor
                code={dockerfileContent}
                language="dockerfile"
                readOnly={false}
                onChange={setDockerfileContent}
                fileName="Dockerfile"
                showCopyButton={true}
                style={{ minHeight: "400px", height: "100%" }}
              />
            ) : (
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
                    d="M20 7l-8-4-8 4m16 0l-8 4m0 0l-8-4m8 4v10"
                  />
                </svg>
                <p>
                  Enter a GitHub repository URL and click "Generate Dockerfile"
                  to create content
                </p>
              </div>
            )}
          </Card>
        </div>
      </div>
    </div>
  );
};

export default GenerateDockerfile;
