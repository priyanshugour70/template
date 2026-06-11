import org.jetbrains.kotlin.gradle.dsl.JvmTarget

plugins {
    alias(libs.plugins.kotlinMultiplatform)
    alias(libs.plugins.androidMultiplatformLibrary)
    alias(libs.plugins.composeMultiplatform)
    alias(libs.plugins.composeCompiler)
    alias(libs.plugins.kotlinSerialization)
}

kotlin {
    listOf(
        iosArm64(),
        iosSimulatorArm64()
    ).forEach { iosTarget ->
        iosTarget.binaries.framework {
            baseName = "Shared"
            isStatic = true
            // Voyager + Koin live in shared; they must be exported so iOS code
            // can construct screens directly if it ever needs to.
            export(libs.voyager.navigator)
            export(libs.voyager.screenmodel)
        }
    }

    androidLibrary {
        namespace = "com.lssgoo.shared"
        compileSdk = libs.versions.android.compileSdk.get().toInt()
        minSdk = libs.versions.android.minSdk.get().toInt()

        compilerOptions {
            jvmTarget = JvmTarget.JVM_11
        }
        androidResources {
            enable = true
        }
        withHostTest {
            isIncludeAndroidResources = true
        }
    }

    sourceSets {
        // ---------------- commonMain ----------------
        commonMain.dependencies {
            // Compose UI
            implementation(libs.compose.runtime)
            implementation(libs.compose.foundation)
            implementation(libs.compose.material3)
            implementation(libs.compose.ui)
            implementation(libs.compose.components.resources)
            implementation(libs.compose.uiToolingPreview)
            implementation(libs.compose.materialIconsCore)
            implementation(libs.androidx.lifecycle.viewmodelCompose)
            implementation(libs.androidx.lifecycle.runtimeCompose)

            // Async + serialization
            implementation(libs.kotlinx.coroutines.core)
            implementation(libs.kotlinx.serialization.json)
            implementation(libs.kotlinx.datetime)

            // Network (Ktor)
            api(libs.ktor.client.core)
            implementation(libs.ktor.client.contentNegotiation)
            implementation(libs.ktor.client.logging)
            implementation(libs.ktor.client.auth)
            implementation(libs.ktor.serialization.kotlinxJson)

            // DI (Koin)
            api(libs.koin.core)
            implementation(libs.koin.compose)
            implementation(libs.koin.composeViewmodel)

            // Secure key-value (Keychain on iOS, SharedPrefs on Android)
            implementation(libs.multiplatformSettings.noArg)
            implementation(libs.multiplatformSettings.coroutines)

            // Logging
            implementation(libs.napier)

            // Navigation
            api(libs.voyager.navigator)
            api(libs.voyager.screenmodel)
            implementation(libs.voyager.transitions)
            implementation(libs.voyager.tab)
            implementation(libs.voyager.koin)
        }

        // ---------------- androidMain ----------------
        androidMain.dependencies {
            implementation(libs.compose.uiToolingPreview)
            implementation(libs.kotlinx.coroutines.android)
            implementation(libs.ktor.client.okhttp)
            implementation(libs.koin.android)
        }

        // ---------------- iosMain ----------------
        iosMain.dependencies {
            implementation(libs.ktor.client.darwin)
        }

        // ---------------- commonTest ----------------
        commonTest.dependencies {
            implementation(libs.kotlin.test)
            implementation(libs.kotlinx.coroutines.core)
        }
    }
}

dependencies {
    androidRuntimeClasspath(libs.compose.uiTooling)
}
