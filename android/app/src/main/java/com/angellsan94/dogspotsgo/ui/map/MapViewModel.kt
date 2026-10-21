package com.angellsan94.dogspotsgo.ui.map

import android.location.Location
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.angellsan94.dogspotsgo.data.location.LocationManager
import com.angellsan94.dogspotsgo.domain.model.Place
import com.angellsan94.dogspotsgo.domain.repository.PlaceRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

data class MapUiState(
    val places: List<Place> = emptyList(),
    val selectedPlace: Place? = null,
    val userLocation: Location? = null,
    val isLoading: Boolean = false
)

@HiltViewModel
class MapViewModel @Inject constructor(
    private val placeRepository: PlaceRepository,
    private val locationManager: LocationManager
) : ViewModel() {

    private val _uiState = MutableStateFlow(MapUiState())
    val uiState: StateFlow<MapUiState> = _uiState.asStateFlow()

    init {
        fetchUserLocation()
    }

    fun fetchUserLocation() {
        viewModelScope.launch {
            val loc = locationManager.getLastKnownLocation()
            if (loc != null) {
                _uiState.update { it.copy(userLocation = loc) }
                loadPlacesForDefaultBbox(loc.latitude, loc.longitude)
            } else {
                loadPlacesForDefaultBbox(40.4168, -3.7038)
            }
        }
    }

    private fun loadPlacesForDefaultBbox(lat: Double, lng: Double) {
        val delta = 0.1
        loadPlaces(lat - delta, lng - delta, lat + delta, lng + delta)
    }

    fun loadPlaces(minLat: Double, minLng: Double, maxLat: Double, maxLng: Double) {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true) }
            try {
                val places = placeRepository.getPlacesByBbox(minLat, minLng, maxLat, maxLng)
                _uiState.update { it.copy(places = places, isLoading = false) }
            } catch (e: Exception) {
                _uiState.update { it.copy(isLoading = false) }
            }
        }
    }

    fun selectPlace(place: Place?) {
        _uiState.update { it.copy(selectedPlace = place) }
    }
}
