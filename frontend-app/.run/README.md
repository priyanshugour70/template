# Run configurations

These show up in Android Studio's run-config dropdown (top toolbar) after the
next Gradle sync. Each one is a `.run.xml` that AS reads automatically.

| Config | What it does | When to use |
|--------|--------------|-------------|
| **androidApp** | Builds + installs the debug APK, launches `MainActivity` on the selected device. | Daily dev loop on Android. |
| **Assemble Android Debug** | Runs `:androidApp:assembleDebug` only. | CI-style check that nothing is broken without needing a device attached. |
| **Shared Common Tests** | Runs `:shared:testAndroidHostTest` (JVM unit tests covering commonMain + androidMain). | Re-run after touching domain / data layers. |
| **iosApp** | Builds the Xcode `iosApp` scheme and runs on the selected iOS simulator. **Requires the Kotlin Multiplatform plugin for Android Studio.** | Daily dev loop on iOS without leaving AS. |
| **Embed iOS Framework** | Runs `:shared:embedAndSignAppleFrameworkForXcode`. | One-time before opening Xcode, or whenever you change `shared/` and want to refresh the `Shared.framework` Xcode uses. |

## If "androidApp" shows a module-name error

The `<module name="frontend-app.androidApp.main" />` line assumes Gradle
project name `frontend-app`. If Android Studio names it differently, just:

1. Open Run → Edit Configurations → androidApp
2. Pick the right Module from the dropdown
3. Save — AS rewrites this file with the correct name.

## iOS run config requires the KMP plugin

If the **iosApp** run config doesn't show up, install:
**Settings → Plugins → Marketplace → "Kotlin Multiplatform"** (by JetBrains).

Without it you can still build/run iOS from Xcode directly:

```bash
./gradlew :shared:embedAndSignAppleFrameworkForXcode
open iosApp/iosApp.xcodeproj
```
