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

const CodeComments = () => {
  const location = useLocation();
  const navigate = useNavigate();
  const [url, setUrl] = useState(location.state?.url || "");
  const [branch, setBranch] = useState("");
  const [filePath, setFilePath] = useState("");
  const [loading, setLoading] = useState(false);
  const [originalCode, setOriginalCode] = useState("");
  const [commentedCode, setCommentedCode] = useState("");
  const [pushing, setPushing] = useState(false);
  const [repoInfo, setRepoInfo] = useState(null);
  const [errors, setErrors] = useState({});
  const [fileLanguage, setFileLanguage] = useState("");

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

  const fetchFileContent = async () => {
    if (!url || !filePath) {
      setErrors({
        url: !url ? "Please enter a GitHub repository URL" : "",
        filePath: !filePath ? "Please enter a file path" : "",
      });
      return;
    }

    setLoading(true);
    setErrors({});

    try {
      const branchToUse =
        branch || (repoInfo ? repoInfo.default_branch : "main");
      const response = await repositoryService.getFile(
        url,
        filePath,
        branchToUse
      );
      setOriginalCode(response.data.content);

      // Detect language from file extension
      const fileExt = filePath.split(".").pop().toLowerCase();
      setFileLanguage(detectLanguage(fileExt));

      toast.success("File content loaded successfully");

      // Fetch repo info if not already loaded
      if (!repoInfo) {
        await fetchRepositoryInfo(url);
      }
    } catch (error) {
      console.error("Error fetching file:", error);

      let errorMessage = "Failed to fetch file content";
      if (error.response?.data?.error) {
        errorMessage = error.response.data.error;
      }

      setErrors({ filePath: errorMessage });
      toast.error(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  const generateComments = async () => {
    if (!originalCode) {
      toast.error("Please fetch file content first");
      return;
    }

    setLoading(true);

    try {
      const response = await generateService.comments(url, filePath, branch);
      setCommentedCode(response.data.content);
      toast.success("Comments generated successfully");
    } catch (error) {
      console.error("Error generating comments:", error);

      let errorMessage = "Failed to generate comments";
      if (error.response?.data?.error) {
        errorMessage = error.response.data.error;
      }

      toast.error(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  const pushCommentedCode = async () => {
    if (!commentedCode || !url || !filePath) {
      toast.error("Please generate commented code first");
      return;
    }

    setPushing(true);

    try {
      const branchToUse =
        branch || (repoInfo ? repoInfo.default_branch : "main");
      await pushService.file(
        url,
        filePath,
        commentedCode,
        `Add comments to ${filePath} via GitHub Agent`,
        branchToUse
      );
      toast.success("Commented code pushed to repository successfully");
    } catch (error) {
      console.error("Error pushing code:", error);

      let errorMessage = "Failed to push code to repository";
      if (error.response?.data?.error) {
        errorMessage = error.response.data.error;
      }

      toast.error(errorMessage);
    } finally {
      setPushing(false);
    }
  };

  const detectLanguage = (fileExtension) => {
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

    return extMap[fileExtension] || "text";
  };

  return (
    <div>
      <motion.div
        initial={{ opacity: 0, y: -20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.3 }}
      >
        <h1 className="text-2xl font-bold text-gray-900 dark:text-white mb-2">
          Add Code Comments
        </h1>
        <p className="text-gray-600 dark:text-gray-300 mb-6">
          Enhance your code with detailed and helpful comments
        </p>
      </motion.div>

      <div className="mb-6">
        <Card title="File Information">
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

          <div className="mt-4 flex flex-col md:flex-row gap-4">
            <div className="flex-1">
              <Input
                label="File Path"
                placeholder="src/components/App.js"
                value={filePath}
                onChange={(e) => setFilePath(e.target.value)}
                error={errors.filePath}
                required
              />
            </div>
            <div className="flex items-end">
              <Button
                onClick={fetchFileContent}
                isLoading={loading && !commentedCode}
                disabled={loading}
              >
                Fetch File
              </Button>
            </div>
          </div>
        </Card>
      </div>

      {originalCode && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <div>
            <Card title="Original Code">
              <CodeEditor
                code={originalCode}
                language={fileLanguage}
                readOnly={true}
                lineNumbers={true}
                showCopyButton={true}
                fileName={filePath.split("/").pop()}
                style={{ minHeight: "400px" }}
              />
            </Card>
          </div>

          <div>
            <Card
              title="Commented Code"
              headerAction={
                <div className="flex space-x-2">
                  {!commentedCode && (
                    <Button
                      variant="primary"
                      size="sm"
                      onClick={generateComments}
                      isLoading={loading}
                      disabled={loading}
                    >
                      Generate Comments
                    </Button>
                  )}

                  {commentedCode && (
                    <Button
                      variant="secondary"
                      size="sm"
                      onClick={pushCommentedCode}
                      isLoading={pushing}
                      disabled={pushing}
                    >
                      Push to Repository
                    </Button>
                  )}
                </div>
              }
            >
              {commentedCode ? (
                <CodeEditor
                  code={commentedCode}
                  language={fileLanguage}
                  readOnly={false}
                  onChange={setCommentedCode}
                  lineNumbers={true}
                  showCopyButton={true}
                  fileName={filePath.split("/").pop()}
                  style={{ minHeight: "400px" }}
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
                      d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
                    />
                  </svg>
                  <p>Click "Generate Comments" to add comments to your code</p>
                </div>
              )}
            </Card>
          </div>
        </div>
      )}
    </div>
  );
};

export default CodeComments;
