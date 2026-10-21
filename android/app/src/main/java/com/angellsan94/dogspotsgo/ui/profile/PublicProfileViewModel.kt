package com.angellsan94.dogspotsgo.ui.profile

import androidx.lifecycle.SavedStateHandle
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.angellsan94.dogspotsgo.domain.model.PublicUser
import com.angellsan94.dogspotsgo.domain.repository.UserRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

sealed interface PublicProfileUiState {
    data object Loading : PublicProfileUiState
    data class Success(val user: PublicUser) : PublicProfileUiState
    data class Error(val message: String) : PublicProfileUiState
}

@HiltViewModel
class PublicProfileViewModel @Inject constructor(
    private val userRepository: UserRepository,
    savedStateHandle: SavedStateHandle
) : ViewModel() {

    private val userId: String = checkNotNull(savedStateHandle["userId"])

    private val _uiState = MutableStateFlow<PublicProfileUiState>(PublicProfileUiState.Loading)
    val uiState: StateFlow<PublicProfileUiState> = _uiState.asStateFlow()

    init {
        loadPublicProfile()
    }

    fun loadPublicProfile() {
        viewModelScope.launch {
            _uiState.update { PublicProfileUiState.Loading }
            try {
                val user = userRepository.getPublicProfile(userId)
                _uiState.update { PublicProfileUiState.Success(user) }
            } catch (e: Exception) {
                _uiState.update { PublicProfileUiState.Error(e.localizedMessage ?: "Error al cargar el perfil público.") }
            }
        }
    }
}
