import React, { useState } from "react";
import { useLocation } from "react-router-dom";
import Card from "../components/ui/Card";
import Button from "../components/ui/Button";
import Input from "../components/ui/Input";
import CodeEditor from "../components/ui/CodeEditor";
import { toast } from "sonner";
import { motion } from "framer-motion";
import { prSummaryService, repositoryService } from "../services/api";
import Badge from "../components/ui/Badge";

const PRSummary = () => {
  const location = useLocation();
  const [url, setUrl] = useState(location.state?.url || "");
  const [loading, setLoading] = useState(false);
  const [summary, setSummary] = useState(null);
  const [selectedGroup, setSelectedGroup] = useState(null);
  const [errors, setErrors] = useState({});

  const handleGetSummary = async (e) => {
    e.preventDefault();

    if (!url) {
      setErrors({ url: "Please enter a GitHub PR URL" });
      return;
    }

    setLoading(true);
    setErrors({});
    setSummary(null);
    setSelectedGroup(null);

    try {
      const response = await prSummaryService.getSummary(url);
      setSummary(response.data.summary);

      // Select the first file group by default if available
      if (
        response.data.summary.file_groups &&
        response.data.summary.file_groups.length > 0
      ) {
        setSelectedGroup(response.data.summary.file_groups[0]);
      }

      toast.success("PR Summary generated successfully");
    } catch (error) {
      console.error("Error generating PR summary:", error);

      let errorMessage = "Failed to generate PR summary";
      if (error.response?.data?.error) {
        errorMessage = error.response.data.error;
      }

      toast.error(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  // Format repository URL for linking
  const getRepoUrl = (repository) => {
    if (!repository) return "#";
    return `https://github.com/${repository}`;
  };

  // Sort file groups by importance
  const getSortedFileGroups = () => {
    if (!summary || !summary.file_groups) return [];

    return [...summary.file_groups].sort((a, b) => b.importance - a.importance);
  };

  return (
    <div>
      <motion.div
        initial={{ opacity: 0, y: -20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.3 }}
      >
        <h1 className="text-2xl font-bold text-gray-900 dark:text-white mb-2">
          Pull Request Summary
        </h1>
        <p className="text-gray-600 dark:text-gray-300 mb-6">
          Generate a comprehensive summary of any GitHub pull request
        </p>
      </motion.div>

      <div className="mb-6">
        <Card title="PR Information">
          <form onSubmit={handleGetSummary}>
            <Input
              label="GitHub Pull Request URL"
              placeholder="https://github.com/username/repository/pull/123"
              value={url}
              onChange={(e) => setUrl(e.target.value)}
              error={errors.url}
              required
              className="mb-6"
            />

            <Button type="submit" isLoading={loading} disabled={loading}>
              {loading ? "Analyzing PR..." : "Generate Summary"}
            </Button>
          </form>
        </Card>
      </div>

      {summary && (
        <div className="space-y-6">
          {/* PR Header */}
          <Card>
            <div className="flex flex-col md:flex-row md:items-center justify-between mb-4">
              <div>
                <h2 className="text-xl font-bold text-gray-900 dark:text-white">
                  {summary.title}
                </h2>
                <div className="mt-1">
                  <a
                    href={summary.pr_url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-sm text-primary-600 dark:text-primary-400 hover:underline"
                  >
                    {summary.repository} #{summary.pr_number}
                  </a>
                  <span className="mx-2 text-gray-400">•</span>
                  <span className="text-sm text-gray-600 dark:text-gray-400">
                    By {summary.author}
                  </span>
                </div>
              </div>

              <div className="mt-3 md:mt-0 flex flex-wrap gap-2">
                <Badge color="blue">
                  {summary.changed_files} file
                  {summary.changed_files !== 1 ? "s" : ""}
                </Badge>
                <Badge color="green">+{summary.additions}</Badge>
                <Badge color="red">-{summary.deletions}</Badge>
              </div>
            </div>

            <div className="border-t border-gray-200 dark:border-gray-700 pt-4">
              <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-2">
                Description
              </h3>
              <p className="text-gray-700 dark:text-gray-300">
                {summary.description}
              </p>
            </div>
          </Card>

          {/* Main content grid */}
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            {/* Left column */}
            <div className="space-y-6">
              {/* Main Points */}
              <Card title="Main Points">
                <ul className="list-disc list-inside space-y-2 text-gray-700 dark:text-gray-300">
                  {summary.main_points.map((point, index) => (
                    <li key={index}>{point}</li>
                  ))}
                </ul>
              </Card>

              {/* Key Changes */}
              <Card title="Key Technical Changes">
                <ul className="list-disc list-inside space-y-2 text-gray-700 dark:text-gray-300">
                  {summary.key_changes.map((change, index) => (
                    <li key={index}>{change}</li>
                  ))}
                </ul>
              </Card>

              {/* Impact Assessment */}
              <Card title="Potential Impact">
                <div className="prose dark:prose-invert">
                  <p className="text-gray-700 dark:text-gray-300">
                    {summary.potential_impact}
                  </p>
                </div>
              </Card>
            </div>

            {/* Right column */}
            <div className="lg:col-span-2 space-y-6">
              {/* File Groups */}
              <Card title="File Groups">
                <div className="space-y-4">
                  {getSortedFileGroups().map((group) => (
                    <div
                      key={group.name}
                      className={`p-4 rounded-lg cursor-pointer transition-colors ${
                        selectedGroup?.name === group.name
                          ? "bg-primary-50 dark:bg-primary-900/30 border border-primary-200 dark:border-primary-800"
                          : "bg-gray-50 dark:bg-gray-800/50 hover:bg-gray-100 dark:hover:bg-gray-800 border border-gray-200 dark:border-gray-700"
                      }`}
                      onClick={() => setSelectedGroup(group)}
                    >
                      <div className="flex justify-between items-center">
                        <h3 className="text-lg font-medium text-gray-900 dark:text-white">
                          {group.name}
                        </h3>
                        <Badge
                          color={
                            group.importance > 7
                              ? "red"
                              : group.importance > 4
                              ? "yellow"
                              : "blue"
                          }
                        >
                          Priority: {group.importance}/10
                        </Badge>
                      </div>
                      <p className="mt-1 text-gray-600 dark:text-gray-400">
                        {group.description}
                      </p>
                      <div className="mt-3">
                        <span className="text-sm text-gray-500 dark:text-gray-400">
                          {group.files?.length || 0} file
                          {(group.files?.length || 0) !== 1 ? "s" : ""}
                        </span>
                      </div>
                    </div>
                  ))}
                </div>
              </Card>

              {/* File List */}
              {selectedGroup &&
                selectedGroup.files &&
                selectedGroup.files.length > 0 && (
                  <Card title={`Files in ${selectedGroup.name}`}>
                    <ul className="divide-y divide-gray-200 dark:divide-gray-700">
                      {selectedGroup.files.map((file, index) => (
                        <li key={index} className="py-3">
                          <div className="flex items-center space-x-4">
                            <svg
                              xmlns="http://www.w3.org/2000/svg"
                              className="h-5 w-5 text-gray-500 dark:text-gray-400"
                              viewBox="0 0 20 20"
                              fill="currentColor"
                            >
                              <path
                                fillRule="evenodd"
                                d="M4 4a2 2 0 012-2h4.586A2 2 0 0112 2.586L15.414 6A2 2 0 0116 7.414V16a2 2 0 01-2 2H6a2 2 0 01-2-2V4z"
                                clipRule="evenodd"
                              />
                            </svg>
                            <a
                              href={`${getRepoUrl(
                                summary.repository
                              )}/blob/HEAD/${file}`}
                              target="_blank"
                              rel="noopener noreferrer"
                              className="text-primary-600 dark:text-primary-400 hover:underline break-all"
                            >
                              {file}
                            </a>
                          </div>
                        </li>
                      ))}
                    </ul>
                  </Card>
                )}

              {/* Technical Details */}
              <Card title="Additional Technical Details">
                <div className="prose dark:prose-invert max-w-none">
                  <p className="text-gray-700 dark:text-gray-300 whitespace-pre-line">
                    {summary.technical_details}
                  </p>
                </div>
              </Card>

              {/* Suggested Reviewers */}
              <Card title="Suggested Reviewers">
                <div className="flex flex-wrap gap-2">
                  {summary.suggested_reviewers.map((reviewer, index) => (
                    <Badge key={index} color="purple">
                      {reviewer}
                    </Badge>
                  ))}
                </div>
              </Card>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default PRSummary;
