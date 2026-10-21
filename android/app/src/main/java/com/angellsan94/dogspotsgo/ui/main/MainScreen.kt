package com.angellsan94.dogspotsgo.ui.main

import androidx.compose.foundation.layout.*
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp

@Composable
fun MainScreen(onLogout: () -> Unit) {
    Surface(modifier = Modifier.fillMaxSize()) {
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(24.dp),
            verticalArrangement = Arrangement.Center,
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            Text(text = "¡Bienvenido a DogSpots Go!", style = MaterialTheme.typography.headlineMedium)
            Spacer(modifier = Modifier.height(16.dp))
            Text(text = "Aquí se mostrará el mapa de lugares para perros (Fase 8).")
            Spacer(modifier = Modifier.height(32.dp))
            Button(onClick = onLogout) {
                Text("Cerrar Sesión")
            }
        }
    }
}
