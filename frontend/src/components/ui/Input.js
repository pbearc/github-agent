import React from "react";

const Input = ({
  type = "text",
  label,
  id,
  name,
  value,
  onChange,
  placeholder,
  error,
  required = false,
  disabled = false,
  className = "",
  containerClassName = "",
  helpText,
  leftIcon,
  rightIcon,
  ...props
}) => {
  const inputId = id || name;

  return (
    <div className={`mb-4 ${containerClassName}`}>
      {label && (
        <label
          htmlFor={inputId}
          className="block mb-2 text-sm font-medium text-gray-700 dark:text-gray-200"
        >
          {label}
          {required && <span className="ml-1 text-red-500">*</span>}
        </label>
      )}

      <div className="relative">
        {leftIcon && (
          <div className="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none">
            {leftIcon}
          </div>
        )}

        <input
          type={type}
          id={inputId}
          name={name}
          value={value}
          onChange={onChange}
          placeholder={placeholder}
          disabled={disabled}
          className={`
            w-full px-4 py-2 border rounded-md transition-colors duration-200
            focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-primary-500
            ${
              error
                ? "border-red-500 dark:border-red-500"
                : "border-gray-300 dark:border-gray-600"
            }
            ${
              disabled
                ? "bg-gray-100 dark:bg-gray-700 cursor-not-allowed"
                : "bg-white dark:bg-dark-100"
            }
            ${leftIcon ? "pl-10" : ""}
            ${rightIcon ? "pr-10" : ""}
            text-gray-900 dark:text-white
            ${className}
          `}
          required={required}
          {...props}
        />

        {rightIcon && (
          <div className="absolute inset-y-0 right-0 flex items-center pr-3">
            {rightIcon}
          </div>
        )}
      </div>

      {helpText && !error && (
        <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
          {helpText}
        </p>
      )}

      {error && <p className="mt-1 text-sm text-red-500">{error}</p>}
    </div>
  );
};

export default Input;
