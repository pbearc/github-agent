import React from "react";
import { Link } from "react-router-dom";

const Footer = () => {
  const currentYear = new Date().getFullYear();

  const navigation = {
    features: [
      { name: "Generate README", href: "/generate/readme" },
      { name: "Generate Dockerfile", href: "/generate/dockerfile" },
      { name: "Add Comments", href: "/generate/comments" },
      { name: "Refactor Code", href: "/generate/refactor" },
    ],
    resources: [
      { name: "Repository Info", href: "/repository" },
      { name: "Code Search", href: "/search" },
      { name: "Settings", href: "/settings" },
    ],
    company: [
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
    <footer className="bg-white dark:bg-dark-100 border-t border-gray-200 dark:border-gray-700">
      <div className="max-w-7xl mx-auto py-12 px-4 sm:px-6 lg:px-8">
        <div className="xl:grid xl:grid-cols-4 xl:gap-8">
          <div className="space-y-8 xl:col-span-1">
            <Link to="/" className="flex items-center">
              <svg
                xmlns="http://www.w3.org/2000/svg"
                className="h-8 w-8 mr-2 text-primary-600"
                viewBox="0 0 20 20"
                fill="currentColor"
              >
                <path
                  fillRule="evenodd"
                  d="M10 1c4.42 0 8 3.58 8 8 0 3.54-2.29 6.53-5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0018 9c0-4.42-3.58-8-8-8z"
                  clipRule="evenodd"
                />
              </svg>
              <span className="text-xl font-semibold text-primary-600 dark:text-primary-500">
                GitHub Agent
              </span>
            </Link>
            <p className="text-gray-500 dark:text-gray-400 text-sm">
              AI-powered tools to enhance your GitHub repositories with
              documentation, analysis, and code improvements.
            </p>
          </div>
          <div className="mt-12 grid grid-cols-3 gap-8 xl:mt-0 xl:col-span-3">
            <div>
              <h3 className="text-sm font-semibold text-gray-500 dark:text-gray-400 tracking-wider uppercase">
                Features
              </h3>
              <ul className="mt-4 space-y-4">
                {navigation.features.map((item) => (
                  <li key={item.name}>
                    {item.external ? (
                      <a
                        href={item.href}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-base text-gray-600 dark:text-gray-300 hover:text-primary-600 dark:hover:text-primary-400"
                      >
                        {item.name}
                      </a>
                    ) : (
                      <Link
                        to={item.href}
                        className="text-base text-gray-600 dark:text-gray-300 hover:text-primary-600 dark:hover:text-primary-400"
                      >
                        {item.name}
                      </Link>
                    )}
                  </li>
                ))}
              </ul>
            </div>
            <div>
              <h3 className="text-sm font-semibold text-gray-500 dark:text-gray-400 tracking-wider uppercase">
                Resources
              </h3>
              <ul className="mt-4 space-y-4">
                {navigation.resources.map((item) => (
                  <li key={item.name}>
                    {item.external ? (
                      <a
                        href={item.href}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-base text-gray-600 dark:text-gray-300 hover:text-primary-600 dark:hover:text-primary-400"
                      >
                        {item.name}
                      </a>
                    ) : (
                      <Link
                        to={item.href}
                        className="text-base text-gray-600 dark:text-gray-300 hover:text-primary-600 dark:hover:text-primary-400"
                      >
                        {item.name}
                      </Link>
                    )}
                  </li>
                ))}
              </ul>
            </div>
            <div>
              <h3 className="text-sm font-semibold text-gray-500 dark:text-gray-400 tracking-wider uppercase">
                GitHub
              </h3>
              <ul className="mt-4 space-y-4">
                {navigation.company.map((item) => (
                  <li key={item.name}>
                    {item.external ? (
                      <a
                        href={item.href}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-base text-gray-600 dark:text-gray-300 hover:text-primary-600 dark:hover:text-primary-400"
                      >
                        {item.name}
                      </a>
                    ) : (
                      <Link
                        to={item.href}
                        className="text-base text-gray-600 dark:text-gray-300 hover:text-primary-600 dark:hover:text-primary-400"
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
        <div className="mt-12 border-t border-gray-200 dark:border-gray-700 pt-8">
          <p className="text-base text-gray-500 dark:text-gray-400 text-center">
            &copy; {currentYear} GitHub Agent. All rights reserved.
          </p>
        </div>
      </div>
    </footer>
  );
};

export default Footer;
