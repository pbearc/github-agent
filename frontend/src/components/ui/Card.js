import React from "react";

const Card = ({
  children,
  className = "",
  title = "",
  subtitle = "",
  footer = null,
  noPadding = false,
  isHoverable = false,
  headerAction = null,
}) => {
  return (
    <div
      className={`
      bg-white dark:bg-dark-100 rounded-lg shadow-custom 
      ${isHoverable ? "hover:shadow-lg transition-shadow duration-300" : ""}
      ${className}
    `}
    >
      {(title || subtitle) && (
        <div className="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-700">
          <div>
            {title && (
              <h3 className="text-lg font-medium text-gray-900 dark:text-white">
                {title}
              </h3>
            )}
            {subtitle && (
              <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
                {subtitle}
              </p>
            )}
          </div>
          {headerAction && <div>{headerAction}</div>}
        </div>
      )}

      <div className={noPadding ? "" : "p-6"}>{children}</div>

      {footer && (
        <div className="px-6 py-4 border-t border-gray-200 dark:border-gray-700">
          {footer}
        </div>
      )}
    </div>
  );
};

export default Card;
