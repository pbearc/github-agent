import React, { useState, useEffect } from "react";
import { useLocation } from "react-router-dom";
import Card from "../components/ui/Card";
import Button from "../components/ui/Button";
import Input from "../components/ui/Input";
import CodeEditor from "../components/ui/CodeEditor";
import { toast } from "sonner";
import { motion } from "framer-motion";
import { repositoryService, navigatorService } from "../services/api";
import ReactMarkdown from "react-markdown";
import rehypeSanitize from "rehype-sanitize";
import Tabs from "../components/ui/Tabs";

// Helper components for different API types
const CommitsView = ({ data }) => (
  <div className="space-y-4">
    {data?.map((commit, index) => (
      <div
        key={index}
        className="border border-gray-800 rounded p-4 bg-dark-200/50"
      >
        <div className="flex justify-between">
          <h3 className="text-sm font-medium text-gray-200">
            {commit.message?.split("\n")[0]}
          </h3>
          <span className="text-xs text-gray-400">
            {new Date(commit.commit_date).toLocaleDateString()}
          </span>
        </div>
        <p className="text-xs text-gray-400 mt-2">Author: {commit.author}</p>
        {commit.files_changed?.length > 0 && (
          <div className="mt-2">
            <p className="text-xs text-gray-400">
              Files changed: {commit.files_changed.length}
            </p>
            <div className="mt-1 text-xs">
              <span className="inline-block px-2 py-1 rounded-full text-xs bg-primary-900/30 text-primary-300 mr-2">
                +{commit.lines_added}
              </span>
              <span className="inline-block px-2 py-1 rounded-full text-xs bg-red-900/30 text-red-300">
                -{commit.lines_deleted}
              </span>
            </div>
          </div>
        )}
      </div>
    ))}
  </div>
);

const PullRequestsView = ({ data }) => (
  <div className="space-y-4">
    {data?.map((pr, index) => (
      <div
        key={index}
        className="border border-gray-800 rounded p-4 bg-dark-200/50"
      >
        <div className="flex justify-between">
          <h3 className="text-sm font-medium text-gray-200">
            #{pr.number} {pr.title}
          </h3>
          <span
            className={`text-xs px-2 py-1 rounded-full ${
              pr.state === "open"
                ? "bg-green-900/30 text-green-300"
                : pr.state === "merged"
                ? "bg-purple-900/30 text-purple-300"
                : "bg-red-900/30 text-red-300"
            }`}
          >
            {pr.state}
          </span>
        </div>
        <p className="text-xs text-gray-400 mt-2">
          By: {pr.author} • {new Date(pr.created_at).toLocaleDateString()}
        </p>
        {pr.labels?.length > 0 && (
          <div className="flex flex-wrap gap-1 mt-2">
            {pr.labels.map((label, i) => (
              <span
                key={i}
                className="inline-block px-2 py-1 rounded-full text-xs bg-secondary-900/30 text-secondary-300"
              >
                {label}
              </span>
            ))}
          </div>
        )}
      </div>
    ))}
  </div>
);

const IssuesView = ({ data }) => (
  <div className="space-y-4">
    {data?.map((issue, index) => (
      <div
        key={index}
        className="border border-gray-800 rounded p-4 bg-dark-200/50"
      >
        <div className="flex justify-between">
          <h3 className="text-sm font-medium text-gray-200">
            #{issue.number} {issue.title}
          </h3>
          <span
            className={`text-xs px-2 py-1 rounded-full ${
              issue.state === "open"
                ? "bg-green-900/30 text-green-300"
                : "bg-red-900/30 text-red-300"
            }`}
          >
            {issue.state}
          </span>
        </div>
        <p className="text-xs text-gray-400 mt-2">
          By: {issue.author} • {new Date(issue.created_at).toLocaleDateString()}
        </p>
        {issue.labels?.length > 0 && (
          <div className="flex flex-wrap gap-1 mt-2">
            {issue.labels.map((label, i) => (
              <span
                key={i}
                className="inline-block px-2 py-1 rounded-full text-xs bg-secondary-900/30 text-secondary-300"
              >
                {label}
              </span>
            ))}
          </div>
        )}
      </div>
    ))}
  </div>
);

const ReleasesView = ({ data }) => (
  <div className="space-y-4">
    {data?.map((release, index) => (
      <div
        key={index}
        className="border border-gray-800 rounded p-4 bg-dark-200/50"
      >
        <div className="flex justify-between">
          <h3 className="text-sm font-medium text-gray-200">
            {release.name || release.tag_name}
          </h3>
          <span
            className={`text-xs px-2 py-1 rounded-full ${
              release.pre_release
                ? "bg-orange-900/30 text-orange-300"
                : "bg-blue-900/30 text-blue-300"
            }`}
          >
            {release.pre_release ? "Pre-release" : "Release"}
          </span>
        </div>
        <p className="text-xs text-gray-400 mt-2">
          Released by: {release.author} •{" "}
          {new Date(
            release.published_at || release.created_at
          ).toLocaleDateString()}
        </p>
        {release.assets_count > 0 && (
          <p className="text-xs text-gray-400 mt-1">
            Assets: {release.assets_count}
          </p>
        )}
      </div>
    ))}
  </div>
);

const StatsView = ({ data }) => {
  // Sample visualization for stats
  return (
    <div className="space-y-4">
      <div className="border border-gray-800 rounded p-4 bg-dark-200/50">
        <h3 className="text-sm font-medium text-gray-200 mb-3">
          Top Contributors
        </h3>
        <div className="space-y-2">
          {data?.contributors?.slice(0, 5).map((contributor, index) => (
            <div key={index} className="flex justify-between items-center">
              <span className="text-sm text-gray-300">
                {contributor.author}
              </span>
              <div className="flex items-center">
                <span className="text-xs text-gray-400 mr-2">
                  {contributor.commits} commits
                </span>
                <div className="w-24 bg-gray-700 rounded-full h-2">
                  <div
                    className="bg-gradient-to-r from-primary-500 to-secondary-500 h-2 rounded-full"
                    style={{
                      width: `${Math.min(
                        100,
                        (contributor.commits /
                          Math.max(
                            ...data.contributors.map((c) => c.commits)
                          )) *
                          100
                      )}%`,
                    }}
                  ></div>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>

      <div className="border border-gray-800 rounded p-4 bg-dark-200/50">
        <h3 className="text-sm font-medium text-gray-200 mb-3">
          Activity by Week
        </h3>
        <div className="flex h-32 items-end space-x-1">
          {Object.entries(data?.participation_stats || {})
            .slice(0, 12)
            .map(([week, count], index) => (
              <div key={index} className="flex-1 flex flex-col items-center">
                <div
                  className="w-full bg-gradient-to-t from-primary-500 to-secondary-500 rounded-t"
                  style={{ height: `${Math.max(4, (count / 100) * 100)}%` }}
                ></div>
                <span className="text-xs text-gray-500 mt-1">
                  {week.replace("week_", "")}
                </span>
              </div>
            ))}
        </div>
      </div>
    </div>
  );
};

const UsersView = ({ data }) => (
  <div className="space-y-4">
    {data?.map((user, index) => (
      <div
        key={index}
        className="border border-gray-800 rounded p-4 bg-dark-200/50 flex items-center"
      >
        {user.avatar_url && (
          <img
            src={user.avatar_url}
            alt={user.username}
            className="w-10 h-10 rounded-full mr-4"
          />
        )}
        <div>
          <h3 className="text-sm font-medium text-gray-200">{user.username}</h3>
          <p className="text-xs text-gray-400 mt-1">
            Contributions: {user.contributions}
          </p>
        </div>
        {user.is_bot && (
          <span className="ml-auto inline-block px-2 py-1 rounded-full text-xs bg-gray-800 text-gray-300">
            BOT
          </span>
        )}
      </div>
    ))}
  </div>
);

const RepoInfoView = ({ data }) => {
  const info = data?.info || {};
  const languages = data?.languages || {};

  // Calculate percentages for language bar
  const totalBytes = Object.values(languages).reduce(
    (sum, bytes) => sum + bytes,
    0
  );
  const languageItems = Object.entries(languages)
    .sort((a, b) => b[1] - a[1])
    .map(([language, bytes]) => ({
      name: language,
      percentage: (bytes / totalBytes) * 100,
      bytes,
    }));

  return (
    <div className="space-y-4">
      <div className="border border-gray-800 rounded p-4 bg-dark-200/50">
        <h3 className="text-sm font-medium text-gray-200 mb-2">
          Repository Info
        </h3>
        <div className="space-y-2">
          <p className="text-xs text-gray-400">
            Owner: <span className="text-gray-200">{info.owner}</span>
          </p>
          <p className="text-xs text-gray-400">
            Name: <span className="text-gray-200">{info.name}</span>
          </p>
          <p className="text-xs text-gray-400">
            Default Branch:{" "}
            <span className="text-gray-200">{info.default_branch}</span>
          </p>
          <p className="text-xs text-gray-400">
            Stars: <span className="text-gray-200">{info.stars}</span>
          </p>
          <p className="text-xs text-gray-400">
            Forks: <span className="text-gray-200">{info.forks}</span>
          </p>
          {info.description && (
            <p className="text-xs text-gray-400 mt-2">
              Description:{" "}
              <span className="text-gray-200">{info.description}</span>
            </p>
          )}
        </div>
      </div>

      {languageItems.length > 0 && (
        <div className="border border-gray-800 rounded p-4 bg-dark-200/50">
          <h3 className="text-sm font-medium text-gray-200 mb-2">Languages</h3>
          <div className="w-full h-4 bg-gray-800 rounded-full overflow-hidden flex">
            {languageItems.map((lang, index) => (
              <div
                key={index}
                className="h-full"
                style={{
                  width: `${lang.percentage}%`,
                  backgroundColor: getLanguageColor(lang.name),
                }}
                title={`${lang.name}: ${lang.percentage.toFixed(1)}%`}
              ></div>
            ))}
          </div>
          <div className="mt-2 flex flex-wrap gap-2">
            {languageItems.slice(0, 5).map((lang, index) => (
              <div key={index} className="flex items-center text-xs">
                <div
                  className="w-3 h-3 rounded-full mr-1"
                  style={{ backgroundColor: getLanguageColor(lang.name) }}
                ></div>
                <span className="text-gray-300">{lang.name}</span>
                <span className="text-gray-500 ml-1">
                  ({lang.percentage.toFixed(1)}%)
                </span>
              </div>
            ))}
            {languageItems.length > 5 && (
              <div className="text-xs text-gray-500">
                +{languageItems.length - 5} more
              </div>
            )}
          </div>
        </div>
      )}

      {data?.topics?.length > 0 && (
        <div className="border border-gray-800 rounded p-4 bg-dark-200/50">
          <h3 className="text-sm font-medium text-gray-200 mb-2">Topics</h3>
          <div className="flex flex-wrap gap-2">
            {data.topics.map((topic, index) => (
              <span
                key={index}
                className="inline-block px-2 py-1 rounded-full text-xs bg-blue-900/30 text-blue-300"
              >
                {topic}
              </span>
            ))}
          </div>
        </div>
      )}
    </div>
  );
};

// A helper function to get colors for language visualization
const getLanguageColor = (language) => {
  const colors = {
    JavaScript: "#f1e05a",
    TypeScript: "#2b7489",
    Python: "#3572A5",
    Java: "#b07219",
    Go: "#00ADD8",
    C: "#555555",
    "C++": "#f34b7d",
    Ruby: "#701516",
    PHP: "#4F5D95",
    CSS: "#563d7c",
    HTML: "#e34c26",
    Swift: "#ffac45",
    Kotlin: "#F18E33",
    Rust: "#dea584",
    Scala: "#c22d40",
    // Default color for other languages
    default: "#8257e6",
  };

  return colors[language] || colors.default;
};

const CodeSearch = () => {
  const location = useLocation();
  const [url, setUrl] = useState(location.state?.url || "");
  const [branch, setBranch] = useState("");
  const [question, setQuestion] = useState("");
  const [loading, setLoading] = useState(false);
  const [indexing, setIndexing] = useState(false);
  const [searchResults, setSearchResults] = useState(null);
  const [repoInfo, setRepoInfo] = useState(null);
  const [selectedResult, setSelectedResult] = useState(null);
  const [errors, setErrors] = useState({});
  const [isIndexed, setIsIndexed] = useState(false);
  const [showSuggestions, setShowSuggestions] = useState(false);
  const [activeTab, setActiveTab] = useState(0);

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

  const handleIndex = async () => {
    if (!url) {
      setErrors({
        url: !url ? "Please enter a GitHub repository URL" : "",
      });
      return;
    }

    setIndexing(true);
    setErrors({});

    try {
      const branchToUse =
        branch || (repoInfo ? repoInfo.default_branch : "main");

      // Use navigatorService to index the repository
      const response = await navigatorService.index(url, branchToUse);

      toast.success("Repository indexed successfully");
      setIsIndexed(true);

      // Fetch repo info if not already loaded
      if (!repoInfo) {
        await fetchRepositoryInfo(url);
      }
    } catch (error) {
      console.error("Error indexing repository:", error);

      let errorMessage = "Failed to index repository";
      if (error.response?.data?.error) {
        errorMessage = error.response.data.error;
      }

      toast.error(errorMessage);
    } finally {
      setIndexing(false);
    }
  };

  const handleSearch = async (e) => {
    if (e) e.preventDefault();

    if (!url || !question) {
      setErrors({
        url: !url ? "Please enter a GitHub repository URL" : "",
        question: !question ? "Please enter a question" : "",
      });
      return;
    }

    setLoading(true);
    setErrors({});
    setSearchResults(null);
    setSelectedResult(null);

    try {
      const branchToUse =
        branch || (repoInfo ? repoInfo.default_branch : "main");

      // Show indexing toast
      toast.info("Analyzing repository...");

      // Use smartNavigate instead of the old question method
      const response = await navigatorService.smartNavigate(
        url,
        question,
        branchToUse
      );

      setSearchResults(response.data);
      setShowSuggestions(false);

      // Reset the active tab
      setActiveTab(0);

      // For code search responses, handle file selection
      if (
        response.data.api_type === "code_search" &&
        response.data.relevant_files &&
        response.data.relevant_files.length > 0
      ) {
        // Select the first result by default
        setSelectedResult(response.data.relevant_files[0]);
        toast.success(
          `Found ${response.data.relevant_files.length} relevant ${
            response.data.relevant_files.length === 1 ? "file" : "files"
          }`
        );
      } else {
        // For other API types, show a generic success message
        toast.success("Analysis complete!");
      }

      // Fetch repo info if not already loaded
      if (!repoInfo) {
        await fetchRepositoryInfo(url);
      }
    } catch (error) {
      console.error("Error analyzing repository:", error);

      let errorMessage = "Failed to analyze repository";
      if (error.response?.data?.error) {
        errorMessage = error.response.data.error;
      }

      toast.error(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  const detectLanguage = (filePath) => {
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

  // Sample questions
  const sampleQuestions = [
    "How is authentication implemented?",
    "What's the main entry point of the application?",
    "How does data flow through the codebase?",
    "Explain the directory structure of this project",
    "What design patterns are used in this codebase?",
    "Who are the top contributors to this repository?",
    "When was the last release and what did it include?",
    "What open issues need to be fixed?",
    "Show me recent pull requests",
    "What programming languages are used in this project?",
  ];

  // Determine which tabs to show based on API type
  const getTabsForResults = () => {
    if (!searchResults) return [];

    const tabs = [{ label: "Answer", content: "answer" }];

    // Add tabs based on API type
    switch (searchResults.api_type) {
      case "code_search":
        if (searchResults.relevant_files?.length > 0) {
          tabs.push({ label: "Code", content: "code" });
        }
        break;
      case "commits":
        if (searchResults.extra_data?.commits?.length > 0) {
          tabs.push({ label: "Commits", content: "commits" });
        }
        break;
      case "pulls":
        if (searchResults.extra_data?.pull_requests?.length > 0) {
          tabs.push({ label: "Pull Requests", content: "pulls" });
        }
        break;
      case "issues":
        if (searchResults.extra_data?.issues?.length > 0) {
          tabs.push({ label: "Issues", content: "issues" });
        }
        break;
      case "releases":
        if (searchResults.extra_data?.releases?.length > 0) {
          tabs.push({ label: "Releases", content: "releases" });
        }
        break;
      case "stats":
        if (searchResults.extra_data?.stats) {
          tabs.push({ label: "Statistics", content: "stats" });
        }
        break;
      case "users":
        if (searchResults.extra_data?.contributors?.length > 0) {
          tabs.push({ label: "Contributors", content: "users" });
        }
        break;
      case "repos":
        if (searchResults.extra_data?.repository) {
          tabs.push({ label: "Repository Info", content: "repo" });
        }
        break;
      default:
        break;
    }

    return tabs;
  };

  // Render content based on active tab
  const renderTabContent = () => {
    if (!searchResults) return null;

    const tabs = getTabsForResults();
    if (activeTab >= tabs.length) return null;

    const content = tabs[activeTab].content;

    switch (content) {
      case "answer":
        return (
          <div className="prose dark:prose-invert max-w-none text-gray-200">
            <ReactMarkdown
              rehypePlugins={[rehypeSanitize]}
              components={{
                code({ node, inline, className, children, ...props }) {
                  const match = /language-(\w+)/.exec(className || "");
                  return !inline && match ? (
                    <CodeEditor
                      code={String(children).replace(/\n$/, "")}
                      language={match[1]}
                      readOnly={true}
                      lineNumbers={true}
                      showCopyButton={true}
                    />
                  ) : (
                    <code className={className} {...props}>
                      {children}
                    </code>
                  );
                },
              }}
            >
              {searchResults.answer}
            </ReactMarkdown>
          </div>
        );
      case "code":
        return (
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            {/* Files list on the left */}
            <div>
              <Card
                title={`Relevant Files (${
                  searchResults.relevant_files?.length || 0
                })`}
                noPadding
                className="h-full bg-gradient-to-b from-dark-100/70 to-dark-100 border-none shadow-glow"
              >
                <div className="h-full flex flex-col">
                  <div className="p-4 border-b border-gray-800">
                    <p className="text-sm text-gray-400">
                      Question:{" "}
                      <span className="font-medium text-white">{question}</span>
                    </p>
                  </div>

                  <div
                    className="flex-1 overflow-y-auto scrollbar-thin scrollbar-thumb-gray-700 scrollbar-track-transparent"
                    style={{ maxHeight: "500px" }}
                  >
                    {searchResults.relevant_files &&
                    searchResults.relevant_files.length > 0 ? (
                      <ul className="divide-y divide-gray-800">
                        {searchResults.relevant_files.map((file, index) => (
                          <li
                            key={index}
                            className={`
                              p-4 hover:bg-dark-200 cursor-pointer transition-colors duration-150
                              ${
                                selectedResult &&
                                selectedResult.path === file.path
                                  ? "bg-secondary-900/20 border-l-2 border-secondary-500"
                                  : ""
                              }
                            `}
                            onClick={() => setSelectedResult(file)}
                          >
                            <h3 className="text-sm font-medium text-gray-200 truncate">
                              {file.path.split("/").pop()}
                            </h3>
                            <p className="mt-1 text-xs text-gray-500 truncate">
                              {file.path}
                            </p>
                            <div className="mt-1 text-xs">
                              <span
                                className={`inline-block px-2 py-1 rounded-full text-xs
                                  ${
                                    file.relevance > 80
                                      ? "bg-accent-900/30 text-accent-300"
                                      : file.relevance > 50
                                      ? "bg-secondary-900/30 text-secondary-300"
                                      : "bg-primary-900/30 text-primary-300"
                                  }
                                `}
                              >
                                Relevance: {file.relevance.toFixed(1)}%
                              </span>
                            </div>
                          </li>
                        ))}
                      </ul>
                    ) : (
                      <div className="p-6 text-center text-gray-500">
                        No relevant files found for your question.
                      </div>
                    )}
                  </div>
                </div>
              </Card>
            </div>

            {/* Code view on the right */}
            <div className="lg:col-span-2">
              {selectedResult ? (
                <Card
                  title={selectedResult.path.split("/").pop()}
                  subtitle={selectedResult.path}
                  className="bg-gradient-to-b from-dark-100/70 to-dark-100 border-none shadow-glow"
                >
                  <CodeEditor
                    code={selectedResult.snippet}
                    language={detectLanguage(selectedResult.path)}
                    readOnly={true}
                    lineNumbers={true}
                    showCopyButton={true}
                    style={{ minHeight: "500px" }}
                  />
                </Card>
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
                        d="M8 16l2.879-2.879m0 0a3 3 0 104.243-4.242 3 3 0 00-4.243 4.242zM21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                      />
                    </svg>
                    <p className="text-lg font-medium">
                      Select a file from the list to view the code
                    </p>
                    <p className="mt-2 text-sm text-gray-600">
                      The AI has identified these files as most relevant to your
                      question
                    </p>
                  </div>
                </Card>
              )}
            </div>
          </div>
        );
      case "commits":
        return <CommitsView data={searchResults.extra_data?.commits} />;
      case "pulls":
        return (
          <PullRequestsView data={searchResults.extra_data?.pull_requests} />
        );
      case "issues":
        return <IssuesView data={searchResults.extra_data?.issues} />;
      case "releases":
        return <ReleasesView data={searchResults.extra_data?.releases} />;
      case "stats":
        return <StatsView data={searchResults.extra_data?.stats} />;
      case "users":
        return <UsersView data={searchResults.extra_data?.contributors} />;
      case "repo":
        return <RepoInfoView data={searchResults.extra_data?.repository} />;
      default:
        return null;
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
              Code Q&A
            </h1>
            <p className="text-gray-600 dark:text-gray-300">
              Ask questions about any GitHub repository and get AI-powered
              answers
            </p>
          </div>
          <div className="mt-4 md:mt-0">
            <div className="text-sm inline-flex items-center rounded-full bg-warning-50 dark:bg-warning-900/20 px-3 py-1 text-warning-700 dark:text-warning-300 border border-warning-200 dark:border-warning-800">
              <svg
                xmlns="http://www.w3.org/2000/svg"
                className="h-4 w-4 mr-1"
                viewBox="0 0 20 20"
                fill="currentColor"
              >
                <path
                  fillRule="evenodd"
                  d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z"
                  clipRule="evenodd"
                />
              </svg>
              <span>Ask questions about code, commits, issues, and more!</span>
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

          <div className="mt-6">
            <form onSubmit={handleSearch}>
              <div className="relative">
                <Input
                  label="Ask a Question About the Repository"
                  placeholder="How is authentication implemented?"
                  value={question}
                  onChange={(e) => {
                    setQuestion(e.target.value);
                    // Show suggestions when user starts typing
                    if (e.target.value && !showSuggestions) {
                      setShowSuggestions(true);
                    } else if (!e.target.value) {
                      setShowSuggestions(false);
                    }
                  }}
                  error={errors.question}
                  onFocus={() => question && setShowSuggestions(true)}
                  onBlur={() =>
                    setTimeout(() => setShowSuggestions(false), 200)
                  }
                  required
                  dark
                />

                {showSuggestions && (
                  <div className="absolute z-10 w-full mt-1 bg-dark-100 border border-gray-700 rounded-md shadow-lg overflow-hidden">
                    <div className="max-h-60 overflow-y-auto">
                      {sampleQuestions
                        .filter(
                          (q) =>
                            q.toLowerCase().includes(question.toLowerCase()) ||
                            question === ""
                        )
                        .map((q, idx) => (
                          <div
                            key={idx}
                            className="px-4 py-2 hover:bg-dark-200 cursor-pointer text-gray-200"
                            onClick={() => {
                              setQuestion(q);
                              setShowSuggestions(false);
                            }}
                          >
                            {q}
                          </div>
                        ))}
                    </div>
                  </div>
                )}
              </div>

              <div className="mt-6 flex flex-col sm:flex-row gap-4">
                <Button
                  type="submit"
                  isLoading={loading}
                  disabled={loading}
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
                      d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-8-3a1 1 0 00-.867.5 1 1 0 11-1.731-1A3 3 0 0113 8a3.001 3.001 0 01-2 2.83V11a1 1 0 11-2 0v-1a1 1 0 011-1 1 1 0 100-2zm0 8a1 1 0 100-2 1 1 0 000 2z"
                      clipRule="evenodd"
                    />
                  </svg>
                  Ask Question
                </Button>

                <Button
                  type="button"
                  variant="outline"
                  onClick={() => setShowSuggestions(!showSuggestions)}
                  className="border-accent-500 text-accent-400 hover:bg-accent-500/10"
                >
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    className="h-5 w-5 mr-2"
                    viewBox="0 0 20 20"
                    fill="currentColor"
                  >
                    <path
                      fillRule="evenodd"
                      d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-8-3a1 1 0 00-.867.5 1 1 0 11-1.731-1A3 3 0 0113 8a3.001 3.001 0 01-2 2.83V11a1 1 0 11-2 0v-1a1 1 0 011-1 1 1 0 100-2zm0 8a1 1 0 100-2 1 1 0 000 2z"
                      clipRule="evenodd"
                    />
                  </svg>
                  Suggested Questions
                </Button>
              </div>
            </form>
          </div>
        </Card>
      </div>

      {searchResults && (
        <div className="space-y-6">
          {/* Results section with tabs */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5 }}
          >
            <Card
              title="AI Analysis"
              subtitle={
                searchResults.api_type
                  ? `Query type: ${searchResults.api_type.replace("_", " ")}`
                  : ""
              }
              className="bg-gradient-to-b from-dark-100/70 to-dark-100 border-none shadow-glow"
            >
              {/* Tabs for different content types */}
              <Tabs
                tabs={getTabsForResults().map((tab) => tab.label)}
                activeTab={activeTab}
                onChange={setActiveTab}
                className="mb-6"
              />

              {/* Tab content */}
              {renderTabContent()}

              {/* Follow-up questions */}
              {searchResults.followup_questions &&
                searchResults.followup_questions.length > 0 && (
                  <div className="mt-8 border-t border-gray-800 pt-6">
                    <h3 className="text-sm font-medium text-gray-200 mb-3">
                      Follow-up Questions:
                    </h3>
                    <div className="flex flex-wrap gap-2">
                      {searchResults.followup_questions.map((q, idx) => (
                        <button
                          key={idx}
                          className="px-3 py-2 text-sm bg-dark-200 hover:bg-dark-300 text-gray-300 rounded-md transition"
                          onClick={() => {
                            setQuestion(q);
                            handleSearch();
                          }}
                        >
                          {q}
                        </button>
                      ))}
                    </div>
                  </div>
                )}
            </Card>
          </motion.div>
        </div>
      )}

      {!searchResults && !loading && (
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
              Ask Me Anything About the Repository
            </h2>
            <p className="text-gray-400 mb-6 max-w-md mx-auto">
              Enter a GitHub URL, then ask questions to understand the
              repository - from code structure to contributors and everything in
              between.
            </p>
            <div className="flex flex-col gap-4 mt-6">
              <p className="text-sm font-medium text-gray-400">
                Example Questions:
              </p>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                {sampleQuestions.slice(0, 4).map((q, idx) => (
                  <div
                    key={idx}
                    className="text-left p-3 bg-dark-200 rounded-lg cursor-pointer hover:bg-dark-300 border border-gray-800 text-gray-300"
                    onClick={() => {
                      setQuestion(q);
                      // Auto-scroll to question input
                      document
                        .querySelector(
                          'input[placeholder="How is authentication implemented?"]'
                        )
                        ?.scrollIntoView({
                          behavior: "smooth",
                          block: "center",
                        });
                    }}
                  >
                    {q}
                  </div>
                ))}
              </div>
            </div>
          </div>
        </motion.div>
      )}
    </div>
  );
};

export default CodeSearch;
