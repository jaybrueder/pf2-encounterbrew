# Contributing to PF2 Encounterbrew

Thank you for your interest in contributing to PF2 Encounterbrew! This project is a web application for managing Pathfinder 2e encounters, built with Go, HTMX, Templ, Echo, and PostgreSQL.

## Welcome New Contributors

We welcome contributions from developers of all skill levels. Whether you're fixing bugs, adding features, improving documentation, or helping with testing, your contributions are valued. If you're new to the project or need help getting started, feel free to open an issue with your questions.

## Development Setup

### Prerequisites

Before you begin, ensure you have the following installed:

- **Go** (1.23.1 or later)
- **Docker** and **Docker Compose**
- **Node.js** and **npm**
- **pre-commit** (for code quality hooks)
- **golangci-lint** (for Go linting)

### Required Global Dependencies

```bash
# Install TailwindCSS
npm install -g tailwindcss@3.4.17

# Install Templ (Go template engine)
go install github.com/a-h/templ/cmd/templ@v0.2.793

# Install Air (for live reload during development)
go install github.com/air-verse/air@latest

# Install pre-commit
pip install pre-commit
```

### Local Development

1. **Clone the repository**
   ```bash
   git clone https://github.com/jaybrueder/pf2-encounterbrew.git
   cd pf2-encounterbrew
   ```

2. **Set up pre-commit hooks**
   ```bash
   pre-commit install
   ```

3. **Start the development database**
   ```bash
   docker compose -f docker-compose.dev.yml up -d
   ```

4. **Set up environment variables**
   ```bash
   export PF2ENCOUNTERBREW_DB_DSN=postgres://admin:admin@localhost/encounterbrew?sslmode=disable
   ```

5. **Build and run the application**
   ```bash
   make build
   make run
   ```

   Or for live reload during development:
   ```bash
   make watch
   ```

The application will be available at `http://localhost:8080`.

## Project Structure

```
├── cmd/
│   ├── api/          # API server entry point
│   ├── web/          # Web handlers and templates
│   └── seed/         # Database seeding utilities
├── internal/
│   ├── database/     # Database service layer
│   ├── models/       # Data models and business logic
│   ├── seeder/       # Database seeding logic
│   └── server/       # Server configuration
├── migrations/       # Database migration files
├── data/            # Static data files (monsters, conditions, etc.)
└── tmp/             # Temporary build files
```

## Technology Stack

- **Backend**: Go with Echo framework
- **Frontend**: HTMX + Templ templates
- **Styling**: TailwindCSS
- **Database**: PostgreSQL
- **Development**: Air (live reload), pre-commit hooks

## How to Contribute

### Workflow

We use a Git flow approach:
- `main` - Main production-ready branch.
- Feature branches should be created from `main` and used for feature development

New releases are created by adding a new release in the GitHub repo.

### Making Changes

1. **Fork the repository** or create a new branch from `main`
2. Pick an existing **Issue** or create a new one
3. **Create a feature branch** from that **Issue** and switch into it on your workstation
   ```bash
   git checkout -b NAME_OF_FEATURE_BRANCH
   ```

4. **Make your changes** following our coding standards (don't forget to write **tests** if applicable!)
5. **Test your changes** thoroughly:
   ```bash
   make test
   ```

6. **Commit your changes** with clear, descriptive messages
7. **Push your branch** and create a **Pull Request**

### Pull Request Guidelines

- **Title**: Use imperative mood (e.g., "Add monster search functionality", "Fix initiative sorting bug")
- **Description**: Provide clear details about what changes were made and why
- **Testing**: Describe how you tested your changes
- **Breaking Changes**: Clearly note any breaking changes

### Code Style

We use automated code formatting and linting:

- **Go**: Code is formatted with `gofmt` and linted with `golangci-lint`
- **Templates**: Templ files should follow Go conventions
- **Styles**: TailwindCSS for all styling

Pre-commit hooks will automatically check code style. You can also run:
```bash
# Run all pre-commit checks
pre-commit run --all-files

# Run specific checks
go fmt ./...
golangci-lint run
```

### Testing

- Write tests for new functionality
- Ensure existing tests still pass
- Run the full test suite: `make test`
- Manual testing should cover the web interface

### Database Changes

If your changes require database modifications:

1. Create new migration files in the `migrations/` directory
2. Follow the naming convention: `XXXXXX_description.up.sql` and `XXXXXX_description.down.sql`
3. Test both up and down migrations
4. Update seed data if necessary

### Adding Game Content

When adding new Pathfinder 2e content (monsters, spells, etc.):

- Name entities exactly as they appear in source material
- Replace square brackets `[` `]` with parentheses or omit them
- Place data files in appropriate directories under `data/`
- Follow existing JSON structure and validation patterns
- Ensure compliance with **OGL/ORC** licensing

## Getting Help

- **Issues**: Check existing GitHub issues or create a new one
- **Questions**: Open a discussion or issue for general questions
- **Documentation**: Check the [README.md](README.md) for basic setup

## Code of Conduct

Please be respectful and inclusive in all interactions. We want this to be a welcoming community for all contributors.

## License

By contributing to this project, you agree that your contributions will be licensed under the Apache License 2.0. See [LICENSE](LICENSE) for details.

---

Thank you for contributing to **PF2 Encounterbrew**! Every contribution, no matter how small, helps make this tool better for the Pathfinder 2e community.
