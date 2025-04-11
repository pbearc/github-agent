import React, { useState } from "react";
import { BrowserRouter as Router, Routes, Route } from "react-router-dom";
import { Toaster } from "sonner";

// Layout Components
import Navbar from "./components/layout/Navbar";
import Sidebar from "./components/layout/Sidebar";
import Footer from "./components/layout/Footer";

// Pages
import Dashboard from "./pages/Dashboard";
import RepositoryInfo from "./pages/RepositoryInfo";
import GenerateReadme from "./pages/GenerateReadme";
import GenerateDockerfile from "./pages/GenerateDockerfile";
import CodeComments from "./pages/CodeComments";
import CodeRefactor from "./pages/CodeRefactor";
import CodeSearch from "./pages/CodeSearch";
import Settings from "./pages/Settings";
import PRSummary from "./pages/PRSummary";
import CodeWalkthrough from "./pages/CodeWalkthrough";
import FunctionExplainer from "./pages/FunctionExplainer";
import ArchitectureAndPractices from "./pages/ArchitectureAndPractices";

// Context
import { ThemeProvider } from "./context/ThemeContext";

function App() {
  const [sidebarOpen, setSidebarOpen] = useState(false);

  const toggleSidebar = () => {
    setSidebarOpen(!sidebarOpen);
  };

  return (
    <ThemeProvider>
      <Router>
        <div className="flex flex-col min-h-screen bg-gray-900 dark:bg-dark-200 transition-colors duration-300">
          <Navbar toggleSidebar={toggleSidebar} />

          <div className="flex flex-1 relative">
            <main className="flex-1 w-full max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
              <Routes>
                <Route path="/" element={<Dashboard />} />
                <Route path="/repository" element={<RepositoryInfo />} />

                {/* Code understanding routes */}
                <Route path="/search" element={<CodeSearch />} />
                <Route
                  path="/navigate/walkthrough"
                  element={<CodeWalkthrough />}
                />
                <Route
                  path="/navigate/function"
                  element={<FunctionExplainer />}
                />
                <Route
                  path="/navigate/architecture"
                  element={<ArchitectureAndPractices />}
                />
                <Route
                  path="/navigate/practices"
                  element={<ArchitectureAndPractices />}
                />

                {/* Helper functionality routes */}
                <Route path="/generate/readme" element={<GenerateReadme />} />
                <Route
                  path="/generate/dockerfile"
                  element={<GenerateDockerfile />}
                />
                <Route path="/generate/comments" element={<CodeComments />} />
                <Route path="/generate/refactor" element={<CodeRefactor />} />
                <Route path="/pr-summary" element={<PRSummary />} />
                <Route path="/settings" element={<Settings />} />
              </Routes>
            </main>

            {/* Right-side sidebar */}
            <Sidebar isOpen={sidebarOpen} toggleSidebar={toggleSidebar} />
          </div>

          {/* <Footer /> */}
        </div>

        <Toaster position="top-right" richColors />
      </Router>
    </ThemeProvider>
  );
}

export default App;
