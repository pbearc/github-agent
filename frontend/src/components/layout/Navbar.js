import React from "react";
import { Link } from "react-router-dom";

const Navbar = ({ toggleSidebar }) => {
  return (
    <header className="bg-dark-100 border-b border-gray-800 shadow-lg transition-colors duration-300 relative z-10">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between items-center h-16">
          {/* Logo */}
          <div className="flex-shrink-0 flex items-center">
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
              <span className="text-xl font-bold text-white">
                Code Navigator
              </span>
            </Link>
          </div>

          {/* Main navigation - only show a few key items */}
          <nav className="hidden md:flex items-center space-x-1">
            <Link
              to="/"
              className="text-white hover:bg-gray-800 px-3 py-2 rounded-md text-sm font-medium transition-colors"
            >
              Home
            </Link>
            <Link
              to="/search"
              className="text-white hover:bg-gray-800 px-3 py-2 rounded-md text-sm font-medium transition-colors"
            >
              Code Search
            </Link>
            <Link
              to="/navigate/walkthrough"
              className="text-white hover:bg-gray-800 px-3 py-2 rounded-md text-sm font-medium transition-colors"
            >
              Walkthrough
            </Link>
            <Link
              to="/navigate/function"
              className="text-white hover:bg-gray-800 px-3 py-2 rounded-md text-sm font-medium transition-colors"
            >
              Function Explainer
            </Link>
            <a
              href="https://github.com/pbearc/github-agent"
              target="_blank"
              rel="noopener noreferrer"
              className="text-white hover:bg-gray-800 px-3 py-2 rounded-md text-sm font-medium transition-colors"
            >
              GitHub
            </a>
          </nav>

          {/* Right section - Menu button only */}
          <div className="flex items-center">
            {/* Menu button */}
            <button
              onClick={toggleSidebar}
              className="p-2 rounded-md text-white hover:bg-gray-800 focus:outline-none transition-colors relative"
              aria-label="Open menu"
            >
              <div className="absolute inset-0 bg-gradient-to-r from-primary-500 to-secondary-500 rounded-md opacity-0 hover:opacity-10 transition-opacity"></div>
              <svg
                xmlns="http://www.w3.org/2000/svg"
                className="h-6 w-6 relative"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M4 6h16M4 12h16M4 18h16"
                />
              </svg>
            </button>
          </div>
        </div>
      </div>
    </header>
  );
};

export default Navbar;
