# Feature: Dashboard

Mirrors the web's `(dashboard)/dashboard/page.tsx`.

## Layout

```
feature/dashboard/
├── data/                 # DashboardApi, DTOs, DashboardRepositoryImpl
├── domain/
│   ├── model/            # DashboardSummary, KpiCard, ChartSeries
│   ├── DashboardRepository.kt
│   └── usecase/          # GetDashboardSummaryUseCase, RefreshDashboardUseCase
├── presentation/
│   ├── DashboardScreen.kt        # Voyager Screen
│   ├── DashboardScreenModel.kt   # BaseScreenModel<DashboardUiState, DashboardEvent>
│   └── components/               # KpiTile, TrendChart, etc.
├── di/DashboardModule.kt
└── README.md (this file)
```

## Endpoints

| Method | Path                          | Use case |
|--------|-------------------------------|----------|
| GET    | `/dashboard/summary`          | `GetDashboardSummaryUseCase` |
| GET    | `/dashboard/recent-activity`  | `GetRecentActivityUseCase` |

## Pattern to follow

Copy `feature/auth/` and rename. Everything lines up:
1. Add DTOs under `data/dto/` that mirror the backend response.
2. Add an API class wrapping `ApiClient`.
3. Add a `Repository` interface in `domain/` and `RepositoryImpl` in `data/`.
4. Add use cases — one per "verb" the screen calls.
5. Add a `ScreenModel` extending `BaseScreenModel` with state + events.
6. Register everything in `di/DashboardModule.kt`.
