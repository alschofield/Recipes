package com.recipes.mobile.ui.favorites

import android.content.Context
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.Button
import androidx.compose.material3.Card
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalView
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.semantics.heading
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.input.PasswordVisualTransformation
import androidx.compose.ui.unit.dp
import com.recipes.mobile.BuildConfig
import com.recipes.mobile.core.network.AuthSession
import com.recipes.mobile.core.network.FavoriteItem
import com.recipes.mobile.core.network.FavoriteQueueAction
import com.recipes.mobile.core.network.RecipesApiClient
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import org.json.JSONArray
import org.json.JSONObject

private const val PREFS = "recipes_mobile_favorites"
private const val KEY_QUEUE = "queue"

@Composable
fun FavoritesTabScreen(session: AuthSession?) {
    val context = LocalContext.current
    val view = LocalView.current
    val scope = rememberCoroutineScope()
    val client = remember { RecipesApiClient(BuildConfig.API_BASE_URL) }

    var userId by remember(session?.userId) { mutableStateOf(session?.userId ?: "") }
    var token by remember(session?.token) { mutableStateOf(session?.token ?: "") }
    var recipeId by remember { mutableStateOf("") }
    var loading by remember { mutableStateOf(false) }
    var error by remember { mutableStateOf<String?>(null) }
    var notice by remember { mutableStateOf<String?>(null) }
    var favorites by remember { mutableStateOf<List<FavoriteItem>>(emptyList()) }
    var queue by remember { mutableStateOf(loadQueue(context)) }

    LaunchedEffect(notice) {
        if (!notice.isNullOrBlank()) {
            view.announceForAccessibility(notice)
        }
    }

    LaunchedEffect(error) {
        if (!error.isNullOrBlank()) {
            view.announceForAccessibility("Favorites error. $error")
        }
    }

    suspend fun refreshFavorites() {
        if (userId.isBlank() || token.isBlank()) return
        favorites = withContext(Dispatchers.IO) { client.listFavorites(userId.trim(), token.trim()) }
    }

    fun enqueue(op: String, id: String) {
        val updated = compactQueue(queue + FavoriteQueueAction(op = op, recipeId = id, queuedAtEpochMs = System.currentTimeMillis()))
        queue = updated
        saveQueue(context, updated)
        notice = "Saved action offline. Sync queue to reconcile."
    }

    suspend fun syncQueue() {
        if (userId.isBlank() || token.isBlank() || queue.isEmpty()) return
        val remaining = mutableListOf<FavoriteQueueAction>()
        var replayed = 0
        withContext(Dispatchers.IO) {
            queue.forEach { action ->
                try {
                    if (action.op == "add") {
                        client.addFavorite(userId.trim(), action.recipeId, token.trim())
                    } else {
                        client.removeFavorite(userId.trim(), action.recipeId, token.trim())
                    }
                    replayed += 1
                } catch (_: Exception) {
                    remaining += action
                }
            }
        }
        val compacted = compactQueue(remaining)
        queue = compacted
        saveQueue(context, compacted)
        refreshFavorites()
        notice = "Replayed $replayed action(s). Pending: ${compacted.size}."
    }

    LazyColumn(modifier = Modifier.fillMaxWidth().padding(16.dp), verticalArrangement = Arrangement.spacedBy(12.dp)) {
        item { Text("Saved Recipes", modifier = Modifier.semantics { heading() }) }
        item {
            if (session != null) {
                Text("Signed in as ${session.username} (${session.userId})")
            } else {
                Text("Sign in from Profile to use secure session")
            }
        }
        item {
            OutlinedTextField(
                modifier = Modifier
                    .fillMaxWidth()
                    .semantics { contentDescription = "User ID input" },
                value = userId,
                onValueChange = { userId = it },
                label = { Text("User ID") }
            )
        }
        item {
            OutlinedTextField(
                modifier = Modifier
                    .fillMaxWidth()
                    .semantics { contentDescription = "Access token input" },
                value = token,
                onValueChange = { token = it },
                label = { Text("Access token") },
                visualTransformation = PasswordVisualTransformation()
            )
        }
        if (session != null) {
            item {
                Button(onClick = {
                    userId = session.userId
                    token = session.token
                    notice = "Using latest saved session token"
                }, modifier = Modifier.semantics { contentDescription = "Use saved session credentials" }) {
                    Text("Use saved session")
                }
            }
        }
        item {
            Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                Button(onClick = {
                    scope.launch {
                        loading = true
                        error = null
                        try {
                            refreshFavorites()
                        } catch (e: Exception) {
                            error = e.message
                        } finally {
                            loading = false
                        }
                    }
                }, enabled = !loading, modifier = Modifier.semantics { contentDescription = "Load favorites list" }) { Text("Load") }

                Button(onClick = {
                    scope.launch {
                        loading = true
                        error = null
                        try {
                            syncQueue()
                        } catch (e: Exception) {
                            error = e.message
                        } finally {
                            loading = false
                        }
                    }
                }, enabled = !loading, modifier = Modifier.semantics { contentDescription = "Sync favorites queue" }) { Text("Sync queue (${queue.size})") }
            }
        }
        item {
            OutlinedTextField(
                modifier = Modifier
                    .fillMaxWidth()
                    .semantics { contentDescription = "Recipe ID input" },
                value = recipeId,
                onValueChange = { recipeId = it },
                label = { Text("Recipe ID") }
            )
        }
        item {
            Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                Button(onClick = {
                    val id = recipeId.trim()
                    if (id.isBlank()) return@Button
                    scope.launch {
                        loading = true
                        error = null
                        try {
                            withContext(Dispatchers.IO) { client.addFavorite(userId.trim(), id, token.trim()) }
                            refreshFavorites()
                        } catch (_: Exception) {
                            enqueue("add", id)
                        } finally {
                            loading = false
                        }
                    }
                }, enabled = !loading, modifier = Modifier.semantics { contentDescription = "Add recipe to favorites" }) { Text("Add") }

                Button(onClick = {
                    val id = recipeId.trim()
                    if (id.isBlank()) return@Button
                    scope.launch {
                        loading = true
                        error = null
                        try {
                            withContext(Dispatchers.IO) { client.removeFavorite(userId.trim(), id, token.trim()) }
                            refreshFavorites()
                        } catch (_: Exception) {
                            enqueue("remove", id)
                        } finally {
                            loading = false
                        }
                    }
                }, enabled = !loading, modifier = Modifier.semantics { contentDescription = "Remove recipe from favorites" }) { Text("Remove") }
            }
        }

        if (error != null) {
            item { Card { Text("Error: $error", modifier = Modifier.padding(12.dp)) } }
        }

        if (notice != null) {
            item { Card { Text(notice.orEmpty(), modifier = Modifier.padding(12.dp)) } }
        }

        item { Text("Saved (${favorites.size})", modifier = Modifier.semantics { heading() }) }
        items(favorites) { item ->
            Card(modifier = Modifier.fillMaxWidth()) {
                Column(modifier = Modifier.padding(12.dp), verticalArrangement = Arrangement.spacedBy(6.dp)) {
                    Text(item.recipeName ?: item.recipeId)
                    Text("recipeId: ${item.recipeId}")
                }
            }
        }
    }
}

private fun loadQueue(context: Context): List<FavoriteQueueAction> {
    val prefs = context.getSharedPreferences(PREFS, Context.MODE_PRIVATE)
    val raw = prefs.getString(KEY_QUEUE, "[]") ?: "[]"
    val array = JSONArray(raw)
    val out = mutableListOf<FavoriteQueueAction>()
    for (i in 0 until array.length()) {
        val item = array.optJSONObject(i) ?: continue
        out += FavoriteQueueAction(
            op = item.optString("op", "add"),
            recipeId = item.optString("recipeId", ""),
            queuedAtEpochMs = item.optLong("queuedAtEpochMs", 0L)
        )
    }
    return compactQueue(out.filter { it.recipeId.isNotBlank() })
}

private fun saveQueue(context: Context, queue: List<FavoriteQueueAction>) {
    val prefs = context.getSharedPreferences(PREFS, Context.MODE_PRIVATE)
    val array = JSONArray()
    queue.forEach { action ->
        array.put(
            JSONObject().apply {
                put("op", action.op)
                put("recipeId", action.recipeId)
                put("queuedAtEpochMs", action.queuedAtEpochMs)
            }
        )
    }
    prefs.edit().putString(KEY_QUEUE, array.toString()).apply()
}

private fun compactQueue(queue: List<FavoriteQueueAction>): List<FavoriteQueueAction> {
    val out = mutableListOf<FavoriteQueueAction>()
    for (action in queue.sortedBy { it.queuedAtEpochMs }) {
        val existingIndex = out.indexOfLast { it.recipeId == action.recipeId }
        if (existingIndex == -1) {
            out += action
            continue
        }

        val existing = out[existingIndex]
        if (existing.op == "add" && action.op == "remove") {
            out.removeAt(existingIndex)
            continue
        }

        if (existing.op == "remove" && action.op == "add") {
            out[existingIndex] = action
            continue
        }

        if (existing.op == action.op) {
            out[existingIndex] = action
            continue
        }

        out += action
    }

    return out
}
