package com.angellsan94.dogspotsgo.ui.auth

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.angellsan94.dogspotsgo.domain.repository.AuthRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

sealed interface AuthUiState {
    data object Idle : AuthUiState
    data object Loading : AuthUiState
    data object Success : AuthUiState
    data class Error(val message: String) : AuthUiState
}

@HiltViewModel
class AuthViewModel @Inject constructor(
    private val authRepository: AuthRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow<AuthUiState>(AuthUiState.Idle)
    val uiState: StateFlow<AuthUiState> = _uiState.asStateFlow()

    fun login(email: String, pass: String) {
        if (email.isBlank() || pass.isBlank()) {
            _uiState.update { AuthUiState.Error("Por favor, completa todos los campos.") }
            return
        }
        if (pass.length < 10) {
            _uiState.update { AuthUiState.Error("La contraseña debe tener al menos 10 caracteres.") }
            return
        }

        viewModelScope.launch {
            _uiState.update { AuthUiState.Loading }
            try {
                authRepository.login(email.trim(), pass)
                _uiState.update { AuthUiState.Success }
            } catch (e: Exception) {
                _uiState.update { AuthUiState.Error(e.localizedMessage ?: "Error al iniciar sesión.") }
            }
        }
    }

    fun register(email: String, pass: String, name: String) {
        if (email.isBlank() || pass.isBlank() || name.isBlank()) {
            _uiState.update { AuthUiState.Error("Por favor, completa todos los campos.") }
            return
        }
        if (pass.length < 10) {
            _uiState.update { AuthUiState.Error("La contraseña debe tener al menos 10 caracteres.") }
            return
        }

        viewModelScope.launch {
            _uiState.update { AuthUiState.Loading }
            try {
                authRepository.register(email.trim(), pass, name.trim())
                _uiState.update { AuthUiState.Success }
            } catch (e: Exception) {
                _uiState.update { AuthUiState.Error(e.localizedMessage ?: "Error al registrarse.") }
            }
        }
    }

    fun resetState() {
        _uiState.update { AuthUiState.Idle }
    }
}
