package com.angellsan94.dogspotsgo.ui.map

import android.Manifest
import android.content.pm.PackageManager
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.layout.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.core.content.ContextCompat
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.unit.dp
import androidx.compose.ui.viewinterop.AndroidView
import com.angellsan94.dogspotsgo.domain.model.Place
import org.maplibre.android.MapLibre
import org.maplibre.android.camera.CameraPosition
import org.maplibre.android.geometry.LatLng
import org.maplibre.android.maps.MapView
import org.maplibre.android.maps.Style

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun MapScreen(
    viewModel: MapViewModel,
    onLogout: () -> Unit
) {
    val context = LocalContext.current
    val uiState by viewModel.uiState.collectAsState()

    remember {
        MapLibre.getInstance(context)
        true
    }

    var hasLocationPermission by remember {
        mutableStateOf(
            ContextCompat.checkSelfPermission(
                context,
                Manifest.permission.ACCESS_FINE_LOCATION
            ) == PackageManager.PERMISSION_GRANTED
        )
    }

    val permissionLauncher = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.RequestPermission()
    ) { granted ->
        hasLocationPermission = granted
        if (granted) {
            viewModel.fetchUserLocation()
        }
    }

    LaunchedEffect(Unit) {
        if (!hasLocationPermission) {
            permissionLauncher.launch(Manifest.permission.ACCESS_FINE_LOCATION)
        }
    }

    val sheetState = rememberModalBottomSheetState()

    Scaffold(
        floatingActionButton = {
            Button(
                onClick = { viewModel.fetchUserLocation() },
                modifier = Modifier.padding(bottom = if (uiState.selectedPlace != null) 180.dp else 16.dp)
            ) {
                Text("Mi ubicación")
            }
        }
    ) { paddingValues ->
        Box(
            modifier = Modifier
                .fillMaxSize()
                .padding(paddingValues)
        ) {
            AndroidView(
                factory = { ctx ->
                    MapView(ctx).apply {
                        onCreate(null)
                        getMapAsync { map ->
                            map.setStyle("https://tiles.openfreemap.org/styles/liberty") { style ->
                                val initialLat = uiState.userLocation?.latitude ?: 40.4168
                                val initialLng = uiState.userLocation?.longitude ?: -3.7038
                                map.cameraPosition = CameraPosition.Builder()
                                    .target(LatLng(initialLat, initialLng))
                                    .zoom(14.0)
                                    .build()

                                map.addOnCameraIdleListener {
                                    val bounds = map.projection.visibleRegion.latLngBounds
                                    viewModel.loadPlaces(
                                        minLat = bounds.latitudeSouth,
                                        minLng = bounds.longitudeWest,
                                        maxLat = bounds.latitudeNorth,
                                        maxLng = bounds.longitudeEast
                                    )
                                }
                            }
                        }
                    }
                },
                modifier = Modifier.fillMaxSize()
            )

            if (uiState.isLoading) {
                CircularProgressIndicator(
                    modifier = Modifier
                        .align(Alignment.TopCenter)
                        .padding(top = 16.dp)
                )
            }

            if (uiState.selectedPlace != null) {
                val place = uiState.selectedPlace!!
                ModalBottomSheet(
                    onDismissRequest = { viewModel.selectPlace(null) },
                    sheetState = sheetState
                ) {
                    Column(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(24.dp)
                    ) {
                        Text(text = place.title, style = MaterialTheme.typography.titleLarge)
                        Spacer(modifier = Modifier.height(8.dp))
                        Text(text = "Categoría: ${place.category.name}", style = MaterialTheme.typography.bodyMedium)
                        if (place.ratingAvg != null) {
                            Spacer(modifier = Modifier.height(4.dp))
                            Text(text = "Valoración: ${place.ratingAvg} (${place.ratingCount} opiniones)", style = MaterialTheme.typography.bodyMedium)
                        }
                        if (place.distanceM != null) {
                            Spacer(modifier = Modifier.height(4.dp))
                            Text(text = "Distancia: ${place.distanceM} m", style = MaterialTheme.typography.bodyMedium)
                        }
                        Spacer(modifier = Modifier.height(24.dp))
                        Button(
                            onClick = { /* TODO: Detail */ },
                            modifier = Modifier.fillMaxWidth()
                        ) {
                            Text("Ver Detalle")
                        }
                        Spacer(modifier = Modifier.height(24.dp))
                    }
                }
            }
        }
    }
}
