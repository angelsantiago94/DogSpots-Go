package com.angellsan94.dogspotsgo.ui.place

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.angellsan94.dogspotsgo.domain.model.Place
import com.angellsan94.dogspotsgo.domain.repository.PlaceRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

sealed interface MyPlacesUiState {
    data object Loading : MyPlacesUiState
    data class Success(val places: List<Place>) : MyPlacesUiState
    data class Error(val message: String) : MyPlacesUiState
}

@HiltViewModel
class MyPlacesViewModel @Inject constructor(
    private val placeRepository: PlaceRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow<MyPlacesUiState>(MyPlacesUiState.Loading)
    val uiState: StateFlow<MyPlacesUiState> = _uiState.asStateFlow()

    init {
        loadMyPlaces()
    }

    fun loadMyPlaces() {
        viewModelScope.launch {
            _uiState.update { MyPlacesUiState.Loading }
            try {
                val places = placeRepository.getMyPlaces()
                _uiState.update { MyPlacesUiState.Success(places) }
            } catch (e: Exception) {
                _uiState.update { MyPlacesUiState.Error(e.localizedMessage ?: "Error al cargar tus lugares.") }
            }
        }
    }

    fun deletePlace(placeId: String) {
        viewModelScope.launch {
            try {
                placeRepository.deletePlace(placeId)
                loadMyPlaces()
            } catch (_: Exception) {
                // Handle deletion error if needed
            }
        }
    }
}
