import React, { useState, useEffect } from "react";
import { Link } from "react-router-dom";
import Card from "../components/ui/Card";
import Button from "../components/ui/Button";
import { motion } from "framer-motion";

const FeatureCard = ({ icon, title, description, link, delay = 0 }) => (
  <motion.div
    initial={{ opacity: 0, y: 20 }}
    whileInView={{ opacity: 1, y: 0 }}
    viewport={{ once: true, margin: "-100px" }}
    transition={{ duration: 0.5, delay }}
    whileHover={{ y: -5, transition: { duration: 0.2 } }}
    className="col-span-1"
  >
    <Card className="h-full">
      <div className="flex flex-col h-full">
        <div className="flex items-center mb-4">
          <div className="p-2 rounded-md bg-primary-50 dark:bg-primary-900/20 text-primary-600 dark:text-primary-400">
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
    <div className="space-y-24">
      {/* Hero Section */}
      <section className="pt-12 md:pt-20 pb-16">
        <div className="text-center max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
          <motion.h1
            initial={{ opacity: 0, y: -10 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6 }}
            className="text-4xl sm:text-5xl md:text-6xl font-extrabold tracking-tight text-gray-900 dark:text-white"
          >
            <span className="block">Enhance your</span>
            <span className="block bg-clip-text text-transparent bg-gradient-to-r from-primary-600 to-secondary-600">
              GitHub repositories
            </span>
            <span className="block">with AI</span>
          </motion.h1>

          <motion.p
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ duration: 0.6, delay: 0.3 }}
            className="mt-6 max-w-2xl mx-auto text-xl text-gray-600 dark:text-gray-300"
          >
            GitHub Agent uses AI to automate documentation, analyze code, and
            enhance your development workflow.
          </motion.p>

          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6, delay: 0.5 }}
            className="mt-10 max-w-sm mx-auto sm:max-w-none sm:flex sm:justify-center gap-4"
          >
            <Link to="/repository">
              <Button variant="primary" size="lg">
                Get Started
              </Button>
            </Link>
            <a
              href="https://github.com/pbearc/github-agent"
              target="_blank"
              rel="noopener noreferrer"
            >
              <Button variant="outline" size="lg">
                View on GitHub
              </Button>
            </a>
          </motion.div>
        </div>
      </section>

      {/* Product overview section */}
      <section className="py-12">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <motion.div
            initial={{ opacity: 0 }}
            whileInView={{ opacity: 1 }}
            viewport={{ once: true }}
            transition={{ duration: 0.5 }}
            className="lg:grid lg:grid-cols-2 lg:gap-16 items-center"
          >
            <div>
              <h2 className="text-3xl font-bold text-gray-900 dark:text-white">
                What is GitHub Agent?
              </h2>
              <p className="mt-4 text-lg text-gray-600 dark:text-gray-300">
                GitHub Agent is an AI-powered tool that helps developers improve
                their GitHub repositories with automated code analysis,
                documentation, and enhancement.
              </p>
              <div className="mt-8 space-y-4">
                <div className="flex items-start">
                  <div className="flex-shrink-0">
                    <div className="flex items-center justify-center h-8 w-8 rounded-md bg-primary-500 text-white">
                      <svg
                        className="h-5 w-5"
                        xmlns="http://www.w3.org/2000/svg"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke="currentColor"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth="2"
                          d="M5 13l4 4L19 7"
                        />
                      </svg>
                    </div>
                  </div>
                  <div className="ml-4">
                    <h3 className="text-lg font-medium text-gray-900 dark:text-white">
                      Automatic Documentation
                    </h3>
                    <p className="mt-2 text-base text-gray-600 dark:text-gray-300">
                      Generate professional READMEs and Dockerfiles with a
                      single click, customized to your repository.
                    </p>
                  </div>
                </div>
                <div className="flex items-start">
                  <div className="flex-shrink-0">
                    <div className="flex items-center justify-center h-8 w-8 rounded-md bg-primary-500 text-white">
                      <svg
                        className="h-5 w-5"
                        xmlns="http://www.w3.org/2000/svg"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke="currentColor"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth="2"
                          d="M5 13l4 4L19 7"
                        />
                      </svg>
                    </div>
                  </div>
                  <div className="ml-4">
                    <h3 className="text-lg font-medium text-gray-900 dark:text-white">
                      Code Improvements
                    </h3>
                    <p className="mt-2 text-base text-gray-600 dark:text-gray-300">
                      Add comments, refactor code, and find bugs with
                      intelligent AI analysis.
                    </p>
                  </div>
                </div>
                <div className="flex items-start">
                  <div className="flex-shrink-0">
                    <div className="flex items-center justify-center h-8 w-8 rounded-md bg-primary-500 text-white">
                      <svg
                        className="h-5 w-5"
                        xmlns="http://www.w3.org/2000/svg"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke="currentColor"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth="2"
                          d="M5 13l4 4L19 7"
                        />
                      </svg>
                    </div>
                  </div>
                  <div className="ml-4">
                    <h3 className="text-lg font-medium text-gray-900 dark:text-white">
                      Streamlined Workflow
                    </h3>
                    <p className="mt-2 text-base text-gray-600 dark:text-gray-300">
                      Push changes directly to your repository with no
                      additional setup.
                    </p>
                  </div>
                </div>
              </div>
            </div>
            <div className="mt-10 lg:mt-0 flex justify-center items-center">
              <motion.div
                initial={{ opacity: 0, scale: 0.9 }}
                whileInView={{ opacity: 1, scale: 1 }}
                viewport={{ once: true }}
                transition={{ duration: 0.5, delay: 0.2 }}
                className="relative rounded-xl bg-gradient-to-r from-primary-50 to-secondary-50 dark:from-primary-900/20 dark:to-secondary-900/20 shadow-xl overflow-hidden p-8"
              >
                <div className="w-full h-full max-w-md rounded-md overflow-hidden shadow-2xl">
                  <img
                    className="w-full rounded-t-md"
                    src="/api/placeholder/600/300"
                    alt="GitHub Agent Screenshot"
                  />
                  <div className="bg-white dark:bg-dark-100 p-5 rounded-b-md">
                    <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
                      AI-Powered Repository Enhancement
                    </h3>
                    <p className="mt-2 text-gray-600 dark:text-gray-300">
                      Transform your repositories with intelligent documentation
                      and code analysis.
                    </p>
                  </div>
                </div>

                {/* Decorative elements */}
                <div className="absolute top-0 right-0 -mt-4 -mr-4 w-24 h-24 bg-primary-500 opacity-20 rounded-full"></div>
                <div className="absolute bottom-0 left-0 -mb-4 -ml-4 w-16 h-16 bg-secondary-500 opacity-20 rounded-full"></div>
              </motion.div>
            </div>
          </motion.div>
        </div>
      </section>

      {/* How it works section */}
      <section className="py-12 bg-gray-50 dark:bg-gray-800">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center">
            <motion.h2
              initial={{ opacity: 0 }}
              whileInView={{ opacity: 1 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5 }}
              className="text-3xl font-bold text-gray-900 dark:text-white"
            >
              How It Works
            </motion.h2>
            <motion.p
              initial={{ opacity: 0 }}
              whileInView={{ opacity: 1 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5, delay: 0.2 }}
              className="mt-4 max-w-2xl mx-auto text-xl text-gray-600 dark:text-gray-300"
            >
              Simple, powerful, and efficient workflow
            </motion.p>
          </div>

          <div className="mt-16">
            <div className="lg:grid lg:grid-cols-3 lg:gap-8">
              <motion.div
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ duration: 0.5 }}
                className="text-center"
              >
                <div className="flex items-center justify-center h-16 w-16 rounded-full bg-primary-100 dark:bg-primary-900/30 text-primary-600 dark:text-primary-400 mx-auto">
                  <svg
                    className="h-8 w-8"
                    xmlns="http://www.w3.org/2000/svg"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth="2"
                      d="M13 10V3L4 14h7v7l9-11h-7z"
                    />
                  </svg>
                </div>
                <h3 className="mt-6 text-xl font-medium text-gray-900 dark:text-white">
                  1. Connect
                </h3>
                <p className="mt-2 text-base text-gray-600 dark:text-gray-300">
                  Paste your GitHub repository URL and connect instantly with no
                  setup required.
                </p>
              </motion.div>

              <motion.div
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ duration: 0.5, delay: 0.2 }}
                className="mt-10 lg:mt-0 text-center"
              >
                <div className="flex items-center justify-center h-16 w-16 rounded-full bg-primary-100 dark:bg-primary-900/30 text-primary-600 dark:text-primary-400 mx-auto">
                  <svg
                    className="h-8 w-8"
                    xmlns="http://www.w3.org/2000/svg"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth="2"
                      d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z"
                    />
                  </svg>
                </div>
                <h3 className="mt-6 text-xl font-medium text-gray-900 dark:text-white">
                  2. Enhance
                </h3>
                <p className="mt-2 text-base text-gray-600 dark:text-gray-300">
                  Choose from our suite of AI-powered tools to enhance your
                  repository's code and documentation.
                </p>
              </motion.div>

              <motion.div
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ duration: 0.5, delay: 0.4 }}
                className="mt-10 lg:mt-0 text-center"
              >
                <div className="flex items-center justify-center h-16 w-16 rounded-full bg-primary-100 dark:bg-primary-900/30 text-primary-600 dark:text-primary-400 mx-auto">
                  <svg
                    className="h-8 w-8"
                    xmlns="http://www.w3.org/2000/svg"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth="2"
                      d="M5 13l4 4L19 7"
                    />
                  </svg>
                </div>
                <h3 className="mt-6 text-xl font-medium text-gray-900 dark:text-white">
                  3. Deploy
                </h3>
                <p className="mt-2 text-base text-gray-600 dark:text-gray-300">
                  Review the AI-generated content and push changes directly to
                  your repository.
                </p>
              </motion.div>
            </div>
          </div>
        </div>
      </section>

      {/* Features grid section */}
      <section className="py-12">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center">
            <motion.h2
              initial={{ opacity: 0 }}
              whileInView={{ opacity: 1 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5 }}
              className="text-3xl font-bold text-gray-900 dark:text-white"
            >
              Features
            </motion.h2>
            <motion.p
              initial={{ opacity: 0 }}
              whileInView={{ opacity: 1 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5, delay: 0.2 }}
              className="mt-4 max-w-2xl mx-auto text-xl text-gray-600 dark:text-gray-300"
            >
              Everything you need to optimize your GitHub repositories
            </motion.p>
          </div>

          <div className="mt-16 grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
            {features.map((feature, index) => (
              <FeatureCard
                key={feature.title}
                icon={feature.icon}
                title={feature.title}
                description={feature.description}
                link={feature.link}
                delay={0.1 * index}
              />
            ))}
          </div>
        </div>
      </section>

      {/* Recently used repositories section */}
      {recentlyUsed.length > 0 && (
        <section className="py-12">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <motion.h2
              initial={{ opacity: 0 }}
              whileInView={{ opacity: 1 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5 }}
              className="text-2xl font-bold text-gray-900 dark:text-white mb-8"
            >
              Recently Used Repositories
            </motion.h2>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              {recentlyUsed.map((repo, index) => (
                <motion.div
                  key={repo.url}
                  initial={{ opacity: 0, y: 20 }}
                  whileInView={{ opacity: 1, y: 0 }}
                  viewport={{ once: true }}
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
          </div>
        </section>
      )}

      {/* CTA section */}
      <section className="py-12">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <motion.div
            initial={{ opacity: 0, scale: 0.95 }}
            whileInView={{ opacity: 1, scale: 1 }}
            viewport={{ once: true }}
            transition={{ duration: 0.5 }}
            className="rounded-2xl bg-gradient-to-r from-primary-600 to-secondary-700 shadow-xl overflow-hidden"
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
                      AI-assisted development. No account required - just paste
                      your repository URL and get started.
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
      </section>
    </div>
  );
};

export default Dashboard;
