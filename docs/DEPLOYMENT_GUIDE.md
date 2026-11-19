# Deployment Guide

## Local development

1. Copy `.env.example` to `.env` and replace DEV placeholders.
2. Run `docker compose up --build`.
3. Call `/health/live` and `/health/ready`.
4. Run the Flutter client in mock mode or HTTP mode.

## Commands

```bash
docker compose config
docker compose up --build
go test ./...
flutter pub get
flutter analyze
flutter test
```

Docker startup and Flutter device execution require local infrastructure and must be reported with actual output.

## Safety

- Never commit `.env`.
- Never run Beta or Production with the DEV JWT placeholder.
- Run migrations and smoke tests before exposing a new deployment.
