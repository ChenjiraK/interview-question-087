# Frontend

Vue 3 + TypeScript frontend connected to the User API through Axios.

## Development

```bash
npm install
cp .env.example .env.local
npm run dev
```

Start the backend in `../back-go` on `http://localhost:8080`. Vite proxies browser
requests from `/api` to this backend, so development does not require backend CORS
configuration. Change `VITE_API_PROXY_TARGET` in `.env.local` if the backend runs
on a different host or port, then restart Vite.

- Opening the page calls `GET /persons?page=1&limit=10` to load the table. Previous/next
  buttons request the corresponding page and use the response's `pagination`
  metadata for the total count and page boundaries.
- Saving the add form calls `POST /person`, then refreshes the table.
- Clicking **View data** calls `GET /person/:id` before opening the details dialog.

The service maps API fields `firstname` and `lastname` to the form's `firstName`
and `lastName`. Only the four supported fields are sent when creating a user.
Loading, empty results, and request errors are displayed in the UI. A failed save
keeps the form open with its contents intact.

## Production

```bash
npm run build
```

The Vite development proxy is not included in the build. Configure the hosting
server to forward `/api/*` to the backend with the `/api` prefix removed, or set
`VITE_API_BASE_URL=https://your-api.example.com` before building and allow the
frontend origin in the backend's CORS configuration. `VITE_*` variables are public
and must not contain secrets.

## Checks

```bash
npm run type-check
npm test
npm run build
```

TypeScript runs in strict mode. `npm run type-check` checks Vue components,
application code, the Vite and Playwright configurations, and service and E2E tests. `npm run build`
runs these checks before creating the production bundle. Shared user and form
types are defined in `src/types/person.ts`.

Use Node.js 22.18+ (or a newer LTS release). Service tests run TypeScript directly
with Node's type stripping.

## End-to-end tests

Install Playwright's Chromium browser once, then run the tests:

```bash
npx playwright install chromium
npm run test:e2e
```

Playwright starts and stops a dedicated Vite server at `http://127.0.0.1:4173`.
Keep that port available; a separate development server is not required. API
requests are intercepted with isolated fixtures, so these browser tests do not
require the Go backend and do not verify backend/database integration. The test
server overrides `VITE_API_BASE_URL` to `/api`, regardless of local env settings.

The suite runs in desktop and mobile Chromium. It covers listing, pagination,
loading and empty states, request failures and retry, read-only details, creating
users, validation, cancellation, and save failures. Browser time is fixed so age
and future-birthday assertions remain stable.

```bash
npm run test:e2e -- --project=chromium  # Desktop only
npm run test:e2e:ui                   # Interactive runner
npm run test:e2e:report               # Open the last HTML report
```

Failed tests retain screenshots and traces in `test-results/`. The HTML report
is written to `playwright-report/`; both directories are ignored by Git.
For CI, run `npm ci`, `npx playwright install --with-deps chromium`, and
`npm run test:e2e`. See the [Playwright documentation](https://playwright.dev/docs/test-configuration)
for runner configuration.
