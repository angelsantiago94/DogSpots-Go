package com.angellsan94.dogspotsgo.ui.profile

import androidx.compose.foundation.layout.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ProfileScreen(
    viewModel: ProfileViewModel,
    onLogoutSuccess: () -> Unit,
    onBack: () -> Unit
) {
    val uiState by viewModel.uiState.collectAsState()
    var nameInput by remember { mutableStateOf("") }
    var showDeleteDialog by remember { mutableStateOf(false) }

    LaunchedEffect(uiState) {
        if (uiState is ProfileUiState.Success && (uiState as ProfileUiState.Success).accountDeleted) {
            onLogoutSuccess()
        }
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Mi Perfil") },
                navigationIcon = {
                    TextButton(onClick = onBack) {
                        Text("< Volver")
                    }
                }
            )
        }
    ) { paddingValues ->
        Box(
            modifier = Modifier
                .fillMaxSize()
                .padding(paddingValues),
            contentAlignment = Alignment.Center
        ) {
            when (val state = uiState) {
                is ProfileUiState.Loading -> {
                    CircularProgressIndicator()
                }
                is ProfileUiState.Error -> {
                    Column(horizontalAlignment = Alignment.CenterHorizontally) {
                        Text(text = state.message, color = MaterialTheme.colorScheme.error)
                        Spacer(modifier = Modifier.height(16.dp))
                        Button(onClick = { viewModel.loadProfile() }) {
                            Text("Reintentar")
                        }
                    }
                }
                is ProfileUiState.Success -> {
                    val user = state.user
                    Column(
                        modifier = Modifier
                            .fillMaxSize()
                            .padding(24.dp),
                        verticalArrangement = Arrangement.Top
                    ) {
                        Text(text = "Correo:", style = MaterialTheme.typography.labelMedium)
                        Text(text = user.email, style = MaterialTheme.typography.bodyLarge)
                        Spacer(modifier = Modifier.height(16.dp))

                        Text(text = "Nombre:", style = MaterialTheme.typography.labelMedium)
                        if (state.isEditingName) {
                            OutlinedTextField(
                                value = nameInput,
                                onValueChange = { nameInput = it },
                                label = { Text("Nuevo nombre") },
                                modifier = Modifier.fillMaxWidth()
                            )
                            Spacer(modifier = Modifier.height(8.dp))
                            Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                                Button(onClick = { viewModel.updateName(nameInput) }) {
                                    Text("Guardar")
                                }
                                TextButton(onClick = { viewModel.setEditingName(false) }) {
                                    Text("Cancelar")
                                }
                            }
                        } else {
                            Row(
                                modifier = Modifier.fillMaxWidth(),
                                horizontalArrangement = Arrangement.SpaceBetween,
                                verticalAlignment = Alignment.CenterVertically
                            ) {
                                Text(text = user.name, style = MaterialTheme.typography.titleMedium)
                                TextButton(onClick = {
                                    nameInput = user.name
                                    viewModel.setEditingName(true)
                                }) {
                                    Text("Editar")
                                }
                            }
                        }

                        Spacer(modifier = Modifier.height(16.dp))
                        Text(text = "Miembro desde: ${user.createdAt.take(10)}", style = MaterialTheme.typography.bodySmall)

                        Spacer(modifier = Modifier.height(48.dp))
                        Button(
                            onClick = { viewModel.logout(onLogoutSuccess) },
                            modifier = Modifier.fillMaxWidth()
                        ) {
                            Text("Cerrar Sesión")
                        }

                        Spacer(modifier = Modifier.height(24.dp))
                        OutlinedButton(
                            onClick = { showDeleteDialog = true },
                            colors = ButtonDefaults.outlinedButtonColors(contentColor = MaterialTheme.colorScheme.error),
                            modifier = Modifier.fillMaxWidth()
                        ) {
                            Text("Eliminar Cuenta")
                        }
                    }
                }
            }

            if (showDeleteDialog) {
                AlertDialog(
                    onDismissRequest = { showDeleteDialog = false },
                    title = { Text("¿Eliminar cuenta?") },
                    text = { Text("Esta acción eliminará permanentemente tu cuenta y todos tus datos asociados. No se puede deshacer.") },
                    confirmButton = {
                        TextButton(
                            onClick = {
                                showDeleteDialog = false
                                viewModel.deleteAccount()
                            }
                        ) {
                            Text("Eliminar", color = MaterialTheme.colorScheme.error)
                        }
                    },
                    dismissButton = {
                        TextButton(onClick = { showDeleteDialog = false }) {
                            Text("Cancelar")
                        }
                    }
                )
            }
        }
    }
}
