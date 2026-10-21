package com.angellsan94.dogspotsgo.ui.place

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.angellsan94.dogspotsgo.domain.model.Tag
import com.angellsan94.dogspotsgo.domain.repository.PlaceRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

sealed interface CreatePlaceUiState {
    data object Idle : CreatePlaceUiState
    data object Loading : CreatePlaceUiState
    data object Success : CreatePlaceUiState
    data class Error(val message: String) : CreatePlaceUiState
}

@HiltViewModel
class CreatePlaceViewModel @Inject constructor(
    private val placeRepository: PlaceRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow<CreatePlaceUiState>(CreatePlaceUiState.Idle)
    val uiState: StateFlow<CreatePlaceUiState> = _uiState.asStateFlow()

    private val _tags = MutableStateFlow<List<Tag>>(emptyList())
    val tags: StateFlow<List<Tag>> = _tags.asStateFlow()

    init {
        loadTags()
    }

    fun loadTags() {
        viewModelScope.launch {
            try {
                val fetchedTags = placeRepository.getTags()
                _tags.update { fetchedTags }
            } catch (_: Exception) {
                // Ignore tag load failure
            }
        }
    }

    fun createPlace(
        title: String,
        description: String?,
        lat: Double,
        lng: Double,
        category: String,
        selectedTags: List<String>
    ) {
        if (title.isBlank()) {
            _uiState.update { CreatePlaceUiState.Error("El título es obligatorio.") }
            return
        }

        viewModelScope.launch {
            _uiState.update { CreatePlaceUiState.Loading }
            try {
                placeRepository.createPlace(
                    title = title.trim(),
                    description = description?.trim().takeIf { !it.isNullOrBlank() },
                    lat = lat,
                    lng = lng,
                    category = category,
                    tags = selectedTags
                )
                _uiState.update { CreatePlaceUiState.Success }
            } catch (e: Exception) {
                _uiState.update { CreatePlaceUiState.Error(e.localizedMessage ?: "Error al crear el lugar.") }
            }
        }
    }

    fun resetState() {
        _uiState.update { CreatePlaceUiState.Idle }
    }
}
