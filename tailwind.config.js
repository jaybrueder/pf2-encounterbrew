/** @type {import('tailwindcss').Config} */
module.exports = {
    content: [
      "./cmd/web/**/*.html", "./cmd/web/**/*.templ",
    ],
    theme: {
        extend: {
          fontWeight: {
            normal: 300,
            bold: 700,
          },
        },
        fontFamily: {
          'sans': ['"Roboto"', 'sans-serif'],
        },
    },
    plugins: [],
}
