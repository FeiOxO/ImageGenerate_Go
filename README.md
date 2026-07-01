# ImageGenerate Go Backend

Go backend for the AI image generation demo.

## Features

- JWT authentication
- User registration and login
- Persistent token validation
- MySQL persistence for users and generated image records
- Worker pool for image generation jobs
- Base64 or URL image result saving as PNG
- Structured JSON logs

## Local Setup

1. Copy environment config:

```bash
cp .env.example .env
```

2. Update `.env` with MySQL and image API credentials.

3. Start MySQL and create the database:

```sql
CREATE DATABASE ai_image_demo CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

4. Run the server:

```bash
go run ./cmd/api
```

The service listens on port `3000` by default.

## Build

```bash
go build -o ai-image-api ./cmd/api
```

## API

- `POST /api/auth/register`
- `POST /api/auth/login`
- `GET /api/auth/me`
- `POST /api/images/generate`
- `GET /api/images`
- `GET /api/health`
