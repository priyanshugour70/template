# frontend-app

Kotlin Multiplatform mobile app for Android **and** iOS, sharing UI via Compose
Multiplatform. Speaks to the Go backend in [`../backend`](../backend) and mirrors
the design system used by [`../frontend-web`](../frontend-web).

## What's in this repo

```
frontend-app/
├── androidApp/           # Android host: Application + Activity + manifest
├── iosApp/               # iOS host: SwiftUI App + ContentView (Xcode project)
├── shared/               # Kotlin Multiplatform library — 95% of the app lives here
│   └── src/
│       ├── commonMain/   # Shared business logic + Compose UI
│       ├── androidMain/  # Platform-specific impls of `expect` declarations
│       ├── iosMain/      # ditto for iOS
│       └── *Test/        # Unit tests per source set
├── gradle/libs.versions.toml
├── settings.gradle.kts
├── build.gradle.kts
└── STRUCTURE.md          # Where every file goes (read this next)
```

The detailed engineering map — what each package is for, where to add new
features, naming conventions — lives in [STRUCTURE.md](./STRUCTURE.md). The
agent-facing playbook is [AGENTS.md](./AGENTS.md).

## Stack

| Concern        | Library                                                |
|----------------|--------------------------------------------------------|
| UI             | Compose Multiplatform 1.11.1                           |
| Architecture   | Clean Architecture (data / domain / presentation) + MVI |
| DI             | Koin 4 (`io.insert-koin:koin-core` + `-compose`)       |
| Networking     | Ktor 3 (OkHttp on Android, Darwin on iOS)              |
| Serialization  | `kotlinx.serialization` JSON                           |
| Async          | `kotlinx.coroutines`                                   |
| Navigation     | Voyager 1.1 (`Navigator` + `TabNavigator` + `ScreenModel`) |
| Persistence    | `multiplatform-settings` (SharedPreferences / NSUserDefaults) |
| Logging        | Napier                                                 |

## Architecture in one paragraph

Every feature is a vertical slice with three sub-packages: `data/` (DTOs, API
classes, repository implementations), `domain/` (models, repository interfaces,
use cases), and `presentation/` (Compose screens + MVI ScreenModels). Cross-cutting
infrastructure (HTTP client, DI graph, theme, base classes) lives in
`shared/src/commonMain/kotlin/com/lssgoo/core/`. Platforms (`androidApp/`,
`iosApp/`) are thin hosts that hand the shared module its environment and render
[`App()`](./shared/src/commonMain/kotlin/com/lssgoo/App.kt).

For the deeper why behind each choice, see [STRUCTURE.md](./STRUCTURE.md#design-principles).

## Running

### Android

1. Open this directory in Android Studio (Giraffe / Koala+).
2. Make sure the backend is up on `localhost:8080` (`cd backend && make run`).
3. The Android emulator reaches the host as `10.0.2.2` — already wired in
   [`MainApplication.kt`](./androidApp/src/main/kotlin/com/lssgoo/MainApplication.kt).
4. Run the **androidApp** configuration, or:

   ```bash
   ./gradlew :androidApp:installDebug
   ```

### iOS

1. Build the shared framework first so Xcode can find it:

   ```bash
   ./gradlew :shared:embedAndSignAppleFrameworkForXcode
   ```

2. Open `iosApp/iosApp.xcodeproj` in Xcode.
3. Pick an iPhone simulator and hit Run.
4. Default backend URL is `http://localhost:8080` (see
   [`ContentView.swift`](./iosApp/iosApp/ContentView.swift) → adjust there or
   wire it through `Info.plist` for production builds).

### Tests

```bash
./gradlew :shared:testAndroidHostTest         # commonMain + androidMain tests via JVM
./gradlew :shared:iosSimulatorArm64Test       # iOS simulator tests
```

## Pointing at a real backend

Edit the `AppConfig(apiBaseUrl = ...)` constructor in:

- **Android** — [`MainApplication.kt`](./androidApp/src/main/kotlin/com/lssgoo/MainApplication.kt)
- **iOS** — pass to `MainViewControllerKt.MainViewController(apiBaseUrl: ...)`
  from [`ContentView.swift`](./iosApp/iosApp/ContentView.swift)

For a real release wire this through gradle build types / Xcode build
configurations. Don't ship a hard-coded URL.

## Adding a feature

Copy [`shared/src/commonMain/kotlin/com/lssgoo/feature/auth/`](./shared/src/commonMain/kotlin/com/lssgoo/feature/auth/)
and rename. Every feature looks the same: `data/`, `domain/`, `presentation/`,
`di/`. Register the new Koin module in [`KoinInitializer.kt`](./shared/src/commonMain/kotlin/com/lssgoo/core/di/KoinInitializer.kt).

The exact recipe lives in the per-feature READMEs — start with
[`feature/dashboard/README.md`](./shared/src/commonMain/kotlin/com/lssgoo/feature/dashboard/README.md).
