import React, { useState, useEffect } from "react";
import { Link } from "react-router-dom";
import Card from "../components/ui/Card";
import Button from "../components/ui/Button";
import { motion } from "framer-motion";

const FeatureCard = ({ icon, title, description, link, delay }) => (
  <motion.div
    initial={{ opacity: 0, y: 20 }}
    animate={{ opacity: 1, y: 0 }}
    transition={{ duration: 0.5, delay }}
    whileHover={{ y: -5, transition: { duration: 0.2 } }}
    className="col-span-1"
  >
    <Card className="h-full">
      <div className="flex flex-col h-full">
        <div className="flex items-center mb-4">
          <div className="p-2 rounded-md bg-primary-100 dark:bg-primary-900 text-primary-600 dark:text-primary-300">
            {icon}
          </div>
        </div>
        <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">
          {title}
        </h3>
        <p className="text-gray-600 dark:text-gray-300 mb-4 flex-grow">
          {description}
        </p>
        <div className="mt-auto">
          <Link to={link}>
            <Button variant="outline" fullWidth>
              Try it now
            </Button>
          </Link>
        </div>
      </div>
    </Card>
  </motion.div>
);

const Dashboard = () => {
  const [recentlyUsed, setRecentlyUsed] = useState([]);

  useEffect(() => {
    // Fetch recently used repositories from localStorage
    const recentRepos = JSON.parse(
      localStorage.getItem("recentRepositories") || "[]"
    );
    setRecentlyUsed(recentRepos.slice(0, 3)); // Take only the last 3
  }, []);

  const features = [
    {
      icon: (
        <svg
          xmlns="http://www.w3.org/2000/svg"
          className="h-6 w-6"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
          />
        </svg>
      ),
      title: "Generate README",
      description:
        "Create professional README files for your repositories with a single click.",
      link: "/generate/readme",
    },
    {
      icon: (
        <svg
          xmlns="http://www.w3.org/2000/svg"
          className="h-6 w-6"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"
          />
        </svg>
      ),
      title: "Generate Dockerfile",
      description:
        "Automatically create optimized Dockerfiles based on your project structure.",
      link: "/generate/dockerfile",
    },
    {
      icon: (
        <svg
          xmlns="http://www.w3.org/2000/svg"
          className="h-6 w-6"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M7 8h10M7 12h4m1 8l-4-4H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-3l-4 4z"
          />
        </svg>
      ),
      title: "Add Comments",
      description:
        "Enhance code readability by adding comprehensive comments to your files.",
      link: "/generate/comments",
    },
    {
      icon: (
        <svg
          xmlns="http://www.w3.org/2000/svg"
          className="h-6 w-6"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
          />
        </svg>
      ),
      title: "Refactor Code",
      description:
        "Clean up and optimize your code with intelligent refactoring suggestions.",
      link: "/generate/refactor",
    },
    {
      icon: (
        <svg
          xmlns="http://www.w3.org/2000/svg"
          className="h-6 w-6"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
          />
        </svg>
      ),
      title: "Code Search",
      description:
        "Quickly find and explore code in your repositories with powerful search capabilities.",
      link: "/search",
    },
    {
      icon: (
        <svg
          xmlns="http://www.w3.org/2000/svg"
          className="h-6 w-6"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M15 15l-2 5L9 9l11 4-5 2zm0 0l5 5M7.188 2.239l.777 2.897M5.136 7.965l-2.898-.777M13.95 4.05l-2.122 2.122m-5.657 5.656l-2.12 2.122"
          />
        </svg>
      ),
      title: "Repository Info",
      description:
        "Get detailed information about any GitHub repository with a single click.",
      link: "/repository",
    },
  ];

  return (
    <div>
      {/* Hero Section */}
      <div className="relative overflow-hidden mb-16">
        <div className="absolute inset-0 bg-gradient-to-r from-primary-600 to-secondary-600 opacity-10 dark:opacity-20"></div>
        <div className="relative px-4 py-16 sm:px-6 lg:px-8 lg:py-24">
          <motion.div
            initial={{ opacity: 0, y: -20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.8 }}
            className="text-center"
          >
            <motion.h1
              initial={{ opacity: 0, scale: 0.9 }}
              animate={{ opacity: 1, scale: 1 }}
              transition={{ duration: 0.5, delay: 0.2 }}
              className="text-4xl font-extrabold tracking-tight text-gray-900 dark:text-white sm:text-5xl md:text-6xl"
            >
              <span className="block">Supercharge your</span>
              <span className="block text-primary-600 dark:text-primary-500">
                GitHub Repositories
              </span>
            </motion.h1>

            <motion.p
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              transition={{ duration: 0.5, delay: 0.4 }}
              className="mt-4 max-w-2xl mx-auto text-xl text-gray-500 dark:text-gray-300"
            >
              AI-powered tools to generate, analyze, and improve your code. From
              README creation to code refactoring, all in one place.
            </motion.p>

            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5, delay: 0.6 }}
              className="mt-10 max-w-sm mx-auto sm:max-w-none sm:flex sm:justify-center"
            >
              <div className="space-y-4 sm:space-y-0 sm:mx-auto sm:inline-grid sm:grid-cols-2 sm:gap-5">
                <Link to="/repository">
                  <Button variant="primary" size="lg" fullWidth>
                    Get Started
                  </Button>
                </Link>
                <a
                  href="https://github.com/pbearc/github-agent"
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  <Button variant="outline" size="lg" fullWidth>
                    View on GitHub
                  </Button>
                </a>
              </div>
            </motion.div>
          </motion.div>
        </div>

        {/* Animated blob shapes */}
        <div className="absolute top-0 left-0 -translate-x-1/2 transform">
          <motion.div
            animate={{
              scale: [1, 1.2, 1],
              x: [0, 20, 0],
              y: [0, 30, 0],
            }}
            transition={{
              duration: 8,
              repeat: Infinity,
              repeatType: "reverse",
            }}
            className="w-72 h-72 bg-primary-500 opacity-10 rounded-full"
          />
        </div>
        <div className="absolute bottom-0 right-0 translate-x-1/3 translate-y-1/3 transform">
          <motion.div
            animate={{
              scale: [1, 1.3, 1],
              x: [0, -20, 0],
              y: [0, -30, 0],
            }}
            transition={{
              duration: 10,
              repeat: Infinity,
              repeatType: "reverse",
            }}
            className="w-72 h-72 bg-secondary-500 opacity-10 rounded-full"
          />
        </div>
      </div>

      {/* Features section */}
      <div className="mb-16">
        <div className="text-center mb-12">
          <motion.h2
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5 }}
            className="text-3xl font-bold text-gray-900 dark:text-white"
          >
            Powerful Features
          </motion.h2>
          <motion.p
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ duration: 0.5, delay: 0.2 }}
            className="mt-4 max-w-2xl mx-auto text-xl text-gray-500 dark:text-gray-400"
          >
            Everything you need to boost your GitHub repositories
          </motion.p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
          {features.map((feature, index) => (
            <FeatureCard
              key={feature.title}
              icon={feature.icon}
              title={feature.title}
              description={feature.description}
              link={feature.link}
              delay={0.1 + index * 0.1}
            />
          ))}
        </div>
      </div>

      {/* Recently used repositories section */}
      {recentlyUsed.length > 0 && (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ duration: 0.5, delay: 0.5 }}
          className="mb-16"
        >
          <h2 className="text-2xl font-bold text-gray-900 dark:text-white mb-6">
            Recently Used Repositories
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            {recentlyUsed.map((repo, index) => (
              <motion.div
                key={repo.url}
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.3, delay: 0.1 * index }}
              >
                <Card>
                  <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-1">
                    {repo.name}
                  </h3>
                  <p className="text-sm text-gray-500 dark:text-gray-400 mb-4">
                    {repo.owner}
                  </p>
                  <div className="flex space-x-2 mt-2">
                    <Link to="/repository" state={{ url: repo.url }}>
                      <Button size="sm" variant="primary">
                        View Info
                      </Button>
                    </Link>
                    <a
                      href={repo.url}
                      target="_blank"
                      rel="noopener noreferrer"
                    >
                      <Button size="sm" variant="outline">
                        Open in GitHub
                      </Button>
                    </a>
                  </div>
                </Card>
              </motion.div>
            ))}
          </div>
        </motion.div>
      )}

      {/* Testimonials / How it works */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5, delay: 0.3 }}
        className="mb-16 bg-gray-50 dark:bg-gray-800 rounded-lg p-8"
      >
        <div className="text-center mb-8">
          <h2 className="text-3xl font-bold text-gray-900 dark:text-white">
            How It Works
          </h2>
          <p className="mt-4 max-w-2xl mx-auto text-xl text-gray-500 dark:text-gray-400">
            Simple yet powerful workflow to enhance your GitHub repositories
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.4 }}
            className="text-center"
          >
            <div className="mx-auto flex items-center justify-center h-12 w-12 rounded-md bg-primary-500 text-white">
              <svg
                xmlns="http://www.w3.org/2000/svg"
                className="h-6 w-6"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M13 10V3L4 14h7v7l9-11h-7z"
                />
              </svg>
            </div>
            <h3 className="mt-6 text-xl font-medium text-gray-900 dark:text-white">
              1. Connect
            </h3>
            <p className="mt-2 text-base text-gray-500 dark:text-gray-400">
              Paste your GitHub repository URL and connect instantly
            </p>
          </motion.div>

          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.5 }}
            className="text-center"
          >
            <div className="mx-auto flex items-center justify-center h-12 w-12 rounded-md bg-primary-500 text-white">
              <svg
                xmlns="http://www.w3.org/2000/svg"
                className="h-6 w-6"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M19.428 15.428a2 2 0 00-1.022-.547l-2.387-.477a6 6 0 00-3.86.517l-.318.158a6 6 0 01-3.86.517L6.05 15.21a2 2 0 00-1.806.547M8 4h8l-1 1v5.172a2 2 0 00.586 1.414l5 5c1.26 1.26.367 3.414-1.415 3.414H4.828c-1.782 0-2.674-2.154-1.414-3.414l5-5A2 2 0 009 10.172V5L8 4z"
                />
              </svg>
            </div>
            <h3 className="mt-6 text-xl font-medium text-gray-900 dark:text-white">
              2. Select Tool
            </h3>
            <p className="mt-2 text-base text-gray-500 dark:text-gray-400">
              Choose from our suite of AI-powered tools to enhance your
              repository
            </p>
          </motion.div>

          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.6 }}
            className="text-center"
          >
            <div className="mx-auto flex items-center justify-center h-12 w-12 rounded-md bg-primary-500 text-white">
              <svg
                xmlns="http://www.w3.org/2000/svg"
                className="h-6 w-6"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M5 13l4 4L19 7"
                />
              </svg>
            </div>
            <h3 className="mt-6 text-xl font-medium text-gray-900 dark:text-white">
              3. Generate & Push
            </h3>
            <p className="mt-2 text-base text-gray-500 dark:text-gray-400">
              Review the AI-generated content and push directly to your
              repository
            </p>
          </motion.div>
        </div>
      </motion.div>

      {/* CTA section */}
      <motion.div
        initial={{ opacity: 0, scale: 0.95 }}
        animate={{ opacity: 1, scale: 1 }}
        transition={{ duration: 0.5, delay: 0.7 }}
        className="bg-primary-600 dark:bg-primary-800 rounded-lg shadow-xl overflow-hidden"
      >
        <div className="px-6 py-12 sm:px-12 lg:px-16 lg:py-16">
          <div className="max-w-7xl mx-auto">
            <div className="lg:grid lg:grid-cols-2 lg:gap-8 lg:items-center">
              <div>
                <h2 className="text-3xl font-extrabold text-white">
                  Ready to enhance your GitHub repositories?
                </h2>
                <p className="mt-4 text-lg text-primary-100 dark:text-primary-200">
                  Start using GitHub Agent today and experience the power of
                  AI-assisted development. No account required - just paste your
                  repository URL and get started.
                </p>
                <div className="mt-8">
                  <div className="inline-flex rounded-md shadow">
                    <Link to="/repository">
                      <Button variant="white" size="lg">
                        Get Started Now
                      </Button>
                    </Link>
                  </div>
                </div>
              </div>
              <div className="mt-10 lg:mt-0 lg:col-start-2">
                <div className="flex justify-center lg:justify-end">
                  <img
                    className="w-full max-w-md rounded-lg shadow-lg"
                    src="/api/placeholder/600/400"
                    alt="GitHub Agent"
                  />
                </div>
              </div>
            </div>
          </div>
        </div>
      </motion.div>
    </div>
  );
};

export default Dashboard;
