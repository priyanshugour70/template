package com.lssgoo.navigation

import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.asSharedFlow

/**
 * Lightweight global event bus for top-level navigation. Feature screens emit;
 * [com.lssgoo.navigation.RootScreen] listens.
 *
 * Why a global object instead of injecting a navigator?
 *  - Feature `Screen`s shouldn't know they live inside a Voyager Navigator —
 *    keeps them composable in any container (TabNavigator, Preview, tests).
 *  - Cross-feature transitions (e.g. logout from anywhere → land on login) need
 *    a stable channel that survives screen recreation.
 *
 * Used sparingly. Within-feature navigation goes through Voyager's `Navigator`
 * directly via `LocalNavigator.currentOrThrow.push(...)`.
 */
object RootEvents {
    private val _flow = MutableSharedFlow<RootSwitchEvent>(replay = 0, extraBufferCapacity = 8)
    val flow: SharedFlow<RootSwitchEvent> = _flow.asSharedFlow()

    suspend fun emit(event: RootSwitchEvent) {
        _flow.emit(event)
    }
}

sealed interface RootSwitchEvent {
    data object NavigateHome : RootSwitchEvent
    data object NavigateLogin : RootSwitchEvent
}
