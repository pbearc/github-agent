import React from "react";
import PropTypes from "prop-types";

const Button = ({
  children,
  type = "button",
  variant = "primary",
  size = "md",
  fullWidth = false,
  isLoading = false,
  disabled = false,
  onClick,
  className = "",
  ...props
}) => {
  const baseStyles =
    "inline-flex items-center justify-center font-medium rounded-md focus:outline-none transition-all duration-200";

  const sizeStyles = {
    xs: "px-2 py-1 text-xs",
    sm: "px-3 py-1.5 text-sm",
    md: "px-4 py-2 text-base",
    lg: "px-6 py-3 text-lg",
  };

  const variantStyles = {
    primary:
      "bg-gradient-to-r from-primary-600 to-secondary-600 text-white hover:from-primary-500 hover:to-secondary-500 shadow-lg",
    secondary:
      "bg-gradient-to-r from-secondary-600 to-pink-600 text-white hover:from-secondary-500 hover:to-pink-500 shadow-lg",
    accent:
      "bg-gradient-to-r from-orange-500 to-pink-600 text-white hover:from-orange-400 hover:to-pink-500 shadow-lg",
    success:
      "bg-gradient-to-r from-lime-600 to-primary-600 text-white hover:from-lime-500 hover:to-primary-500 shadow-lg",
    outline:
      "bg-transparent border border-gray-700 text-gray-300 hover:border-gray-600 hover:bg-gray-800",
    white: "bg-white text-gray-900 hover:bg-gray-100 shadow-lg",
    danger:
      "bg-gradient-to-r from-pink-600 to-orange-600 text-white hover:from-pink-500 hover:to-orange-500 shadow-lg",
  };

  const disabledStyles = "opacity-50 cursor-not-allowed";
  const loadingStyles = "cursor-wait";
  const fullWidthStyles = "w-full";

  let combinedStyles = `${baseStyles} ${sizeStyles[size]} ${variantStyles[variant]}`;

  if (disabled) {
    combinedStyles += ` ${disabledStyles}`;
  }

  if (isLoading) {
    combinedStyles += ` ${loadingStyles}`;
  }

  if (fullWidth) {
    combinedStyles += ` ${fullWidthStyles}`;
  }

  // Apply custom className at the end to allow overrides
  if (className) {
    combinedStyles += ` ${className}`;
  }

  return (
    <button
      type={type}
      className={combinedStyles}
      onClick={onClick}
      disabled={disabled || isLoading}
      {...props}
    >
      {isLoading && (
        <svg
          className="animate-spin -ml-1 mr-2 h-4 w-4 text-white"
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
      {children}
    </button>
  );
};

Button.propTypes = {
  children: PropTypes.node.isRequired,
  type: PropTypes.oneOf(["button", "submit", "reset"]),
  variant: PropTypes.oneOf([
    "primary",
    "secondary",
    "accent",
    "success",
    "outline",
    "white",
    "danger",
  ]),
  size: PropTypes.oneOf(["xs", "sm", "md", "lg"]),
  fullWidth: PropTypes.bool,
  isLoading: PropTypes.bool,
  disabled: PropTypes.bool,
  onClick: PropTypes.func,
  className: PropTypes.string,
};

export default Button;
