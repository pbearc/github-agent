import React from "react";
import { Link } from "react-router-dom";

const Footer = () => {
  const currentYear = new Date().getFullYear();

  const navigation = {
    codeExploration: [
      { name: "Code Search", href: "/search" },
      { name: "Code Walkthrough", href: "/navigate/walkthrough" },
      { name: "Function Explainer", href: "/navigate/function" },
      { name: "Architecture Visualizer", href: "/navigate/architecture" },
      { name: "Best Practices Guide", href: "/navigate/practices" },
    ],
    helperTools: [
      { name: "Generate README", href: "/generate/readme" },
      { name: "Generate Dockerfile", href: "/generate/dockerfile" },
      { name: "Add Comments", href: "/generate/comments" },
      { name: "Refactor Code", href: "/generate/refactor" },
      { name: "PR Summary", href: "/pr-summary" },
    ],
    resources: [
      { name: "Repository Info", href: "/repository" },
      { name: "Settings", href: "/settings" },
      {
        name: "GitHub",
        href: "https://github.com/pbearc/github-agent",
        external: true,
      },
      {
        name: "Documentation",
        href: "https://github.com/pbearc/github-agent#readme",
        external: true,
      },
    ],
  };

  return (
    <footer className="bg-dark-100 border-t border-gray-800 relative">
      {/* Gradient bottom border at the top of the footer */}
      <div className="h-1 w-full bg-gradient-to-r from-primary-500 via-orange-500 to-pink-500 absolute top-0"></div>

      <div className="max-w-7xl mx-auto py-12 px-4 sm:px-6 lg:px-8">
        <div className="xl:grid xl:grid-cols-4 xl:gap-8">
          <div className="space-y-8 xl:col-span-1">
            <Link to="/" className="flex items-center">
              <div className="relative mr-2">
                <div className="absolute inset-0 bg-gradient-to-r from-primary-500 to-secondary-500 blur-md opacity-60 rounded-full"></div>
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  className="h-8 w-8 relative text-white"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                  strokeWidth={2}
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4"
                  />
                </svg>
              </div>
              <span className="text-xl font-bold bg-clip-text text-transparent bg-gradient-to-r from-primary-400 via-orange-400 to-pink-400">
                Code Navigator
              </span>
            </Link>
            <p className="text-gray-400 text-sm">
              AI-powered tools to help newcomers quickly understand any GitHub
              repository's structure, patterns, and functionality.
            </p>
            <div className="flex space-x-6">
              <a
                href="https://github.com/pbearc/github-agent"
                target="_blank"
                rel="noopener noreferrer"
                className="text-gray-400 hover:text-gray-300"
              >
                <span className="sr-only">GitHub</span>
                <svg
                  className="h-6 w-6"
                  fill="currentColor"
                  viewBox="0 0 24 24"
                  aria-hidden="true"
                >
                  <path
                    fillRule="evenodd"
                    d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z"
                    clipRule="evenodd"
                  />
                </svg>
              </a>
            </div>
          </div>
          <div className="mt-12 grid grid-cols-3 gap-8 xl:mt-0 xl:col-span-3">
            <div>
              <h3 className="text-sm font-semibold text-pink-400 tracking-wider uppercase">
                Code Exploration
              </h3>
              <ul className="mt-4 space-y-3">
                {navigation.codeExploration.map((item) => (
                  <li key={item.name}>
                    {item.external ? (
                      <a
                        href={item.href}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-gray-400 hover:text-white transition-colors"
                      >
                        {item.name}
                      </a>
                    ) : (
                      <Link
                        to={item.href}
                        className="text-gray-400 hover:text-white transition-colors"
                      >
                        {item.name}
                      </Link>
                    )}
                  </li>
                ))}
              </ul>
            </div>
            <div>
              <h3 className="text-sm font-semibold text-primary-400 tracking-wider uppercase">
                Helper Tools
              </h3>
              <ul className="mt-4 space-y-3">
                {navigation.helperTools.map((item) => (
                  <li key={item.name}>
                    {item.external ? (
                      <a
                        href={item.href}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-gray-400 hover:text-white transition-colors"
                      >
                        {item.name}
                      </a>
                    ) : (
                      <Link
                        to={item.href}
                        className="text-gray-400 hover:text-white transition-colors"
                      >
                        {item.name}
                      </Link>
                    )}
                  </li>
                ))}
              </ul>
            </div>
            <div>
              <h3 className="text-sm font-semibold text-orange-400 tracking-wider uppercase">
                Resources
              </h3>
              <ul className="mt-4 space-y-3">
                {navigation.resources.map((item) => (
                  <li key={item.name}>
                    {item.external ? (
                      <a
                        href={item.href}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-gray-400 hover:text-white transition-colors"
                      >
                        {item.name}
                      </a>
                    ) : (
                      <Link
                        to={item.href}
                        className="text-gray-400 hover:text-white transition-colors"
                      >
                        {item.name}
                      </Link>
                    )}
                  </li>
                ))}
              </ul>
            </div>
          </div>
        </div>
        <div className="mt-12 border-t border-gray-800 pt-8">
          <p className="text-gray-400 text-center text-sm">
            &copy; {currentYear} Code Navigator. All rights reserved.
          </p>
          <p className="text-gray-500 text-center text-xs mt-2">
            Helping developers understand unfamiliar codebases with AI-powered
            tools.
          </p>
        </div>
      </div>
    </footer>
  );
};

export default Footer;
