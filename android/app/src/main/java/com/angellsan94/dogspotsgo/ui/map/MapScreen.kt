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
import com.angellsan94.dogspotsgo.ui.place.PlaceDetailScreen
import com.angellsan94.dogspotsgo.ui.place.PlaceDetailViewModel
import org.maplibre.android.MapLibre
import org.maplibre.android.camera.CameraPosition
import org.maplibre.android.camera.CameraUpdateFactory
import org.maplibre.android.geometry.LatLng
import org.maplibre.android.annotations.MarkerOptions
import org.maplibre.android.maps.MapLibreMap
import org.maplibre.android.maps.MapView

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun MapScreen(
    viewModel: MapViewModel,
    placeDetailViewModel: PlaceDetailViewModel,
    onLogout: () -> Unit,
    onNavigateToCreate: () -> Unit,
    onNavigateToCreateAt: (Double, Double) -> Unit,
    onNavigateToMyPlaces: () -> Unit,
    onNavigateToProfile: () -> Unit
) {
    val context = LocalContext.current
    val uiState by viewModel.uiState.collectAsState()
    var mapLibreMap by remember { mutableStateOf<MapLibreMap?>(null) }
    var createPlaceCoordinates by remember { mutableStateOf<LatLng?>(null) }

    remember {
        MapLibre.getInstance(context)
        true
    }

    var hasLocationPermission by remember {
        mutableStateOf(
            ContextCompat.checkSelfPermission(context, Manifest.permission.ACCESS_FINE_LOCATION) == PackageManager.PERMISSION_GRANTED ||
                ContextCompat.checkSelfPermission(context, Manifest.permission.ACCESS_COARSE_LOCATION) == PackageManager.PERMISSION_GRANTED
        )
    }

    val permissionLauncher = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.RequestMultiplePermissions()
    ) { granted ->
        hasLocationPermission = granted.values.any { it }
        if (hasLocationPermission) {
            viewModel.fetchUserLocation()
        }
    }

    LaunchedEffect(Unit) {
        if (hasLocationPermission) {
            viewModel.fetchUserLocation()
        } else {
            permissionLauncher.launch(
                arrayOf(Manifest.permission.ACCESS_FINE_LOCATION, Manifest.permission.ACCESS_COARSE_LOCATION)
            )
        }
    }

    LaunchedEffect(mapLibreMap, uiState.userLocation) {
        val map = mapLibreMap ?: return@LaunchedEffect
        val location = uiState.userLocation ?: return@LaunchedEffect
        map.animateCamera(
            CameraUpdateFactory.newLatLngZoom(
                LatLng(location.latitude, location.longitude),
                14.0
            )
        )
    }

    val sheetState = rememberModalBottomSheetState()

    LaunchedEffect(mapLibreMap, uiState.places) {
        val map = mapLibreMap ?: return@LaunchedEffect
        map.clear()
        uiState.places.forEach { place ->
            map.addMarker(
                MarkerOptions()
                    .position(LatLng(place.lat, place.lng))
                    .title(place.title)
                    .snippet(place.id)
            )
        }
        map.setOnMarkerClickListener { marker ->
            uiState.places.firstOrNull { it.id == marker.snippet }?.let { place ->
                viewModel.selectPlace(place)
                placeDetailViewModel.loadPlaceDetailAndComments(place.id)
            }
            true
        }
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("DogSpots Go") },
                actions = {
                    TextButton(onClick = onNavigateToCreate) {
                        Text("+ Crear")
                    }
                    TextButton(onClick = onNavigateToMyPlaces) {
                        Text("Mis Lugares")
                    }
                    TextButton(onClick = onNavigateToProfile) {
                        Text("Perfil")
                    }
                    TextButton(onClick = onLogout) {
                        Text("Salir")
                    }
                }
            )
        },
        floatingActionButton = {
            Button(
                onClick = {
                    if (hasLocationPermission) {
                        viewModel.fetchUserLocation()
                    } else {
                        permissionLauncher.launch(
                            arrayOf(Manifest.permission.ACCESS_FINE_LOCATION, Manifest.permission.ACCESS_COARSE_LOCATION)
                        )
                    }
                },
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
                            map.setStyle("https://tiles.openfreemap.org/styles/liberty") { _ ->
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
                                map.addOnMapLongClickListener { coordinates ->
                                    createPlaceCoordinates = coordinates
                                    true
                                }
                                mapLibreMap = map
                            }
                        }
                    }
                },
                modifier = Modifier.fillMaxSize()
            )

            createPlaceCoordinates?.let { coordinates ->
                AlertDialog(
                    onDismissRequest = { createPlaceCoordinates = null },
                    title = { Text("Crear lugar aquí") },
                    text = {
                        Text("¿Quieres crear un lugar en ${coordinates.latitude}, ${coordinates.longitude}?")
                    },
                    confirmButton = {
                        TextButton(
                            onClick = {
                                onNavigateToCreateAt(coordinates.latitude, coordinates.longitude)
                                createPlaceCoordinates = null
                            }
                        ) {
                            Text("Crear lugar")
                        }
                    },
                    dismissButton = {
                        TextButton(onClick = { createPlaceCoordinates = null }) {
                            Text("Cancelar")
                        }
                    }
                )
            }

            if (uiState.isLoading) {
                CircularProgressIndicator(
                    modifier = Modifier
                        .align(Alignment.TopCenter)
                        .padding(top = 16.dp)
                )
            }

            if (uiState.selectedPlace != null) {
                ModalBottomSheet(
                    onDismissRequest = { viewModel.selectPlace(null) },
                    sheetState = sheetState
                ) {
                    PlaceDetailScreen(
                        viewModel = placeDetailViewModel,
                        onBack = { viewModel.selectPlace(null) }
                    )
                }
            }
        }
    }
}
