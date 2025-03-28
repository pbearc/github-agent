import React from "react";
import { NavLink } from "react-router-dom";
import { Transition } from "@headlessui/react";

const Sidebar = ({ isOpen, toggleSidebar }) => {
  const navItems = [
    {
      section: "General",
      items: [
        {
          name: "Dashboard",
          path: "/",
          icon: (
            <svg
              xmlns="http://www.w3.org/2000/svg"
              className="h-5 w-5"
              viewBox="0 0 20 20"
              fill="currentColor"
            >
              <path d="M10.707 2.293a1 1 0 00-1.414 0l-7 7a1 1 0 001.414 1.414L4 10.414V17a1 1 0 001 1h2a1 1 0 001-1v-2a1 1 0 011-1h2a1 1 0 011 1v2a1 1 0 001 1h2a1 1 0 001-1v-6.586l.293.293a1 1 0 001.414-1.414l-7-7z" />
            </svg>
          ),
        },
        {
          name: "Repository Info",
          path: "/repository",
          icon: (
            <svg
              xmlns="http://www.w3.org/2000/svg"
              className="h-5 w-5"
              viewBox="0 0 20 20"
              fill="currentColor"
            >
              <path
                fillRule="evenodd"
                d="M4 4a2 2 0 012-2h4.586A2 2 0 0112 2.586L15.414 6A2 2 0 0116 7.414V16a2 2 0 01-2 2H6a2 2 0 01-2-2V4z"
                clipRule="evenodd"
              />
            </svg>
          ),
        },
      ],
    },
    {
      section: "Generate",
      items: [
        {
          name: "Generate README",
          path: "/generate/readme",
          icon: (
            <svg
              xmlns="http://www.w3.org/2000/svg"
              className="h-5 w-5"
              viewBox="0 0 20 20"
              fill="currentColor"
            >
              <path
                fillRule="evenodd"
                d="M4 4a2 2 0 012-2h4.586A2 2 0 0112 2.586L15.414 6A2 2 0 0116 7.414V16a2 2 0 01-2 2H6a2 2 0 01-2-2V4z"
                clipRule="evenodd"
              />
            </svg>
          ),
        },
        {
          name: "Generate Dockerfile",
          path: "/generate/dockerfile",
          icon: (
            <svg
              xmlns="http://www.w3.org/2000/svg"
              className="h-5 w-5"
              viewBox="0 0 20 20"
              fill="currentColor"
            >
              <path
                fillRule="evenodd"
                d="M3 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1z"
                clipRule="evenodd"
              />
            </svg>
          ),
        },
        {
          name: "Add Comments",
          path: "/generate/comments",
          icon: (
            <svg
              xmlns="http://www.w3.org/2000/svg"
              className="h-5 w-5"
              viewBox="0 0 20 20"
              fill="currentColor"
            >
              <path
                fillRule="evenodd"
                d="M18 13V5a2 2 0 00-2-2H4a2 2 0 00-2 2v8a2 2 0 002 2h3l3 3 3-3h3a2 2 0 002-2zM5 7a1 1 0 011-1h8a1 1 0 110 2H6a1 1 0 01-1-1zm1 3a1 1 0 100 2h3a1 1 0 100-2H6z"
                clipRule="evenodd"
              />
            </svg>
          ),
        },
        {
          name: "Refactor Code",
          path: "/generate/refactor",
          icon: (
            <svg
              xmlns="http://www.w3.org/2000/svg"
              className="h-5 w-5"
              viewBox="0 0 20 20"
              fill="currentColor"
            >
              <path
                fillRule="evenodd"
                d="M12.316 3.051a1 1 0 01.633 1.265l-4 12a1 1 0 11-1.898-.632l4-12a1 1 0 011.265-.633zM5.707 6.293a1 1 0 010 1.414L3.414 10l2.293 2.293a1 1 0 11-1.414 1.414l-3-3a1 1 0 010-1.414l3-3a1 1 0 011.414 0zm8.586 0a1 1 0 011.414 0l3 3a1 1 0 010 1.414l-3 3a1 1 0 01-1.414-1.414L16.586 10l-2.293-2.293a1 1 0 010-1.414z"
                clipRule="evenodd"
              />
            </svg>
          ),
        },
      ],
    },
    {
      section: "Search",
      items: [
        {
          name: "Code Search",
          path: "/search",
          icon: (
            <svg
              xmlns="http://www.w3.org/2000/svg"
              className="h-5 w-5"
              viewBox="0 0 20 20"
              fill="currentColor"
            >
              <path
                fillRule="evenodd"
                d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l4.817 4.817a1 1 0 01-1.414 1.414l-4.816-4.816A6 6 0 012 8z"
                clipRule="evenodd"
              />
            </svg>
          ),
        },
      ],
    },
    {
      section: "Settings",
      items: [
        {
          name: "Settings",
          path: "/settings",
          icon: (
            <svg
              xmlns="http://www.w3.org/2000/svg"
              className="h-5 w-5"
              viewBox="0 0 20 20"
              fill="currentColor"
            >
              <path
                fillRule="evenodd"
                d="M11.49 3.17c-.38-1.56-2.6-1.56-2.98 0a1.532 1.532 0 01-2.286.948c-1.372-.836-2.942.734-2.106 2.106.54.886.061 2.042-.947 2.287-1.561.379-1.561 2.6 0 2.978a1.532 1.532 0 01.947 2.287c-.836 1.372.734 2.942 2.106 2.106a1.532 1.532 0 012.287.947c.379 1.561 2.6 1.561 2.978 0a1.533 1.533 0 012.287-.947c1.372.836 2.942-.734 2.106-2.106a1.533 1.533 0 01.947-2.287c1.561-.379 1.561-2.6 0-2.978a1.532 1.532 0 01-.947-2.287c.836-1.372-.734-2.942-2.106-2.106a1.532 1.532 0 01-2.287-.947zM10 13a3 3 0 100-6 3 3 0 000 6z"
                clipRule="evenodd"
              />
            </svg>
          ),
        },
      ],
    },
  ];

  // Sidebar content
  const SidebarContent = () => (
    <div className="flex flex-col h-full py-6 overflow-y-auto bg-white dark:bg-dark-100 shadow-xl">
      <div className="flex items-center justify-between px-6 mb-6">
        <h2 className="text-lg font-semibold text-gray-900 dark:text-white">
          Menu
        </h2>
        <button
          onClick={toggleSidebar}
          className="text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200 focus:outline-none"
          aria-label="Close sidebar"
        >
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
              d="M6 18L18 6M6 6l12 12"
            />
          </svg>
        </button>
      </div>

      <div className="mt-2 flex-1 px-3">
        {navItems.map((section) => (
          <div key={section.section} className="mb-8">
            <h3 className="px-3 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">
              {section.section}
            </h3>
            <div className="mt-2 space-y-1">
              {section.items.map((item) => (
                <NavLink
                  key={item.path}
                  to={item.path}
                  className={({ isActive }) =>
                    `flex items-center px-3 py-2 text-sm font-medium rounded-md transition-all duration-200 
                    ${
                      isActive
                        ? "bg-primary-50 dark:bg-primary-900 text-primary-700 dark:text-primary-100"
                        : "text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800"
                    }`
                  }
                  onClick={() => toggleSidebar()}
                >
                  <span className="mr-3">{item.icon}</span>
                  {item.name}
                </NavLink>
              ))}
            </div>
          </div>
        ))}
      </div>

      <div className="px-6 py-4 border-t border-gray-200 dark:border-gray-700">
        <div className="flex flex-col space-y-4">
          <a
            href="https://github.com/pbearc/github-agent"
            target="_blank"
            rel="noopener noreferrer"
            className="flex items-center text-gray-700 dark:text-gray-300 hover:text-primary-600 dark:hover:text-primary-400"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              className="h-5 w-5 mr-2"
              viewBox="0 0 20 20"
              fill="currentColor"
            >
              <path
                fillRule="evenodd"
                d="M10 1C5.03 1 1 5.03 1 10c0 3.97 2.58 7.3 6.16 8.46.45.08.62-.19.62-.43v-1.5c-2.5.54-3.03-1.06-3.03-1.06-.41-1.05-1.03-1.33-1.03-1.33-.83-.56.07-.55.07-.55.92.06 1.42.95 1.42.95.82 1.41 2.16 1 2.69.77.08-.6.32-1 .58-1.23-2.02-.23-4.15-1.01-4.15-4.49 0-1 .36-1.81.93-2.45-.09-.22-.4-1.13.09-2.35 0 0 .76-.24 2.48.92.72-.2 1.49-.3 2.26-.3s1.54.1 2.26.3c1.72-1.17 2.48-.92 2.48-.92.49 1.22.18 2.13.09 2.35.57.64.93 1.45.93 2.45 0 3.49-2.13 4.27-4.16 4.49.33.29.62.84.62 1.7v2.52c0 .24.16.52.63.43C16.42 17.3 19 13.97 19 10c0-4.97-4.03-9-9-9z"
                clipRule="evenodd"
              />
            </svg>
            View on GitHub
          </a>
        </div>
      </div>
    </div>
  );

  return (
    <Transition
      show={isOpen}
      as={React.Fragment}
      enter="transition-transform ease-in-out duration-300"
      enterFrom="transform translate-x-full"
      enterTo="transform translate-x-0"
      leave="transition-transform ease-in-out duration-300"
      leaveFrom="transform translate-x-0"
      leaveTo="transform translate-x-full"
    >
      {/* Overlay */}
      <div className="fixed inset-0 z-20">
        {/* Backdrop */}
        <div
          className="absolute inset-0 bg-gray-600 bg-opacity-50 transition-opacity"
          onClick={toggleSidebar}
        />

        {/* Sidebar panel */}
        <div className="absolute inset-y-0 right-0 max-w-xs w-full flex">
          <SidebarContent />
        </div>
      </div>
    </Transition>
  );
};

export default Sidebar;
