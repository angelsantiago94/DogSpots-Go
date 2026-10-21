package com.angellsan94.dogspotsgo.ui.main

import androidx.compose.runtime.Composable
import androidx.hilt.navigation.compose.hiltViewModel
import com.angellsan94.dogspotsgo.ui.map.MapScreen
import com.angellsan94.dogspotsgo.ui.map.MapViewModel

@Composable
fun MainScreen(
    onLogout: () -> Unit,
    onNavigateToDetail: (String) -> Unit,
    onNavigateToCreate: () -> Unit,
    onNavigateToMyPlaces: () -> Unit,
    onNavigateToProfile: () -> Unit
) {
    val mapViewModel: MapViewModel = hiltViewModel()
    MapScreen(
        viewModel = mapViewModel,
        onLogout = onLogout,
        onNavigateToDetail = onNavigateToDetail,
        onNavigateToCreate = onNavigateToCreate,
        onNavigateToMyPlaces = onNavigateToMyPlaces,
        onNavigateToProfile = onNavigateToProfile
    )
}
