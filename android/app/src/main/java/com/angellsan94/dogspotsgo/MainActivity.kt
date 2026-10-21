package com.angellsan94.dogspotsgo

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.activity.viewModels
import com.angellsan94.dogspotsgo.ui.auth.AuthViewModel
import com.angellsan94.dogspotsgo.ui.navigation.NavGraph
import com.angellsan94.dogspotsgo.ui.theme.DogSpotsGoTheme
import dagger.hilt.android.AndroidEntryPoint

@AndroidEntryPoint
class MainActivity : ComponentActivity() {
    private val authViewModel: AuthViewModel by viewModels()

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableEdgeToEdge()
        setContent {
            DogSpotsGoTheme {
                NavGraph(authViewModel = authViewModel)
            }
        }
    }
}
