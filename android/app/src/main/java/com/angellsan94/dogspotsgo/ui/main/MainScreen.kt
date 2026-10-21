package com.angellsan94.dogspotsgo.ui.main

import androidx.compose.runtime.Composable
import androidx.hilt.navigation.compose.hiltViewModel
import com.angellsan94.dogspotsgo.ui.map.MapScreen
import com.angellsan94.dogspotsgo.ui.map.MapViewModel
import com.angellsan94.dogspotsgo.ui.place.PlaceDetailViewModel

@Composable
fun MainScreen(
    onLogout: () -> Unit,
    onNavigateToCreate: () -> Unit,
    onNavigateToCreateAt: (Double, Double) -> Unit,
    onNavigateToMyPlaces: () -> Unit,
    onNavigateToProfile: () -> Unit
) {
    val mapViewModel: MapViewModel = hiltViewModel()
    val placeDetailViewModel: PlaceDetailViewModel = hiltViewModel()
    MapScreen(
        viewModel = mapViewModel,
        placeDetailViewModel = placeDetailViewModel,
        onLogout = onLogout,
        onNavigateToCreate = onNavigateToCreate,
        onNavigateToCreateAt = onNavigateToCreateAt,
        onNavigateToMyPlaces = onNavigateToMyPlaces,
        onNavigateToProfile = onNavigateToProfile
    )
}
