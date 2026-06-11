# Project Structure

This is the engineering map. If you're about to write code, find the relevant
section here first — it'll tell you where the file belongs, what to name it,
and which existing files to copy.

## Top level

```
frontend-app/
├── androidApp/                 Android host (Activity + Application + manifest)
├── iosApp/                     iOS host (SwiftUI App + Xcode project)
├── shared/                     Kotlin Multiplatform library
│   └── src/
│       ├── commonMain/         95% of the code: shared business + UI
│       ├── androidMain/        Platform-specific impls
│       ├── iosMain/            Platform-specific impls
│       ├── commonTest/         Pure-Kotlin tests
│       ├── androidHostTest/    JVM-runnable tests for Android-side
│       └── iosTest/            Simulator-runnable tests
├── gradle/libs.versions.toml   Version catalog (single source of truth for deps)
└── settings.gradle.kts         Module list
```

## The shared module

Inside `shared/src/commonMain/kotlin/com/lssgoo/`:

```
com.lssgoo/
├── App.kt                      Root composable: Navigator(RootScreen())
├── Platform.kt                 expect-fun for runtime platform name
│
├── core/                       Cross-cutting infrastructure
│   ├── config/AppConfig.kt     Static config handed in at startup
│   ├── result/                 AppResult sealed type, AppError, ErrorCode
│   ├── network/                Ktor client + auth refresh + envelope parsing
│   │   ├── HttpClientFactory.kt
│   │   ├── ApiClient.kt        Use this from every repository, NEVER raw HttpClient
│   │   ├── ApiResponse.kt      Wire envelope { success, data, error, pagination }
│   │   └── TokenRefresher.kt
│   ├── storage/                Persistent state (Settings-backed)
│   │   ├── TokenStorage.kt
│   │   └── SessionCache.kt
│   ├── di/                     Koin wiring
│   │   ├── KoinInitializer.kt  initApp(config) — single entry point
│   │   ├── CoreModule.kt       Core deps (HTTP, storage, theme)
│   │   └── PlatformDependencies.kt  expect-fun for Settings + HttpClientEngine
│   ├── domain/UseCase.kt       UseCase<P, T> + NoParamUseCase<T> marker interfaces
│   └── presentation/
│       ├── BaseScreenModel.kt  MVI base: state, events, screenModelScope
│       ├── theme/              5 palettes mirroring the web's design tokens
│       │   ├── AppTheme.kt
│       │   ├── Palette.kt
│       │   ├── Palettes.kt     ForestTrail, SunsetHorizon, Tropical, Mountain, Desert
│       │   ├── PaletteController.kt
│       │   ├── Shapes.kt
│       │   ├── Spacing.kt      LocalSpacing — xs/sm/md/lg/xl/xxl/xxxl
│       │   └── Typography.kt
│       └── components/         Reusable primitives (PrimaryButton, EmptyState, etc.)
│
├── feature/                    Vertical slices, one folder per feature
│   ├── auth/                   END-TO-END EXAMPLE — copy this when adding a feature
│   │   ├── data/
│   │   │   ├── dto/AuthDtos.kt
│   │   │   ├── AuthApi.kt
│   │   │   └── AuthRepositoryImpl.kt
│   │   ├── domain/
│   │   │   ├── model/{Session, AuthTenant}.kt
│   │   │   ├── AuthRepository.kt           (interface)
│   │   │   └── usecase/{DiscoverTenants, Login, Logout, BootstrapSession}UseCase.kt
│   │   ├── presentation/
│   │   │   ├── LoginScreen.kt              (Voyager Screen)
│   │   │   └── LoginScreenModel.kt         (BaseScreenModel)
│   │   └── di/AuthModule.kt
│   │
│   ├── dashboard/              Scaffold — has DashboardScreen + DI stub + README
│   ├── billing/                Stub only — see feature/billing/README.md
│   ├── user/                   Stub only — see feature/user/README.md
│   └── settings/               Working palette + dark-mode toggles
│
└── navigation/
    ├── RootEvents.kt           Cross-feature event bus (login success → home, etc.)
    ├── RootScreen.kt           Booting → Auth vs Main based on session state
    └── MainScaffoldScreen.kt   Post-login bottom tabs (Home/Billing/Profile/Settings)
```

## Adding a new feature — step by step

Using "products" as an example:

1. **Make the package**:
   `shared/src/commonMain/kotlin/com/lssgoo/feature/products/`

2. **Domain layer** (`domain/`):
   - `model/Product.kt` — `@Serializable data class Product(val id: ...)`
   - `ProductsRepository.kt` — interface only.
   - `usecase/GetProductsUseCase.kt` — implements `UseCase<Unit, List<Product>>`.

3. **Data layer** (`data/`):
   - `dto/ProductDtos.kt` — wire types if they differ from domain.
   - `ProductsApi.kt` — `client.get<ProductsResponseDto>("/products")`.
   - `ProductsRepositoryImpl.kt` — implements the domain interface, maps DTO→domain.

4. **Presentation layer** (`presentation/`):
   - `ProductsScreenModel.kt` — `class ProductsScreenModel(...) : BaseScreenModel<ProductsUiState, ProductsEvent>(ProductsUiState())`.
   - `ProductsScreen.kt` — `class ProductsScreen : Screen { @Composable override fun Content() { ... } }`.

5. **DI** (`di/ProductsModule.kt`):
   ```kotlin
   val productsModule = module {
       single { ProductsApi(client = get()) }
       single<ProductsRepository> { ProductsRepositoryImpl(api = get()) }
       factoryOf(::GetProductsUseCase)
       factoryOf(::ProductsScreenModel)
   }
   ```

6. **Register** in `core/di/KoinInitializer.kt`:
   ```kotlin
   modules(coreModule(config), authModule, ..., productsModule)
   ```

7. **Mount** somewhere — either add a `ProductsTab` to `MainScaffoldScreen.kt`,
   or push from another screen with `navigator.push(ProductsScreen())`.

## Design principles

These are the rules the codebase enforces. Break them only with a documented reason.

1. **Vertical slices over horizontal layers.** Don't create top-level `data/`,
   `domain/`, `presentation/` folders that span all features. Each `feature/X/`
   stays self-contained.

2. **Domain depends on nothing.** No `import io.ktor`, no `import androidx.compose`
   inside `feature/*/domain/`. Use cases and repositories work in plain Kotlin so
   they're trivially testable on JVM.

3. **One ApiClient.** Repositories inject `ApiClient`, not `HttpClient`. This
   guarantees the auth-refresh logic runs uniformly.

4. **Use cases return `AppResult<T>`.** Never throw across the use-case boundary.
   ScreenModels do `when (val r = useCase(...)) { is Success -> ...; is Failure -> ... }`.

5. **ScreenModels emit events.** Side effects (navigation, snackbars) go through
   the `events: SharedFlow<Event>` channel. The state flow is for rendering.

6. **One root composition for the theme.** `AppTheme { ... }` wraps the Navigator
   exactly once. Nested invocations are a code smell.

7. **No DTOs in the domain layer.** Repository implementations map DTOs to
   domain models on the way out. The wire shape can change without forcing
   the rest of the app to change.

8. **Platform glue stays in `expect`/`actual`.** When you need
   `EncryptedSharedPreferences` on Android and `Keychain` on iOS, declare an
   `expect` in `commonMain/core/di/` and fulfill it in `androidMain` and
   `iosMain`.

9. **Mirror the backend's vertical slices.** A backend module
   `internal/modules/X` should have a corresponding `feature/x/` in the app.
   Same names → easier mental model.

10. **Mirror the web's design tokens.** The 5 palettes in `Palettes.kt` are
    copied byte-for-byte from `frontend-web/src/theme/palettes.ts`. When one
    side rotates a color, mirror it.
