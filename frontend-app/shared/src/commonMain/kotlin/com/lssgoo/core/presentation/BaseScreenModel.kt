package com.lssgoo.core.presentation

import cafe.adriel.voyager.core.model.ScreenModel
import cafe.adriel.voyager.core.model.screenModelScope
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asSharedFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

/**
 * MVI-style base for every screen.
 *
 *  - [State] is the rendered surface (loading flags, data, error text). One source of truth.
 *  - [Event] is a one-shot side effect (navigate, show snackbar, close keyboard).
 *
 * Voyager's [ScreenModel] survives configuration changes the same way an Android
 * ViewModel does. [screenModelScope] is the right `CoroutineScope` for launching
 * use cases — it's cancelled automatically when the screen leaves composition.
 *
 * Usage:
 * ```
 * class LoginScreenModel(private val login: LoginUseCase) :
 *     BaseScreenModel<LoginState, LoginEvent>(LoginState()) {
 *
 *     fun onSubmit() = launch {
 *         updateState { copy(isLoading = true) }
 *         val result = login(...)
 *         updateState { copy(isLoading = false) }
 *         result.onSuccess { emitEvent(LoginEvent.NavigateHome) }
 *               .onFailure { updateState { copy(error = it.message) } }
 *     }
 * }
 * ```
 */
abstract class BaseScreenModel<State, Event>(initialState: State) : ScreenModel {

    private val _state = MutableStateFlow(initialState)
    val state: StateFlow<State> = _state.asStateFlow()

    // SharedFlow over Channel: Composables can collect with collectAsState side effects.
    // replay=0 ensures stale events don't fire on recomposition.
    private val _events = MutableSharedFlow<Event>(replay = 0, extraBufferCapacity = 16)
    val events: SharedFlow<Event> = _events.asSharedFlow()

    protected val scope: CoroutineScope get() = screenModelScope

    protected fun updateState(reducer: State.() -> State) {
        _state.update(reducer)
    }

    protected suspend fun emitEvent(event: Event) {
        _events.emit(event)
    }

    protected fun launch(block: suspend CoroutineScope.() -> Unit) {
        scope.launch(block = block)
    }
}
