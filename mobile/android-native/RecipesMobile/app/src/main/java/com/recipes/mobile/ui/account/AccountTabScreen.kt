package com.recipes.mobile.ui.account

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
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
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.platform.LocalView
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.heading
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.input.PasswordVisualTransformation
import androidx.compose.ui.unit.dp
import com.recipes.mobile.BuildConfig
import com.recipes.mobile.core.network.AuthSession
import com.recipes.mobile.core.network.RecipesApiClient
import com.recipes.mobile.core.network.SessionItem
import com.recipes.mobile.core.session.SessionStore
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext

@Composable
fun AccountTabScreen(
    initialSession: AuthSession?,
    onSessionChanged: (AuthSession?) -> Unit,
) {
    val context = LocalContext.current
    val view = LocalView.current
    val scope = rememberCoroutineScope()
    val client = remember { RecipesApiClient(BuildConfig.API_BASE_URL) }
    val store = remember { SessionStore(context) }

    var username by remember { mutableStateOf("") }
    var password by remember { mutableStateOf("") }
    var loading by remember { mutableStateOf(false) }
    var error by remember { mutableStateOf<String?>(null) }
    var notice by remember { mutableStateOf<String?>(null) }
    var sessions by remember { mutableStateOf<List<SessionItem>>(emptyList()) }

    LaunchedEffect(notice) {
        if (!notice.isNullOrBlank()) {
            view.announceForAccessibility(notice)
        }
    }

    LaunchedEffect(error) {
        if (!error.isNullOrBlank()) {
            view.announceForAccessibility("Account error. $error")
        }
    }

    val session = initialSession

    LazyColumn(
        modifier = Modifier
            .fillMaxWidth()
            .padding(16.dp),
        verticalArrangement = Arrangement.spacedBy(12.dp)
    ) {
        item { Text("Profile", modifier = Modifier.semantics { heading() }) }
        item { Text("Manage login, secure sessions, and device access.") }
        item {
            OutlinedTextField(
                modifier = Modifier
                    .fillMaxWidth()
                    .semantics { contentDescription = "Username or email input" },
                value = username,
                onValueChange = { username = it },
                label = { Text("Username or email") }
            )
        }
        item {
            OutlinedTextField(
                modifier = Modifier
                    .fillMaxWidth()
                    .semantics { contentDescription = "Password input" },
                value = password,
                onValueChange = { password = it },
                label = { Text("Password") },
                visualTransformation = PasswordVisualTransformation()
            )
        }

        item {
            Button(
                onClick = {
                    val user = username.trim()
                    val pass = password
                    if (user.isBlank() || pass.isBlank()) {
                        error = "Username/email and password are required"
                        return@Button
                    }

                    scope.launch {
                        loading = true
                        error = null
                        notice = null
                        try {
                            val login = withContext(Dispatchers.IO) {
                                client.login(user, pass, store.clientSessionId())
                            }
                            val next = AuthSession(
                                userId = login.id,
                                username = login.username,
                                email = login.email,
                                role = login.role,
                                token = login.token,
                                expiresAt = login.expiresAt,
                                refreshToken = login.refreshToken,
                                refreshExpiresAt = login.refreshExpiresAt,
                                sessionId = login.sessionId,
                            )
                            store.save(next)
                            onSessionChanged(next)
                            notice = "Logged in and session saved securely"
                            password = ""
                        } catch (e: Exception) {
                            error = e.message
                        } finally {
                            loading = false
                        }
                    }
                },
                enabled = !loading,
                modifier = Modifier.semantics { contentDescription = "Login" }
            ) {
                Text(if (loading) "Working..." else "Login")
            }
        }

        if (session != null) {
            item {
                Card(modifier = Modifier.fillMaxWidth()) {
                    Column(modifier = Modifier.padding(12.dp), verticalArrangement = Arrangement.spacedBy(4.dp)) {
                        Text("Signed in as ${session.username}")
                        Text("User ID: ${session.userId}")
                        Text("Session ID: ${session.sessionId}")
                        Text("Access expires: ${session.expiresAt}")
                    }
                }
            }

            item {
                Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
                    Button(onClick = {
                        scope.launch {
                            loading = true
                            error = null
                            notice = null
                            try {
                                val refreshed = withContext(Dispatchers.IO) {
                                    client.refresh(session.refreshToken)
                                }
                                val next = session.copy(
                                    token = refreshed.token,
                                    expiresAt = refreshed.expiresAt,
                                    refreshToken = refreshed.refreshToken,
                                    refreshExpiresAt = refreshed.refreshExpiresAt,
                                    sessionId = refreshed.sessionId,
                                )
                                store.save(next)
                                onSessionChanged(next)
                                notice = "Session refreshed"
                            } catch (e: Exception) {
                                error = e.message
                            } finally {
                                loading = false
                            }
                        }
                    }, enabled = !loading, modifier = Modifier.semantics { contentDescription = "Refresh session token" }) {
                        Text("Refresh token")
                    }

                    Button(onClick = {
                        scope.launch {
                            loading = true
                            error = null
                            notice = null
                            try {
                                withContext(Dispatchers.IO) {
                                    client.listSessions(session.userId, session.token)
                                }.also { sessions = it }
                                notice = "Loaded sessions"
                            } catch (e: Exception) {
                                error = e.message
                            } finally {
                                loading = false
                            }
                        }
                    }, enabled = !loading, modifier = Modifier.semantics { contentDescription = "List active sessions" }) {
                        Text("List active sessions")
                    }

                    Button(onClick = {
                        scope.launch {
                            loading = true
                            error = null
                            notice = null
                            try {
                                withContext(Dispatchers.IO) { client.logoutSession(session.refreshToken) }
                                store.clear()
                                onSessionChanged(null)
                                sessions = emptyList()
                                notice = "Current session revoked"
                            } catch (e: Exception) {
                                error = e.message
                            } finally {
                                loading = false
                            }
                        }
                    }, enabled = !loading, modifier = Modifier.semantics { contentDescription = "Logout current session" }) {
                        Text("Logout this session")
                    }

                    Button(onClick = {
                        scope.launch {
                            loading = true
                            error = null
                            notice = null
                            try {
                                withContext(Dispatchers.IO) { client.logout(session.refreshToken) }
                                store.clear()
                                onSessionChanged(null)
                                sessions = emptyList()
                                notice = "Logged out all sessions"
                            } catch (e: Exception) {
                                error = e.message
                            } finally {
                                loading = false
                            }
                        }
                    }, enabled = !loading, modifier = Modifier.semantics { contentDescription = "Logout all sessions" }) {
                        Text("Logout all sessions")
                    }
                }
            }
        }

        if (notice != null) {
            item {
                Card { Text(notice.orEmpty(), modifier = Modifier.padding(12.dp)) }
            }
        }

        if (error != null) {
            item {
                Card { Text("Error: $error", modifier = Modifier.padding(12.dp)) }
            }
        }

        if (sessions.isNotEmpty()) {
            item { Text("Active Sessions (${sessions.size})", modifier = Modifier.semantics { heading() }) }
            items(sessions) { item ->
                Card(modifier = Modifier.fillMaxWidth()) {
                    Column(modifier = Modifier.padding(12.dp), verticalArrangement = Arrangement.spacedBy(4.dp)) {
                        Text(item.sessionId)
                        Text("Created: ${item.createdAt}")
                        if (!item.lastUsedAt.isNullOrBlank()) {
                            Text("Last used: ${item.lastUsedAt}")
                        }
                        if (!item.ipAddress.isNullOrBlank()) {
                            Text("IP: ${item.ipAddress}")
                        }
                    }
                }
            }
        }
    }
}
