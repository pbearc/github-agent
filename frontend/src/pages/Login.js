import React, { useState, useEffect } from "react";
import { useNavigate, Link } from "react-router-dom";
import { motion } from "framer-motion";
import { toast } from "sonner";
import Card from "../components/ui/Card";
import Button from "../components/ui/Button";
import Input from "../components/ui/Input";
import { useAuth } from "../context/AuthContext";

const Login = () => {
  const navigate = useNavigate();
  const { isAuthenticated, login } = useAuth();

  const [token, setToken] = useState("");
  const [username, setUsername] = useState("");
  const [loading, setLoading] = useState(false);
  const [errors, setErrors] = useState({});

  useEffect(() => {
    // Redirect if already authenticated
    if (isAuthenticated) {
      navigate("/");
    }
  }, [isAuthenticated, navigate]);

  const handleLogin = async (e) => {
    e.preventDefault();

    if (!token) {
      setErrors({ token: "Please enter a GitHub token" });
      return;
    }

    if (!username) {
      setErrors({ username: "Please enter your GitHub username" });
      return;
    }

    setLoading(true);
    setErrors({});

    try {
      // In a real app, you might validate the token with GitHub API here
      login(
        token,
        username,
        "https://github.githubassets.com/images/modules/logos_page/GitHub-Mark.png"
      );
      toast.success("Successfully signed in");
      navigate("/");
    } catch (error) {
      console.error("Login error:", error);
      toast.error("Failed to sign in");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-dark-200 px-4 py-12 transition-colors duration-300">
      <motion.div
        initial={{ opacity: 0, scale: 0.95 }}
        animate={{ opacity: 1, scale: 1 }}
        transition={{ duration: 0.3 }}
        className="w-full max-w-md"
      >
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
            GitHub Agent
          </h1>
          <p className="mt-2 text-gray-600 dark:text-gray-400">
            Sign in to access all features
          </p>
        </div>

        <Card>
          <form onSubmit={handleLogin} className="space-y-6">
            <Input
              label="GitHub Username"
              placeholder="Your GitHub username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              error={errors.username}
              required
            />

            <Input
              label="GitHub Token"
              type="password"
              placeholder="Your GitHub personal access token"
              value={token}
              onChange={(e) => setToken(e.target.value)}
              error={errors.token}
              required
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

            <Button
              type="submit"
              fullWidth
              isLoading={loading}
              disabled={loading}
            >
              Sign in
            </Button>
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
            </ol>
          </div>
        </Card>

        <div className="mt-4 text-center">
          <Link
            to="/"
            className="text-sm text-primary-600 dark:text-primary-500 hover:underline"
          >
            Back to home
          </Link>
        </div>
      </motion.div>
    </div>
  );
};

export default Login;
