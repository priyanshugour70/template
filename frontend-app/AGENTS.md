# AGENTS.md — frontend-app

Read this BEFORE writing code in this repo. It gives a coding agent the conventions
that aren't obvious from grepping.

## What this app is

- Kotlin Multiplatform (KMP) targeting Android + iOS via **Compose Multiplatform**.
- UI is shared. Each platform's native module (`androidApp/`, `iosApp/`) is a
  thin host that hands the shared module its environment.
- Backend: Go modular monolith at [`../backend`](../backend), REST `/api/v1/*`
  with `{success, data, error}` envelope. JWT in `Authorization: Bearer`.
- Sibling web app: [`../frontend-web`](../frontend-web) — Next.js 16. Match its
  design tokens (5 palettes already ported) and feature naming where it makes sense.

## Architecture

Clean architecture with vertical-slice features:

```
feature/<name>/
  data/        DTOs, APIs, RepositoryImpl
  domain/      Pure-Kotlin models, Repository interfaces, UseCases
  presentation/ Compose screens + ScreenModels (MVI)
  di/          Koin module
```

Cross-cutting infra is in `core/` (network, storage, DI, theme, base classes).

The detailed map is [STRUCTURE.md](./STRUCTURE.md). Don't skip it.

## Hard rules

1. **Use `ApiClient`, not `HttpClient`.** It handles the envelope and the
   refresh-on-401 retry. Reaching for raw `HttpClient` will silently break auth.

2. **Return `AppResult<T>` from use cases / repositories.** No throwing across
   layer boundaries. Surface errors as structured `AppError` with an `ErrorCode`.

3. **Domain depends on nothing.** No Ktor, no Compose imports in
   `feature/*/domain/`. If you need to add one, you're in the wrong layer.

4. **Register every new class in a Koin module.** And register every new module
   in [`KoinInitializer.kt`](./shared/src/commonMain/kotlin/com/lssgoo/core/di/KoinInitializer.kt).

5. **Mirror the backend.** A new backend module `internal/modules/foo` gets a
   `feature/foo/` on the app side with matching DTOs.

6. **One root `AppTheme { ... }`.** Wraps the Navigator. Never nest.

7. **Hex values in `Palettes.kt` come from the web.** When you tweak a color,
   change `frontend-web/src/theme/palettes.ts` too — they are intentionally
   identical.

8. **`commonMain` is the default.** Only drop into `androidMain` / `iosMain`
   when an `expect`/`actual` is the right tool (platform API, secure storage,
   HTTP engine).

## Naming

- Files use PascalCase: `AuthRepositoryImpl.kt`, `LoginScreen.kt`.
- One public class per file (KMP-iOS-export friendliness).
- Test files end in `Test`: `LoginScreenModelTest.kt`.
- Koin modules: `<feature>Module` as a top-level `val` — e.g. `authModule`,
  `dashboardModule`.

## Patterns you'll see (and should keep using)

### MVI ScreenModel

```kotlin
class XScreenModel(deps...) : BaseScreenModel<XUiState, XEvent>(XUiState()) {
    fun onAction() = launch {
        updateState { copy(isLoading = true) }
        when (val r = useCase(input)) {
            is AppResult.Success -> { updateState { copy(...) }; emitEvent(XEvent....) }
            is AppResult.Failure -> updateState { copy(error = r.error.message) }
        }
    }
}
```

### Voyager Screen

```kotlin
class XScreen : Screen {
    @Composable override fun Content() {
        val model: XScreenModel = koinScreenModel()
        val state by model.state.collectAsState()
        LaunchedEffect(model) {
            model.events.collect { ... }
        }
        ScreenScaffold { ... }
    }
}
```

### Cross-feature navigation

Use [`RootEvents.emit(RootSwitchEvent.NavigateLogin)`](./shared/src/commonMain/kotlin/com/lssgoo/navigation/RootEvents.kt)
for top-level switches (login ↔ main). For within-feature pushes, use Voyager:
`LocalNavigator.currentOrThrow.push(NextScreen())`.

## Running

| Platform | Command                                                       |
|----------|---------------------------------------------------------------|
| Android  | `./gradlew :androidApp:installDebug` (emulator hits `10.0.2.2:8080`) |
| iOS      | `./gradlew :shared:embedAndSignAppleFrameworkForXcode` then open `iosApp/iosApp.xcodeproj` |
| Tests    | `./gradlew :shared:testAndroidHostTest` + `:shared:iosSimulatorArm64Test` |

## Things that will trip you up

- **`com.android.kotlin.multiplatform.library` plugin.** Newer than the classic
  `com.android.library`. Configuration syntax is `androidLibrary { ... }` not
  `android { ... }`.
- **iOS framework name is `Shared`** (capital S). Swift code imports it as
  `import Shared`.
- **The first iOS build needs `embedAndSignAppleFrameworkForXcode`.** Without
  it, Xcode can't find the Kotlin framework.
- **`Voyager 1.1.0-beta03` is the most recent KMP-friendly cut as of this
  scaffold.** If newer is stable, bump in `libs.versions.toml`.

## Backend changes will ripple here

When `backend/internal/modules/<x>/handler.go` adds an endpoint:
1. Update the matching `feature/x/data/*Api.kt` with a new method.
2. Add a DTO if the wire shape doesn't match an existing domain model.
3. Add a use case for the verb.
4. Expose it through the repository interface.

Stay synced with the web team: if they're consuming a new endpoint, the mobile
team probably should be too.
