# Feature: User

Profile, memberships, invites. Maps to the web's `(dashboard)/dashboard/settings/profile`
and the backend's `internal/modules/user`.

## Endpoints

| Method | Path                        |
|--------|-----------------------------|
| GET    | `/users/me`                 |
| PATCH  | `/users/me`                 |
| GET    | `/users/{id}/memberships`   |
| POST   | `/users/me/avatar`          |

## Notes

- Avatar uploads are multipart — keep them off the JSON envelope path; add a
  dedicated `MultipartApiClient` (or use `client.submitFormWithBinaryData`).
- The `Session.user` already covers most read needs. Only call `/users/me` when
  you need fields that aren't in the session payload.
