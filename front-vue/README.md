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

- Opening the page calls `GET /user` to load the table.
- Saving the add form calls `POST /user`, then refreshes the table.
- Clicking **View data** calls `GET /user/:id` before opening the details dialog.

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
application code, the Vite configuration, and service tests. `npm run build`
runs these checks before creating the production bundle. Shared user and form
types are defined in `src/types/person.ts`.

Use Node.js 22.18+ (or a newer LTS release). Service tests run TypeScript directly
with Node's type stripping.
