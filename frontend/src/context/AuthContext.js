import React, { createContext, useState, useEffect, useContext } from "react";

const AuthContext = createContext();

export const useAuth = () => useContext(AuthContext);

export const AuthProvider = ({ children }) => {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [user, setUser] = useState(null);
  const [githubToken, setGithubToken] = useState("");
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Check if user is authenticated on component mount
    const checkAuth = async () => {
      try {
        // Get token from localStorage
        const token = localStorage.getItem("githubToken");

        if (token) {
          setGithubToken(token);
          setIsAuthenticated(true);

          // For demo purposes - in a real app, verify token with backend
          setUser({
            username: localStorage.getItem("username") || "User",
            avatar:
              localStorage.getItem("avatar") ||
              "https://github.githubassets.com/images/modules/logos_page/GitHub-Mark.png",
          });
        }
      } catch (error) {
        console.error("Authentication error:", error);
        logout();
      } finally {
        setLoading(false);
      }
    };

    checkAuth();
  }, []);

  const login = (token, username, avatar) => {
    localStorage.setItem("githubToken", token);
    localStorage.setItem("username", username);
    localStorage.setItem("avatar", avatar);

    setGithubToken(token);
    setUser({ username, avatar });
    setIsAuthenticated(true);
  };

  const logout = () => {
    localStorage.removeItem("githubToken");
    localStorage.removeItem("username");
    localStorage.removeItem("avatar");

    setGithubToken("");
    setUser(null);
    setIsAuthenticated(false);
  };

  return (
    <AuthContext.Provider
      value={{
        isAuthenticated,
        user,
        githubToken,
        loading,
        login,
        logout,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
};
