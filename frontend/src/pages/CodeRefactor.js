import React, { useState, useEffect } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import Card from "../components/ui/Card";
import Button from "../components/ui/Button";
import Input from "../components/ui/Input";
import TextArea from "../components/ui/TextArea";
import CodeEditor from "../components/ui/CodeEditor";
import { toast } from "sonner";
import { motion } from "framer-motion";
import {
  generateService,
  pushService,
  repositoryService,
} from "../services/api";

const CodeRefactor = () => {
  const location = useLocation();
  const navigate = useNavigate();
  const [url, setUrl] = useState(location.state?.url || "");
  const [branch, setBranch] = useState("");
  const [filePath, setFilePath] = useState("");
  const [instructions, setInstructions] = useState("");
  const [loading, setLoading] = useState(false);
  const [originalCode, setOriginalCode] = useState("");
  const [refactoredCode, setRefactoredCode] = useState("");
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

  const refactorCode = async () => {
    if (!originalCode) {
      toast.error("Please fetch file content first");
      return;
    }

    if (!instructions) {
      setErrors({ instructions: "Please provide refactoring instructions" });
      return;
    }

    setLoading(true);
    setErrors({});

    try {
      const response = await generateService.refactor(
        url,
        filePath,
        instructions,
        branch
      );
      setRefactoredCode(response.data.content);
      toast.success("Code refactored successfully");
    } catch (error) {
      console.error("Error refactoring code:", error);

      let errorMessage = "Failed to refactor code";
      if (error.response?.data?.error) {
        errorMessage = error.response.data.error;
      }

      toast.error(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  const pushRefactoredCode = async () => {
    if (!refactoredCode || !url || !filePath) {
      toast.error("Please refactor code first");
      return;
    }

    setPushing(true);

    try {
      const branchToUse =
        branch || (repoInfo ? repoInfo.default_branch : "main");
      await pushService.file(
        url,
        filePath,
        refactoredCode,
        `Refactor ${filePath} via GitHub Agent`,
        branchToUse
      );
      toast.success("Refactored code pushed to repository successfully");
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

  // Sample refactoring instructions
  const sampleInstructions = [
    "Improve performance by using more efficient algorithms",
    "Add error handling and input validation",
    "Convert to TypeScript with proper type definitions",
    "Refactor to use async/await instead of promises",
    "Implement the observer pattern for better event handling",
    "Split large functions into smaller, more manageable ones",
    "Apply SOLID principles to improve code maintainability",
  ];

  return (
    <div>
      <motion.div
        initial={{ opacity: 0, y: -20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.3 }}
      >
        <h1 className="text-2xl font-bold text-gray-900 dark:text-white mb-2">
          Refactor Code
        </h1>
        <p className="text-gray-600 dark:text-gray-300 mb-6">
          Improve your code's quality, readability, and performance
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
                isLoading={loading && !refactoredCode}
                disabled={loading}
              >
                Fetch File
              </Button>
            </div>
          </div>
        </Card>
      </div>

      {originalCode && (
        <>
          <div className="mb-6">
            <Card title="Refactoring Instructions">
              <TextArea
                label="Instructions"
                placeholder="Describe how you want the code to be refactored..."
                value={instructions}
                onChange={(e) => setInstructions(e.target.value)}
                error={errors.instructions}
                rows={3}
                required
              />

              <div className="mt-4">
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Sample Instructions
                </label>
                <div className="flex flex-wrap gap-2">
                  {sampleInstructions.map((sample, index) => (
                    <Button
                      key={index}
                      variant="outline"
                      size="xs"
                      onClick={() => setInstructions(sample)}
                    >
                      {sample}
                    </Button>
                  ))}
                </div>
              </div>

              <div className="mt-6">
                <Button
                  onClick={refactorCode}
                  isLoading={loading}
                  disabled={loading}
                  fullWidth
                >
                  Refactor Code
                </Button>
              </div>
            </Card>
          </div>

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
                title="Refactored Code"
                headerAction={
                  refactoredCode && (
                    <Button
                      variant="primary"
                      size="sm"
                      onClick={pushRefactoredCode}
                      isLoading={pushing}
                      disabled={pushing}
                    >
                      Push to Repository
                    </Button>
                  )
                }
              >
                {refactoredCode ? (
                  <CodeEditor
                    code={refactoredCode}
                    language={fileLanguage}
                    readOnly={false}
                    onChange={setRefactoredCode}
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
                        d="M8 7v8a2 2 0 002 2h6M8 7V5a2 2 0 012-2h4.586a1 1 0 01.707.293l4.414 4.414a1 1 0 01.293.707V15a2 2 0 01-2 2h-2M8 7H6a2 2 0 00-2 2v10a2 2 0 002 2h8a2 2 0 002-2v-2"
                      />
                    </svg>
                    <p>
                      Provide instructions and click "Refactor Code" to see the
                      results
                    </p>
                  </div>
                )}
              </Card>
            </div>
          </div>
        </>
      )}
    </div>
  );
};

export default CodeRefactor;
