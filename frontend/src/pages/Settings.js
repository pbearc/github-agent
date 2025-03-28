import React, { useState, useEffect } from "react";
import Card from "../components/ui/Card";
import Button from "../components/ui/Button";
import Input from "../components/ui/Input";
import { toast } from "sonner";
import { motion } from "framer-motion";
import { useAuth } from "../context/AuthContext";
import { useTheme } from "../context/ThemeContext";

const Settings = () => {
  const { isAuthenticated, user, githubToken, login, logout } = useAuth();
  const { darkMode, toggleDarkMode } = useTheme();

  const [token, setToken] = useState("");
  const [username, setUsername] = useState("");
  const [saving, setSaving] = useState(false);
  const [errors, setErrors] = useState({});

  useEffect(() => {
    if (isAuthenticated) {
      setToken(githubToken || "");
      setUsername(user?.username || "");
    }
  }, [isAuthenticated, githubToken, user]);

  const handleSaveToken = async (e) => {
    e.preventDefault();

    if (!token) {
      setErrors({ token: "Please enter a GitHub token" });
      return;
    }

    if (!username) {
      setErrors({ username: "Please enter your GitHub username" });
      return;
    }

    setSaving(true);
    setErrors({});

    try {
      // In a real app, you might validate the token with GitHub API here
      login(
        token,
        username,
        "https://github.githubassets.com/images/modules/logos_page/GitHub-Mark.png"
      );
      toast.success("GitHub token saved successfully");
    } catch (error) {
      console.error("Error saving token:", error);
      toast.error("Failed to save GitHub token");
    } finally {
      setSaving(false);
    }
  };

  const handleClearHistory = () => {
    localStorage.removeItem("recentRepositories");
    toast.success("Search history cleared successfully");
  };

  return (
    <div>
      <motion.div
        initial={{ opacity: 0, y: -20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.3 }}
      >
        <h1 className="text-2xl font-bold text-gray-900 dark:text-white mb-2">
          Settings
        </h1>
        <p className="text-gray-600 dark:text-gray-300 mb-6">
          Manage your GitHub Agent preferences and account settings
        </p>
      </motion.div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div>
          <Card title="GitHub Authentication">
            <form onSubmit={handleSaveToken}>
              <Input
                label="GitHub Username"
                placeholder="Your GitHub username"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                error={errors.username}
                required
                className="mb-4"
              />

              <Input
                label="GitHub Token"
                type="password"
                placeholder="Your GitHub personal access token"
                value={token}
                onChange={(e) => setToken(e.target.value)}
                error={errors.token}
                required
                className="mb-2"
                helpText={
                  <span>
                    Token needs permissions:{" "}
                    <code className="text-xs bg-gray-100 dark:bg-gray-800 p-1 rounded">
                      repo
                    </code>
                    ,{" "}
                    <code className="text-xs bg-gray-100 dark:bg-gray-800 p-1 rounded">
                      read:user
                    </code>
                  </span>
                }
              />

              <div className="mt-6 space-y-4">
                <Button
                  type="submit"
                  fullWidth
                  isLoading={saving}
                  disabled={saving}
                >
                  Save Token
                </Button>

                {isAuthenticated && (
                  <Button variant="danger" fullWidth onClick={logout}>
                    Sign Out
                  </Button>
                )}
              </div>
            </form>

            <div className="mt-6">
              <p className="text-sm text-gray-600 dark:text-gray-400 mb-2">
                <strong>How to get a GitHub token:</strong>
              </p>
              <ol className="text-sm text-gray-600 dark:text-gray-400 list-decimal list-inside space-y-1">
                <li>
                  Go to{" "}
                  <a
                    href="https://github.com/settings/tokens"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-primary-600 dark:text-primary-500 hover:underline"
                  >
                    GitHub &gt; Settings &gt; Developer settings &gt; Personal
                    access tokens
                  </a>
                </li>
                <li>Click "Generate new token"</li>
                <li>
                  Give your token a name and select the{" "}
                  <code className="bg-gray-100 dark:bg-gray-800 p-1 rounded">
                    repo
                  </code>{" "}
                  and{" "}
                  <code className="bg-gray-100 dark:bg-gray-800 p-1 rounded">
                    read:user
                  </code>{" "}
                  scopes
                </li>
                <li>Click "Generate token" and copy the token</li>
                <li>Paste the token in the field above</li>
              </ol>
            </div>
          </Card>
        </div>

        <div className="space-y-6">
          <Card title="Appearance">
            <div className="flex items-center justify-between py-2">
              <span className="text-gray-700 dark:text-gray-300">
                Dark Mode
              </span>
              <button
                onClick={toggleDarkMode}
                className={`
                  relative inline-flex items-center h-6 rounded-full w-11
                  ${
                    darkMode ? "bg-primary-600" : "bg-gray-300 dark:bg-gray-600"
                  }
                  transition-colors duration-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500
                `}
              >
                <span className="sr-only">Toggle dark mode</span>
                <span
                  className={`
                    inline-block w-5 h-5 rounded-full bg-white 
                    ${darkMode ? "translate-x-6" : "translate-x-1"}
                    transition-transform duration-200
                  `}
                />
              </button>
            </div>
          </Card>

          <Card title="Data Management">
            <div className="space-y-4">
              <p className="text-gray-600 dark:text-gray-400 text-sm">
                Manage your local data and search history used by GitHub Agent.
              </p>

              <div className="pt-2">
                <Button variant="secondary" onClick={handleClearHistory}>
                  Clear Repository History
                </Button>
              </div>
            </div>
          </Card>

          <Card title="About GitHub Agent">
            <div className="space-y-4">
              <p className="text-gray-600 dark:text-gray-400 text-sm">
                GitHub Agent is an AI-powered tool that helps you manage and
                improve your GitHub repositories.
              </p>

              <div className="flex items-center justify-between pt-2">
                <span className="text-gray-600 dark:text-gray-400 text-sm">
                  Version
                </span>
                <span className="text-gray-800 dark:text-gray-200 text-sm font-medium">
                  1.0.0
                </span>
              </div>

              <div className="flex items-center justify-between">
                <span className="text-gray-600 dark:text-gray-400 text-sm">
                  Created by
                </span>
                <a
                  href="https://github.com/pbearc"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-primary-600 dark:text-primary-500 text-sm font-medium hover:underline"
                >
                  pbearc
                </a>
              </div>
            </div>
          </Card>
        </div>
      </div>
    </div>
  );
};

export default Settings;
