package com.angellsan94.dogspotsgo.ui.navigation

import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavType
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import androidx.navigation.navArgument
import com.angellsan94.dogspotsgo.ui.auth.AuthViewModel
import com.angellsan94.dogspotsgo.ui.auth.LoginScreen
import com.angellsan94.dogspotsgo.ui.auth.RegisterScreen
import com.angellsan94.dogspotsgo.ui.main.MainScreen
import com.angellsan94.dogspotsgo.ui.place.*
import com.angellsan94.dogspotsgo.ui.profile.*

sealed class Screen(val route: String) {
    data object Login : Screen("login")
    data object Register : Screen("register")
    data object Main : Screen("main")
    data object PlaceDetail : Screen("place_detail/{placeId}") {
        fun createRoute(placeId: String) = "place_detail/$placeId"
    }
    data object CreatePlace : Screen("create_place?lat={lat}&lng={lng}") {
        fun createRoute() = "create_place"
        fun createRoute(lat: Double, lng: Double) = "create_place?lat=$lat&lng=$lng"
    }
    data object MyPlaces : Screen("my_places")
    data object Profile : Screen("profile")
    data object PublicProfile : Screen("public_profile/{userId}") {
        fun createRoute(userId: String) = "public_profile/$userId"
    }
}

@Composable
fun NavGraph(authViewModel: AuthViewModel) {
    val navController = rememberNavController()

    NavHost(
        navController = navController,
        startDestination = Screen.Login.route
    ) {
        composable(Screen.Login.route) {
            LoginScreen(
                viewModel = authViewModel,
                onLoginSuccess = {
                    navController.navigate(Screen.Main.route) {
                        popUpTo(Screen.Login.route) { inclusive = true }
                    }
                },
                onNavigateToRegister = {
                    navController.navigate(Screen.Register.route)
                }
            )
        }
        composable(Screen.Register.route) {
            RegisterScreen(
                viewModel = authViewModel,
                onRegisterSuccess = {
                    navController.navigate(Screen.Main.route) {
                        popUpTo(Screen.Login.route) { inclusive = true }
                    }
                },
                onNavigateBack = {
                    navController.popBackStack()
                }
            )
        }
        composable(Screen.Main.route) {
            MainScreen(
                onLogout = {
                    navController.navigate(Screen.Login.route) {
                        popUpTo(Screen.Main.route) { inclusive = true }
                    }
                },
                onNavigateToCreate = {
                    navController.navigate(Screen.CreatePlace.createRoute())
                },
                onNavigateToCreateAt = { lat, lng ->
                    navController.navigate(Screen.CreatePlace.createRoute(lat, lng))
                },
                onNavigateToMyPlaces = {
                    navController.navigate(Screen.MyPlaces.route)
                },
                onNavigateToProfile = {
                    navController.navigate(Screen.Profile.route)
                }
            )
        }
        composable(
            route = Screen.PlaceDetail.route,
            arguments = listOf(navArgument("placeId") { type = NavType.StringType })
        ) { entry ->
            val viewModel: PlaceDetailViewModel = hiltViewModel()
            LaunchedEffect(entry) {
                viewModel.loadPlaceDetailAndComments(checkNotNull(entry.arguments?.getString("placeId")))
            }
            PlaceDetailScreen(
                viewModel = viewModel,
                onBack = { navController.popBackStack() }
            )
        }
        composable(
            route = Screen.CreatePlace.route,
            arguments = listOf(
                navArgument("lat") { type = NavType.StringType; nullable = true; defaultValue = null },
                navArgument("lng") { type = NavType.StringType; nullable = true; defaultValue = null }
            )
        ) { entry ->
            val viewModel: CreatePlaceViewModel = hiltViewModel()
            CreatePlaceScreen(
                viewModel = viewModel,
                defaultLat = entry.arguments?.getString("lat")?.toDoubleOrNull() ?: 40.4168,
                defaultLng = entry.arguments?.getString("lng")?.toDoubleOrNull() ?: -3.7038,
                onSuccess = { navController.popBackStack() },
                onBack = { navController.popBackStack() }
            )
        }
        composable(Screen.MyPlaces.route) {
            val viewModel: MyPlacesViewModel = hiltViewModel()
            MyPlacesScreen(
                viewModel = viewModel,
                onPlaceClick = { placeId ->
                    navController.navigate(Screen.PlaceDetail.createRoute(placeId))
                },
                onBack = { navController.popBackStack() }
            )
        }
        composable(Screen.Profile.route) {
            val viewModel: ProfileViewModel = hiltViewModel()
            ProfileScreen(
                viewModel = viewModel,
                onLogoutSuccess = {
                    navController.navigate(Screen.Login.route) {
                        popUpTo(0) { inclusive = true }
                    }
                },
                onBack = { navController.popBackStack() }
            )
        }
        composable(
            route = Screen.PublicProfile.route,
            arguments = listOf(navArgument("userId") { type = NavType.StringType })
        ) {
            val viewModel: PublicProfileViewModel = hiltViewModel()
            PublicProfileScreen(
                viewModel = viewModel,
                onBack = { navController.popBackStack() }
            )
        }
    }
}
