# Feature: Settings

Mirrors `frontend-web/src/app/(dashboard)/dashboard/settings/*`.

## Wired today

- Palette selector (5 palettes from `core.presentation.theme.Palettes`)
- Dark mode (System / Light / Dark)

## Add next

| Section       | Web route                              | Endpoints |
|---------------|----------------------------------------|-----------|
| Profile       | `/dashboard/settings/profile`          | `GET/PATCH /users/me` |
| Security      | `/dashboard/settings/security`         | `POST /auth/change-password` |
| Sessions      | `/dashboard/settings/sessions`         | `GET /auth/sessions`, `DELETE /auth/sessions/{id}` |
| Notifications | `/dashboard/settings/notifications`    | `GET/PATCH /notifications/preferences` |
| Developer     | `/dashboard/settings/developer`        | API key CRUD under `/api-keys` |
| Tenant        | `/dashboard/settings/tenant`           | `GET/PATCH /tenants/me` |

Each one is a sub-screen + sub-ScreenModel; register in `SettingsModule.kt`.
