// components/ui/Tabs.jsx
import React from "react";

const Tabs = ({ tabs, activeTab, onChange, className = "" }) => {
  return (
    <div className={`border-b border-gray-800 ${className}`}>
      <div className="flex space-x-4 overflow-x-auto pb-2">
        {tabs.map((tab, index) => (
          <button
            key={index}
            onClick={() => onChange(index)}
            className={`py-2 px-3 text-sm font-medium transition-colors rounded-t-md focus:outline-none ${
              activeTab === index
                ? "text-primary-500 border-b-2 border-primary-500"
                : "text-gray-400 hover:text-gray-200"
            }`}
          >
            {tab}
          </button>
        ))}
      </div>
    </div>
  );
};

export default Tabs;
