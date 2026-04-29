# Job API: Local Test Guide

## Prerequisites

- Docker + Docker Compose
- `curl` (for quick API checks)
- Postman (optional)

## 1) Prepare environment

If `.env` does not exist yet:

```bash
cp .env-sample .env
```

Check/edit values in `.env` as needed:

- `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`
- `PORT`
- `NGINX_HOST_PORT`
- `LOG_FILE_PATH`

## 2) Start the full stack

```bash
docker compose up --build -d
```

## 3) Verify containers are healthy

```bash
docker compose ps
```

Expected:

- `db` is `Up` and healthy
- `app` is `Up`
- `nginx` is `Up`

## 4) Smoke test through nginx

Use your nginx host port (default `8080`).

```bash
curl http://localhost:8080/
```

Expected:

```json
{"service":"jobs-api"}
```

```bash
curl http://localhost:8080/api/jobs
```

Expected on first run:

```json
[]
```

## 5) API tests (Create, Read, Delete)

### Create a job

```bash
curl -X POST http://localhost:8080/api/jobs \
  -H "Content-Type: application/json" \
  -d '{"title":"Backend Engineer","company":"Acme","description":"Golang role","status":"open"}'
```

Copy the `id` from the response.

### Read by ID

```bash
curl http://localhost:8080/api/jobs/<JOB_ID>
```

### Delete by ID

```bash
curl -i -X DELETE http://localhost:8080/api/jobs/<JOB_ID>
```

Expected: `204 No Content`

### Confirm deletion

```bash
curl -i http://localhost:8080/api/jobs/<JOB_ID>
```

Expected: `{"error":"Job not found"}`

## 6) Verify migrations (up + down)

On first startup:

```bash
docker compose logs app
```

You should see migration apply + complete logs.

Restart app to confirm applied up migrations are skipped:

```bash
docker compose restart app
docker compose logs app
```

For explicit migration direction (without booting API routes), use the migration command locally:

```bash
go run ./cmd/migrate -direction up
go run ./cmd/migrate -direction down
```

`down` rolls back one applied migration (latest first).

## 7) Verify file logging

Tail the application log file:

```bash
tail -f logs/app.log
```

You should see startup, migration, request, and shutdown entries.

## 8) Postman test (optional)

1. Import `postman/Jobs_API.postman_collection.json`
2. Ensure `baseUrl` is `http://localhost:8080`
3. Run:
   - Service info
   - Create job
   - Set `job_id` from response
   - Get job by ID
   - Delete job

## 9) Stop the stack

```bash
docker compose down
```

Remove volumes too (full DB reset):

```bash
docker compose down -v
```

