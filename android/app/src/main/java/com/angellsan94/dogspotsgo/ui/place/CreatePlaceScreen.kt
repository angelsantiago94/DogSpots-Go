package com.angellsan94.dogspotsgo.ui.place

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun CreatePlaceScreen(
    viewModel: CreatePlaceViewModel,
    defaultLat: Double = 40.4168,
    defaultLng: Double = -3.7038,
    onSuccess: () -> Unit,
    onBack: () -> Unit
) {
    val uiState by viewModel.uiState.collectAsState()
    val availableTags by viewModel.tags.collectAsState()

    var title by remember { mutableStateOf("") }
    var description by remember { mutableStateOf("") }
    var selectedCategoryKey by remember { mutableStateOf("park") }
    var dropdownExpanded by remember { mutableStateOf(false) }
    var latText by remember { mutableStateOf(defaultLat.toString()) }
    var lngText by remember { mutableStateOf(defaultLng.toString()) }
    val selectedTags = remember { mutableStateListOf<String>() }

    val categoryOptions = listOf(
        "park" to "Parque",
        "dog_park" to "Área Canina / Pipican",
        "beach" to "Playa",
        "trail" to "Ruta Natural",
        "cafe" to "Cafetería",
        "restaurant" to "Bar / Restaurante",
        "other" to "Otro"
    )

    LaunchedEffect(uiState) {
        if (uiState is CreatePlaceUiState.Success) {
            onSuccess()
            viewModel.resetState()
        }
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Crear Nuevo Lugar") },
                navigationIcon = {
                    TextButton(onClick = onBack) {
                        Text("< Volver")
                    }
                }
            )
        }
    ) { paddingValues ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(paddingValues)
                .verticalScroll(rememberScrollState())
                .padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(16.dp)
        ) {
            OutlinedTextField(
                value = title,
                onValueChange = { title = it },
                label = { Text("Título del lugar *") },
                modifier = Modifier.fillMaxWidth(),
                singleLine = true
            )

            OutlinedTextField(
                value = description,
                onValueChange = { description = it },
                label = { Text("Descripción") },
                modifier = Modifier.fillMaxWidth()
            )

            ExposedDropdownMenuBox(
                expanded = dropdownExpanded,
                onExpandedChange = { dropdownExpanded = !dropdownExpanded },
                modifier = Modifier.fillMaxWidth()
            ) {
                OutlinedTextField(
                    value = categoryOptions.find { it.first == selectedCategoryKey }?.second ?: "Parque",
                    onValueChange = {},
                    readOnly = true,
                    label = { Text("Categoría *") },
                    trailingIcon = { ExposedDropdownMenuDefaults.TrailingIcon(expanded = dropdownExpanded) },
                    modifier = Modifier.menuAnchor().fillMaxWidth()
                )
                ExposedDropdownMenu(
                    expanded = dropdownExpanded,
                    onDismissRequest = { dropdownExpanded = false }
                ) {
                    categoryOptions.forEach { (key, label) ->
                        DropdownMenuItem(
                            text = { Text(label) },
                            onClick = {
                                selectedCategoryKey = key
                                dropdownExpanded = false
                            }
                        )
                    }
                }
            }

            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                OutlinedTextField(
                    value = latText,
                    onValueChange = { latText = it },
                    label = { Text("Latitud") },
                    modifier = Modifier.weight(1f),
                    singleLine = true
                )
                OutlinedTextField(
                    value = lngText,
                    onValueChange = { lngText = it },
                    label = { Text("Longitud") },
                    modifier = Modifier.weight(1f),
                    singleLine = true
                )
            }

            if (availableTags.isNotEmpty()) {
                Text(text = "Etiquetas", style = MaterialTheme.typography.titleMedium)
                Column(verticalArrangement = Arrangement.spacedBy(4.dp)) {
                    availableTags.forEach { tag ->
                        val isSelected = selectedTags.contains(tag.slug)
                        Row(
                            verticalAlignment = Alignment.CenterVertically,
                            modifier = Modifier.fillMaxWidth()
                        ) {
                            Checkbox(
                                checked = isSelected,
                                onCheckedChange = { checked ->
                                    if (checked) selectedTags.add(tag.slug)
                                    else selectedTags.remove(tag.slug)
                                }
                            )
                            Spacer(modifier = Modifier.width(8.dp))
                            Text(text = tag.label)
                        }
                    }
                }
            }

            Spacer(modifier = Modifier.height(16.dp))

            when (uiState) {
                is CreatePlaceUiState.Loading -> {
                    Box(modifier = Modifier.fillMaxWidth(), contentAlignment = Alignment.Center) {
                        CircularProgressIndicator()
                    }
                }
                else -> {
                    Button(
                        onClick = {
                            val lat = latText.toDoubleOrNull() ?: defaultLat
                            val lng = lngText.toDoubleOrNull() ?: defaultLng
                            viewModel.createPlace(
                                title = title,
                                description = description,
                                lat = lat,
                                lng = lng,
                                category = selectedCategoryKey,
                                selectedTags = selectedTags.toList()
                            )
                        },
                        modifier = Modifier.fillMaxWidth()
                    ) {
                        Text("Guardar Lugar")
                    }
                }
            }

            if (uiState is CreatePlaceUiState.Error) {
                Text(
                    text = (uiState as CreatePlaceUiState.Error).message,
                    color = MaterialTheme.colorScheme.error
                )
            }
        }
    }
}
