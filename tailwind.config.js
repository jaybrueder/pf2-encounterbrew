/** @type {import('tailwindcss').Config} */
module.exports = {
    content: [
      "./cmd/web/**/*.html", "./cmd/web/**/*.templ",
    ],
    theme: {
        extend: {},
        fontFamily: {
          'sans': ['"Roboto"', 'sans-serif'],
        },
    },
    plugins: [],
}
