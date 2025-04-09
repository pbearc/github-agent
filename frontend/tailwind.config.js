/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./src/**/*.{js,jsx,ts,tsx}"],
  darkMode: "class",
  theme: {
    extend: {
      colors: {
        // Vibrant primary color (teal-cyan)
        primary: {
          50: "#e6fcff",
          100: "#c9f9ff",
          200: "#9cf2ff",
          300: "#5ee6ff",
          400: "#29d3ff",
          500: "#0cbcff",
          600: "#009aeb",
          700: "#0078c2",
          800: "#00619d",
          900: "#01517f",
          950: "#00315a",
        },
        // Vibrant secondary color (purple-pink)
        secondary: {
          50: "#fdf2ff",
          100: "#fae5ff",
          200: "#f5c9ff",
          300: "#eb9fff",
          400: "#e16aff",
          500: "#d139ff",
          600: "#b423e6",
          700: "#931bbc",
          800: "#7a1b99",
          900: "#631a7a",
          950: "#450854",
        },
        // Bright accent 1 (orange)
        orange: {
          50: "#fff8ed",
          100: "#ffeecf",
          200: "#ffd999",
          300: "#ffbe5c",
          400: "#ffa01f",
          500: "#ff7b00",
          600: "#e65c00",
          700: "#bf4200",
          800: "#9a3500",
          900: "#7e2e05",
          950: "#451501",
        },
        // Bright accent 2 (yellow-lime)
        lime: {
          50: "#f8ffe6",
          100: "#eeffcc",
          200: "#d7ff99",
          300: "#b8f959",
          400: "#9bee29",
          500: "#7cd60a",
          600: "#5caa05",
          700: "#468009",
          800: "#39650e",
          900: "#31550f",
          950: "#173004",
        },
        // Bright accent 3 (pink-red)
        pink: {
          50: "#fff0f5",
          100: "#ffe1ed",
          200: "#ffc2db",
          300: "#ff92bb",
          400: "#ff5090",
          500: "#ff206a",
          600: "#ed0051",
          700: "#ca0046",
          800: "#a80044",
          900: "#8c0241",
          950: "#550022",
        },
        // Dark theme colors
        dark: {
          100: "#171721",
          200: "#13131a",
          300: "#0f0f14",
        },
      },
      fontFamily: {
        sans: ["Inter", "sans-serif"],
      },
      boxShadow: {
        "custom": "0 4px 14px 0 rgba(0, 0, 0, 0.1)",
        "glow-primary": "0 0 20px rgba(12, 188, 255, 0.2)",
        "glow-secondary": "0 0 20px rgba(209, 57, 255, 0.2)",
        "glow-orange": "0 0 20px rgba(255, 123, 0, 0.2)",
        "glow-lime": "0 0 20px rgba(124, 214, 10, 0.2)",
        "glow-pink": "0 0 20px rgba(255, 32, 106, 0.2)",
      },
      backgroundImage: {
        "gradient-radial": "radial-gradient(var(--tw-gradient-stops))",
        "gradient-conic": "conic-gradient(var(--tw-gradient-stops))",
      },
      animation: {
        "pulse-slow": "pulse 4s cubic-bezier(0.4, 0, 0.6, 1) infinite",
        "float": "float 6s ease-in-out infinite",
      },
      keyframes: {
        float: {
          "0%, 100%": { transform: "translateY(0)" },
          "50%": { transform: "translateY(-10px)" },
        },
      },
    },
  },
  plugins: [],
};
