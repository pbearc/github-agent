import React from "react";

const variants = {
  primary:
    "bg-primary-600 hover:bg-primary-700 text-white focus:ring-primary-500",
  secondary:
    "bg-gray-200 hover:bg-gray-300 text-gray-800 dark:bg-gray-700 dark:hover:bg-gray-600 dark:text-white focus:ring-gray-500",
  danger: "bg-red-600 hover:bg-red-700 text-white focus:ring-red-500",
  success: "bg-green-600 hover:bg-green-700 text-white focus:ring-green-500",
  outline:
    "bg-transparent border border-gray-300 dark:border-gray-600 hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-800 dark:text-white focus:ring-gray-500",
  ghost:
    "bg-transparent hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-800 dark:text-white focus:ring-gray-500",
  white:
    "bg-white hover:bg-gray-100 text-primary-600 border border-transparent focus:ring-white",
};

const sizes = {
  xs: "px-2 py-1 text-xs",
  sm: "px-3 py-1.5 text-sm",
  md: "px-4 py-2 text-sm",
  lg: "px-5 py-2.5 text-base",
  xl: "px-6 py-3 text-lg",
};

const Button = ({
  children,
  variant = "primary",
  size = "md",
  className = "",
  disabled = false,
  isLoading = false,
  fullWidth = false,
  icon = null,
  iconPosition = "left",
  ...props
}) => {
  return (
    <button
      className={`
        inline-flex items-center justify-center rounded-md 
        transition-colors duration-200 focus:outline-none focus:ring-2 focus:ring-offset-2
        ${variants[variant] || variants.primary}
        ${sizes[size] || sizes.md}
        ${fullWidth ? "w-full" : ""}
        ${disabled || isLoading ? "opacity-60 cursor-not-allowed" : ""}
        ${className}
      `}
      disabled={disabled || isLoading}
      {...props}
    >
      {isLoading && (
        <svg
          className="animate-spin -ml-1 mr-2 h-4 w-4"
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
        >
          <circle
            className="opacity-25"
            cx="12"
            cy="12"
            r="10"
            stroke="currentColor"
            strokeWidth="4"
          ></circle>
          <path
            className="opacity-75"
            fill="currentColor"
            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
          ></path>
        </svg>
      )}

      {icon && iconPosition === "left" && !isLoading && (
        <span className="mr-2">{icon}</span>
      )}

      {children}

      {icon && iconPosition === "right" && !isLoading && (
        <span className="ml-2">{icon}</span>
      )}
    </button>
  );
};

export default Button;
