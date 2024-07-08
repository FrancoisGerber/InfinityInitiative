# Infinity Initiative

## An Online text based fantasy idle MMORPG

### Tech Used:
- **Language:** Go
- **Front end Framework:** HTMX
- **Back end Framework:** Echo
- **Realtime Connection:** Server-sent Events
- **Database:** Sqlite
- **ORM:** Gorm
- **CSS:** Tailwindcss + daisyui (npm required)
- **Pattern:** MVC
- **Tests:** Go Unit tests + Playwright Automated Testing
- **Hot reload:** Air

**Setup Webserver:**
`go mod tidy`

**Run server:**
`make dev`

**Run Tailwind Compiler:**
`./tailwindcss/npm install & ./tailwindcss/npm run dev`

**Run Tests:**
`make test`

**Build Binary for Prod:**
`make build`
