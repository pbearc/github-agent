import React from "react";

const Select = ({
  label,
  id,
  name,
  value,
  onChange,
  options = [],
  error,
  required = false,
  disabled = false,
  className = "",
  containerClassName = "",
  helpText,
  placeholder = "Select an option",
  ...props
}) => {
  const selectId = id || name;

  return (
    <div className={`mb-4 ${containerClassName}`}>
      {label && (
        <label
          htmlFor={selectId}
          className="block mb-2 text-sm font-medium text-gray-700 dark:text-gray-200"
        >
          {label}
          {required && <span className="ml-1 text-red-500">*</span>}
        </label>
      )}

      <select
        id={selectId}
        name={name}
        value={value}
        onChange={onChange}
        disabled={disabled}
        className={`
          w-full px-4 py-2 border rounded-md appearance-none transition-colors duration-200
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
          text-gray-900 dark:text-white
          ${className}
        `}
        required={required}
        {...props}
      >
        <option value="" disabled>
          {placeholder}
        </option>

        {options.map((option) => (
          <option
            key={option.value}
            value={option.value}
            disabled={option.disabled}
          >
            {option.label}
          </option>
        ))}
      </select>

      {helpText && !error && (
        <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
          {helpText}
        </p>
      )}

      {error && <p className="mt-1 text-sm text-red-500">{error}</p>}
    </div>
  );
};

export default Select;
