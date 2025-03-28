import React, { useState, useEffect } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import Card from "../components/ui/Card";
import Button from "../components/ui/Button";
import Input from "../components/ui/Input";
import CodeEditor from "../components/ui/CodeEditor";
import { toast } from "sonner";
import { motion } from "framer-motion";
import {
  generateService,
  pushService,
  repositoryService,
} from "../services/api";

const GenerateReadme = () => {
  const location = useLocation();
  const navigate = useNavigate();
  const [url, setUrl] = useState(location.state?.url || "");
  const [branch, setBranch] = useState("");
  const [includeFiles, setIncludeFiles] = useState(true);
  const [loading, setLoading] = useState(false);
  const [readmeContent, setReadmeContent] = useState("");
  const [pushing, setPushing] = useState(false);
  const [repoInfo, setRepoInfo] = useState(null);
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

  const generateReadme = async (e) => {
    e.preventDefault();

    if (!url) {
      setErrors({ url: "Please enter a GitHub repository URL" });
      return;
    }

    setLoading(true);
    setErrors({});

    try {
      const response = await generateService.readme(url, includeFiles, branch);
      setReadmeContent(response.data.content);
      toast.success("README generated successfully");

      // Fetch repo info if not already loaded
      if (!repoInfo) {
        await fetchRepositoryInfo(url);
      }
    } catch (error) {
      console.error("Error generating README:", error);

      let errorMessage = "Failed to generate README";
      if (error.response?.data?.error) {
        errorMessage = error.response.data.error;
      }

      setErrors({ generate: errorMessage });
      toast.error(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  const pushReadme = async () => {
    if (!readmeContent || !url || !branch) {
      toast.error("Please generate README content first");
      return;
    }

    setPushing(true);

    try {
      await pushService.file(
        url,
        "README.md",
        readmeContent,
        "Add README.md via GitHub Agent",
        branch
      );
      toast.success("README pushed to repository successfully");
    } catch (error) {
      console.error("Error pushing README:", error);

      let errorMessage = "Failed to push README to repository";
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
          Generate README
        </h1>
        <p className="text-gray-600 dark:text-gray-300 mb-6">
          Create a professional README.md file for your GitHub repository
        </p>
      </motion.div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-1">
          <Card title="Repository Information">
            <form onSubmit={generateReadme}>
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

              <div className="flex items-center mb-6">
                <input
                  id="include-files"
                  type="checkbox"
                  checked={includeFiles}
                  onChange={(e) => setIncludeFiles(e.target.checked)}
                  className="w-4 h-4 text-primary-600 border-gray-300 rounded focus:ring-primary-500 dark:focus:ring-primary-600 dark:ring-offset-gray-800"
                />
                <label
                  htmlFor="include-files"
                  className="ml-2 text-sm text-gray-700 dark:text-gray-300"
                >
                  Include file structure in README
                </label>
              </div>

              <Button
                type="submit"
                fullWidth
                isLoading={loading}
                disabled={loading}
              >
                Generate README
              </Button>

              {errors.generate && (
                <p className="mt-2 text-sm text-red-600 dark:text-red-500">
                  {errors.generate}
                </p>
              )}
            </form>
          </Card>
        </div>

        <div className="lg:col-span-2">
          <Card
            title="README Preview"
            headerAction={
              readmeContent && (
                <Button
                  variant="primary"
                  size="sm"
                  onClick={pushReadme}
                  isLoading={pushing}
                  disabled={pushing}
                >
                  Push to Repository
                </Button>
              )
            }
          >
            {readmeContent ? (
              <CodeEditor
                code={readmeContent}
                language="markdown"
                readOnly={false}
                onChange={setReadmeContent}
                fileName="README.md"
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
                    d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                  />
                </svg>
                <p>
                  Enter a GitHub repository URL and click "Generate README" to
                  create content
                </p>
              </div>
            )}
          </Card>
        </div>
      </div>
    </div>
  );
};

export default GenerateReadme;
