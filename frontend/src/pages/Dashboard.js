import React, { useState, useEffect } from "react";
import { Link } from "react-router-dom";
import Card from "../components/ui/Card";
import Button from "../components/ui/Button";
import { motion } from "framer-motion";

const FeatureCard = ({
  icon,
  title,
  description,
  link,
  colorClass = "from-primary-600 to-pink-600",
  delay = 0,
}) => (
  <motion.div
    initial={{ opacity: 0, y: 20 }}
    whileInView={{ opacity: 1, y: 0 }}
    viewport={{ once: true, margin: "-100px" }}
    transition={{ duration: 0.5, delay }}
    whileHover={{ y: -5, transition: { duration: 0.2 } }}
    className="col-span-1"
  >
    <div className="h-full bg-dark-100 rounded-xl p-6 border border-gray-800 hover:border-gray-700 shadow-xl hover:shadow-2xl transition-all duration-300">
      <div className="flex flex-col h-full">
        <div className="flex items-center mb-4">
          <div
            className={`p-3 rounded-md bg-gradient-to-br ${colorClass} text-white`}
          >
            {icon}
          </div>
        </div>
        <h3 className="text-lg font-semibold text-white mb-2">{title}</h3>
        <p className="text-gray-400 mb-4 flex-grow">{description}</p>
        <div className="mt-auto">
          <Link to={link}>
            <Button
              variant="outline"
              fullWidth
              className={`border-gray-700 text-gray-300 hover:bg-gradient-to-r ${colorClass} hover:text-white hover:border-transparent`}
            >
              Explore Now
            </Button>
          </Link>
        </div>
      </div>
    </div>
  </motion.div>
);

const StepCard = ({
  number,
  icon,
  title,
  description,
  colorClass,
  delay = 0,
}) => (
  <motion.div
    initial={{ opacity: 0, y: 20 }}
    whileInView={{ opacity: 1, y: 0 }}
    viewport={{ once: true }}
    transition={{ duration: 0.5, delay }}
    className="bg-dark-100 p-8 rounded-xl border border-gray-800 relative"
  >
    <div
      className={`absolute -top-5 -left-5 flex items-center justify-center h-10 w-10 rounded-full bg-gradient-to-r ${colorClass} text-white font-bold text-lg`}
    >
      {number}
    </div>
    <div
      className={`flex items-center justify-center h-16 w-16 rounded-full bg-gradient-to-br ${colorClass} bg-opacity-20 text-white mx-auto mb-4`}
    >
      {icon}
    </div>
    <h3 className="text-xl font-medium text-white text-center mb-3">{title}</h3>
    <p className="text-gray-400 text-center">{description}</p>
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
            d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
          />
        </svg>
      ),
      title: "Code Search",
      description:
        "Search through repositories with natural language questions to find and understand specific code patterns.",
      link: "/search",
      colorClass: "from-primary-600 to-secondary-600",
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
            d="M9 20l-5.447-2.724A1 1 0 013 16.382V5.618a1 1 0 011.447-.894L9 7m0 13l6-3m-6 3V7m6 10l4.553 2.276A1 1 0 0021 18.382V7.618a1 1 0 00-.553-.894L15 4m0 13V4m0 0L9 7"
          />
        </svg>
      ),
      title: "Code Walkthrough",
      description:
        "Get guided tours through unfamiliar codebases to quickly understand structure and key components.",
      link: "/navigate/walkthrough",
      colorClass: "from-secondary-600 to-pink-600",
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
            d="M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
          />
        </svg>
      ),
      title: "Function Explainer",
      description:
        "Understand specific functions or components with detailed explanations of parameters, return values, and usage patterns.",
      link: "/navigate/function",
      colorClass: "from-orange-500 to-pink-600",
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
            d="M4 5a1 1 0 011-1h14a1 1 0 011 1v2a1 1 0 01-1 1H5a1 1 0 01-1-1V5zM4 13a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H5a1 1 0 01-1-1v-6zM16 13a1 1 0 011-1h2a1 1 0 011 1v6a1 1 0 01-1 1h-2a1 1 0 01-1-1v-6z"
          />
        </svg>
      ),
      title: "Architecture Visualizer",
      description:
        "Visualize the architecture of any repository with interactive diagrams showing dependencies and data flows.",
      link: "/navigate/architecture",
      colorClass: "from-lime-500 to-primary-600",
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
            d="M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
          />
        </svg>
      ),
      title: "Codebase Q&A",
      description:
        "Ask natural language questions about any codebase and get AI-powered answers with relevant code snippets.",
      link: "/search",
      colorClass: "from-pink-600 to-orange-500",
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
            d="M12 6V4m0 2a2 2 0 100 4m0-4a2 2 0 110 4m-6 8a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4m6 6v10m6-2a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4"
          />
        </svg>
      ),
      title: "Best Practices Guide",
      description:
        "Get custom style guides and best practices based on the patterns used in any codebase to follow team conventions.",
      link: "/navigate/practices",
      colorClass: "from-secondary-600 to-primary-600",
    },
  ];

  // Helper tools features
  const helperTools = [
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
      colorClass: "from-orange-500 to-lime-500",
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
      colorClass: "from-pink-600 to-primary-600",
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
      colorClass: "from-secondary-600 to-lime-500",
    },
  ];

  const steps = [
    {
      number: 1,
      icon: (
        <svg
          xmlns="http://www.w3.org/2000/svg"
          className="h-8 w-8"
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
      ),
      title: "Paste Repository URL",
      description:
        "Simply paste the GitHub repository URL you want to explore - no cloning, no setup.",
      colorClass: "from-primary-600 to-secondary-600",
    },
    {
      number: 2,
      icon: (
        <svg
          xmlns="http://www.w3.org/2000/svg"
          className="h-8 w-8"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z"
          />
        </svg>
      ),
      title: "Choose Your Tool",
      description:
        "Select from our arsenal of AI-powered tools designed to help you understand code quickly.",
      colorClass: "from-orange-500 to-pink-600",
    },
    {
      number: 3,
      icon: (
        <svg
          xmlns="http://www.w3.org/2000/svg"
          className="h-8 w-8"
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
      ),
      title: "Gain Instant Understanding",
      description:
        "Get detailed explanations, diagrams, and walkthroughs that accelerate your understanding.",
      colorClass: "from-lime-500 to-primary-600",
    },
  ];

  return (
    <div className="space-y-20">
      {/* Hero Section */}
      <section className="py-12 md:py-20">
        <div className="relative z-10 overflow-hidden rounded-3xl bg-gradient-to-br from-dark-100 to-dark-200 border border-gray-800 shadow-2xl">
          <div className="absolute inset-0 bg-[radial-gradient(circle_at_top_right,_rgba(209,57,255,0.15),_transparent_40%)]"></div>
          <div className="absolute inset-0 bg-[radial-gradient(circle_at_bottom_left,_rgba(12,188,255,0.15),_transparent_40%)]"></div>
          <div className="absolute inset-0 bg-[radial-gradient(circle_at_center,_rgba(255,123,0,0.1),_transparent_30%)]"></div>

          {/* Floating decorative elements */}
          <div className="absolute top-20 right-20 w-16 h-16 rounded-full bg-gradient-to-r from-pink-500 to-secondary-500 opacity-20 blur-xl animate-float"></div>
          <div
            className="absolute bottom-20 left-20 w-20 h-20 rounded-full bg-gradient-to-r from-primary-500 to-lime-500 opacity-20 blur-xl animate-float"
            style={{ animationDelay: "2s" }}
          ></div>
          <div
            className="absolute top-1/2 left-1/3 w-12 h-12 rounded-full bg-gradient-to-r from-orange-500 to-pink-500 opacity-20 blur-xl animate-float"
            style={{ animationDelay: "1s" }}
          ></div>

          <div className="relative px-6 py-20 sm:px-12 lg:px-16 text-center">
            <motion.h1
              initial={{ opacity: 0, y: -10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.8 }}
              className="mx-auto max-w-4xl text-4xl sm:text-5xl md:text-6xl font-bold tracking-tight"
            >
              <span className="block text-white">Understand Any</span>
              <span className="block mt-1 bg-clip-text text-transparent bg-gradient-to-r from-primary-400 via-orange-400 to-pink-400">
                GitHub Codebase
              </span>
              <span className="block mt-1 text-white">In Minutes</span>
            </motion.h1>

            <motion.p
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              transition={{ duration: 0.8, delay: 0.3 }}
              className="mt-6 mx-auto max-w-2xl text-xl text-gray-400"
            >
              Navigate unfamiliar code with confidence. Our AI-powered tools
              help newcomers quickly understand any repository's structure,
              patterns, and functionality.
            </motion.p>

            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.8, delay: 0.5 }}
              className="mt-10 flex flex-col sm:flex-row gap-4 justify-center"
            >
              <Link to="/search">
                <Button
                  variant="primary"
                  size="lg"
                  className="w-full sm:w-auto px-8 bg-gradient-to-r from-primary-600 via-secondary-600 to-pink-600 hover:from-primary-500 hover:via-secondary-500 hover:to-pink-500 border-none shadow-xl"
                >
                  Try Code Search
                </Button>
              </Link>
              <Link to="/navigate/walkthrough">
                <Button
                  variant="outline"
                  size="lg"
                  className="w-full sm:w-auto px-8 border-gray-700 text-gray-300 hover:bg-gradient-to-r hover:from-orange-600 hover:to-secondary-600 hover:text-white hover:border-transparent"
                >
                  Get a Walkthrough
                </Button>
              </Link>
            </motion.div>
          </div>
        </div>
      </section>

      {/* Main features section */}
      <section>
        <div className="text-center mb-12">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.6 }}
          >
            <h2 className="text-3xl font-bold text-white mb-4">
              Powerful Tools for Code Understanding
            </h2>
            <p className="max-w-2xl mx-auto text-gray-400">
              Explore unfamiliar codebases with confidence using our suite of
              AI-powered navigation tools
            </p>
          </motion.div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {features.map((feature, index) => (
            <FeatureCard
              key={feature.title}
              icon={feature.icon}
              title={feature.title}
              description={feature.description}
              link={feature.link}
              colorClass={feature.colorClass}
              delay={0.1 * index}
            />
          ))}
        </div>
      </section>

      {/* How it works section */}
      <section className="py-16 relative">
        <div className="absolute inset-0 bg-gradient-to-b from-transparent via-dark-100/80 to-transparent"></div>
        <div className="relative">
          <div className="text-center mb-12">
            <motion.div
              initial={{ opacity: 0 }}
              whileInView={{ opacity: 1 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5 }}
            >
              <h2 className="text-3xl font-bold text-white mb-4">
                How It Works
              </h2>
              <p className="max-w-2xl mx-auto text-gray-400">
                Understand any codebase in three simple steps
              </p>
            </motion.div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
            {steps.map((step, index) => (
              <StepCard
                key={index}
                number={step.number}
                icon={step.icon}
                title={step.title}
                description={step.description}
                colorClass={step.colorClass}
                delay={0.1 * index}
              />
            ))}
          </div>
        </div>
      </section>

      {/* Helper tools section */}
      <section>
        <div className="text-center mb-12">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.6 }}
          >
            <h2 className="text-3xl font-bold text-white mb-4">Helper Tools</h2>
            <p className="max-w-2xl mx-auto text-gray-400">
              Additional utilities to help enhance your GitHub repositories
            </p>
          </motion.div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          {helperTools.map((tool, index) => (
            <FeatureCard
              key={tool.title}
              icon={tool.icon}
              title={tool.title}
              description={tool.description}
              link={tool.link}
              colorClass={tool.colorClass}
              delay={0.1 * index}
            />
          ))}
        </div>
      </section>

      {/* Recently explored repositories section */}
      {recentlyUsed.length > 0 && (
        <section>
          <div className="text-center mb-8">
            <motion.h2
              initial={{ opacity: 0 }}
              whileInView={{ opacity: 1 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5 }}
              className="text-2xl font-bold text-white"
            >
              Recently Explored Repositories
            </motion.h2>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            {recentlyUsed.map((repo, index) => (
              <motion.div
                key={repo.url}
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ duration: 0.3, delay: 0.1 * index }}
              >
                <div className="bg-dark-100 rounded-xl p-6 border border-gray-800 hover:border-gray-700 shadow-xl transition-all duration-300">
                  <h3 className="text-lg font-medium text-white mb-1">
                    {repo.name}
                  </h3>
                  <p className="text-sm text-gray-500 mb-4">{repo.owner}</p>
                  <div className="flex flex-wrap gap-2">
                    <Link to="/search" state={{ url: repo.url }}>
                      <Button
                        size="sm"
                        variant="primary"
                        className="bg-gradient-to-r from-primary-600 to-secondary-600 border-none"
                      >
                        Search Code
                      </Button>
                    </Link>
                    <Link to="/navigate/walkthrough" state={{ url: repo.url }}>
                      <Button
                        size="sm"
                        variant="outline"
                        className="border-gray-700 text-gray-300 hover:bg-gradient-to-r hover:from-orange-600 hover:to-pink-600 hover:text-white hover:border-transparent"
                      >
                        Get Walkthrough
                      </Button>
                    </Link>
                  </div>
                </div>
              </motion.div>
            ))}
          </div>
        </section>
      )}

      {/* CTA section */}
      <section className="py-10 my-10">
        <motion.div
          initial={{ opacity: 0, scale: 0.95 }}
          whileInView={{ opacity: 1, scale: 1 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5 }}
          className="rounded-2xl shadow-2xl overflow-hidden border border-gray-800 relative"
        >
          <div className="absolute inset-0 bg-gradient-to-br from-dark-100/80 to-dark-200/90"></div>
          <div className="absolute inset-0 bg-[radial-gradient(circle_at_top_right,_rgba(209,57,255,0.2),_transparent_40%)]"></div>
          <div className="absolute inset-0 bg-[radial-gradient(circle_at_bottom_left,_rgba(255,123,0,0.2),_transparent_40%)]"></div>

          {/* Moving gradient background */}
          <div className="absolute inset-0 opacity-10">
            <div className="absolute inset-0 bg-gradient-conic from-primary-500 via-secondary-500 to-pink-500 animate-pulse-slow"></div>
          </div>

          <div className="relative px-6 py-16 sm:px-12 lg:px-16">
            <div className="max-w-7xl mx-auto">
              <div className="lg:grid lg:grid-cols-2 lg:gap-8 lg:items-center">
                <div>
                  <h2 className="text-3xl font-extrabold text-white">
                    Ready to understand any GitHub repository?
                  </h2>
                  <p className="mt-4 text-lg text-gray-400">
                    Stop being intimidated by unfamiliar code. Start exploring
                    with confidence using our AI-powered navigation tools.
                  </p>
                  <div className="mt-8">
                    <Link to="/search">
                      <Button
                        variant="primary"
                        size="lg"
                        className="bg-gradient-to-r from-orange-500 via-pink-600 to-secondary-600 hover:from-orange-400 hover:via-pink-500 hover:to-secondary-500 border-none shadow-xl"
                      >
                        Try Code Search Now
                      </Button>
                    </Link>
                  </div>
                </div>
                <div className="mt-10 lg:mt-0 lg:col-start-2">
                  <div className="flex justify-center lg:justify-end">
                    <div className="relative">
                      <div className="absolute -inset-1 bg-gradient-to-r from-primary-500 to-secondary-500 rounded-lg blur-sm opacity-40"></div>
                      <img
                        className="relative w-full max-w-md rounded-lg shadow-lg"
                        src="/api/placeholder/600/400"
                        alt="Code Understanding"
                      />
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </motion.div>
      </section>
    </div>
  );
};

export default Dashboard;
