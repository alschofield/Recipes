package com.recipes.mobile.ui

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.padding
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Favorite
import androidx.compose.material.icons.filled.Person
import androidx.compose.material.icons.filled.Search
import androidx.compose.material3.CenterAlignedTopAppBar
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.NavigationBar
import androidx.compose.material3.NavigationBarItem
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import com.recipes.mobile.core.network.AuthSession
import com.recipes.mobile.core.session.SessionStore
import com.recipes.mobile.ui.account.AccountTabScreen
import com.recipes.mobile.ui.favorites.FavoritesTabScreen
import com.recipes.mobile.ui.search.SearchTabScreen

private enum class RootTab(val label: String) {
    Search("Discover"),
    Favorites("Saved"),
    Account("Profile");

    @Composable
    fun icon() {
        when (this) {
            Search -> Icon(Icons.Default.Search, contentDescription = label)
            Favorites -> Icon(Icons.Default.Favorite, contentDescription = label)
            Account -> Icon(Icons.Default.Person, contentDescription = label)
        }
    }
}

@Composable
@OptIn(ExperimentalMaterial3Api::class)
fun RecipesApp() {
    val context = LocalContext.current
    val store = remember { SessionStore(context) }
    var selectedTab by remember { mutableStateOf(RootTab.Search) }
    var session by remember { mutableStateOf<AuthSession?>(store.load()) }

    Scaffold(
        topBar = {
            CenterAlignedTopAppBar(
                title = {
                    Column {
                        Text("Ingrediential", style = MaterialTheme.typography.titleLarge)
                        Text(
                            when (selectedTab) {
                                RootTab.Search -> "Discover recipes from your pantry"
                                RootTab.Favorites -> "Keep saved recipes in sync"
                                RootTab.Account -> "Manage sessions and profile"
                            },
                            style = MaterialTheme.typography.bodySmall
                        )
                    }
                }
            )
        },
        bottomBar = {
            NavigationBar {
                RootTab.entries.forEach { tab ->
                    NavigationBarItem(
                        selected = selectedTab == tab,
                        onClick = { selectedTab = tab },
                        icon = { tab.icon() },
                        label = { Text(tab.label) }
                    )
                }
            }
        }
    ) { padding ->
        Box(modifier = Modifier.padding(padding)) {
            when (selectedTab) {
                RootTab.Search -> SearchTabScreen()
                RootTab.Favorites -> FavoritesTabScreen(session = session)
                RootTab.Account -> {
                    AccountTabScreen(
                        initialSession = session,
                        onSessionChanged = {
                            session = it
                        }
                    )
                }
            }
        }
    }
}
