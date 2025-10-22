# Docker: single-container setup (backend serves frontend)

This repository includes a multi-stage `Dockerfile` that:

- Builds the frontend (Vite) in a node builder stage and creates `frontend/dist`.
- Builds the Go backend in a Go builder stage and copies `frontend/dist` into the backend folder so the server serves static assets.
- Applies SQLite `up.sql` migration files sequentially (one-by-one) during the image build to produce a pre-migrated DB included in the image.

Files added:

- `Dockerfile` - multi-stage build, builds frontend + backend, applies migrations.
- `docker-compose.yml` - single service that exposes port 8080 and mounts uploads.
 - `docker-compose.multi.yml` - multi-container compose that runs `backend` (Go) and `frontend` (nginx) services.

How to build & run locally

1. Build and start the containers (detached):

```bash
docker compose -f docker-compose.multi.yml up --build -d
```

2. Check containers and sizes:

```bash
docker ps -a
docker image ls
```

Look for a container named `social-network-app` and an image built from this repo. Container size should be non-zero (docker ps -a shows container size column on some platforms) and `docker image ls` will show image sizes.

3. Open the app in your browser:

- Frontend (nginx) serves the SPA at http://localhost:5173 and backend API is at http://localhost:8080

DB & migrations (development-friendly)
-------------------------------------
The compose file mounts `./backend/socialnetwork.db` into the backend container so the SQLite DB is persisted on the host across container restarts. Migrations are applied at container startup (not build time) by an entrypoint script which iterates the `*up.sql` files in `backend/db/migrations/sqlite` and pipes them into the mounted DB file.

If the host DB file does not exist, the container will create it and then apply migrations. This makes development iterations and data persistence easier.

Audit notes for your questions

- You asked whether there are two containers (backend + frontend): with `docker-compose.multi.yml` you will have two containers named `social-backend` and `social-frontend`.
- To confirm they have non-zero sizes, run `docker ps -a` and `docker image ls`. The images will list sizes; if an image size is 0, that indicates a build issue.
- To access the app in a browser, open http://localhost:5173 (frontend). The frontend will make API calls to the backend at port 8080.

Audit checklist (what you asked):

- Confirm there are two containers (backend and frontend): this Docker setup uses a single container where the backend serves the frontend. If you specifically need two containers (separate frontend + backend), see the "Optional: multi-container" section below.
- Confirm both containers have non-zero sizes: run `docker ps -a` and `docker image ls` — you should see non-zero sizes for the built images/containers.
- Access via browser: open http://localhost:8080 and verify the app loads.

If you want true separation (frontend container + backend container), I can add a `docker-compose.multi.yml` and separate Dockerfiles that run the frontend in its own container (dev or static served by nginx) and the backend in another.

Notes & limitations

- Migrations are applied at image build time by piping each `*up.sql` into `sqlite3`. This means the image contains a pre-migrated `socialnetwork.db`. If you prefer migrations run at container startup, I can change Dockerfile to run migrations in the entrypoint.
- Building the frontend inside the image requires network access to npm registries during build. If your environment blocks outbound network, build locally and `docker build` with `--no-cache` and offline artifacts.
