package com.angellsan94.dogspotsgo.ui.profile

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.angellsan94.dogspotsgo.domain.model.User
import com.angellsan94.dogspotsgo.domain.repository.AuthRepository
import com.angellsan94.dogspotsgo.domain.repository.UserRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

sealed interface ProfileUiState {
    data object Loading : ProfileUiState
    data class Success(
        val user: User,
        val isEditingName: Boolean = false,
        val accountDeleted: Boolean = false
    ) : ProfileUiState
    data class Error(val message: String) : ProfileUiState
}

@HiltViewModel
class ProfileViewModel @Inject constructor(
    private val userRepository: UserRepository,
    private val authRepository: AuthRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow<ProfileUiState>(ProfileUiState.Loading)
    val uiState: StateFlow<ProfileUiState> = _uiState.asStateFlow()

    init {
        loadProfile()
    }

    fun loadProfile() {
        viewModelScope.launch {
            _uiState.update { ProfileUiState.Loading }
            try {
                val user = userRepository.getMyProfile()
                _uiState.update { ProfileUiState.Success(user = user) }
            } catch (e: Exception) {
                _uiState.update { ProfileUiState.Error(e.localizedMessage ?: "Error al cargar el perfil.") }
            }
        }
    }

    fun updateName(newName: String) {
        if (newName.isBlank()) return
        viewModelScope.launch {
            try {
                val updatedUser = userRepository.updateMyProfile(newName.trim())
                val currentState = _uiState.value
                if (currentState is ProfileUiState.Success) {
                    _uiState.update { currentState.copy(user = updatedUser, isEditingName = false) }
                }
            } catch (e: Exception) {
                _uiState.update { ProfileUiState.Error(e.localizedMessage ?: "Error al actualizar el nombre.") }
            }
        }
    }

    fun setEditingName(editing: Boolean) {
        val currentState = _uiState.value
        if (currentState is ProfileUiState.Success) {
            _uiState.update { currentState.copy(isEditingName = editing) }
        }
    }

    fun logout(onLogoutSuccess: () -> Unit) {
        viewModelScope.launch {
            authRepository.logout()
            onLogoutSuccess()
        }
    }

    fun deleteAccount() {
        viewModelScope.launch {
            try {
                userRepository.deleteAccount()
                val currentState = _uiState.value
                if (currentState is ProfileUiState.Success) {
                    _uiState.update { currentState.copy(accountDeleted = true) }
                }
            } catch (e: Exception) {
                _uiState.update { ProfileUiState.Error(e.localizedMessage ?: "Error al eliminar la cuenta.") }
            }
        }
    }
}
