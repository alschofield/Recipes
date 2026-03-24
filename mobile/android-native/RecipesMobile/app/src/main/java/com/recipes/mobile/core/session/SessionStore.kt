package com.recipes.mobile.core.session

import android.content.Context
import android.content.SharedPreferences
import com.recipes.mobile.core.network.AuthSession
import androidx.security.crypto.EncryptedSharedPreferences
import androidx.security.crypto.MasterKey
import java.util.UUID

class SessionStore(context: Context) {
    private val prefs: SharedPreferences = createPrefs(context)

    fun load(): AuthSession? {
        val userId = prefs.getString(KEY_USER_ID, null) ?: return null
        val username = prefs.getString(KEY_USERNAME, null) ?: return null
        val email = prefs.getString(KEY_EMAIL, null) ?: return null
        val role = prefs.getString(KEY_ROLE, null) ?: return null
        val token = prefs.getString(KEY_TOKEN, null) ?: return null
        val expiresAt = prefs.getString(KEY_EXPIRES_AT, null) ?: return null
        val refreshToken = prefs.getString(KEY_REFRESH_TOKEN, null) ?: return null
        val refreshExpiresAt = prefs.getString(KEY_REFRESH_EXPIRES_AT, null) ?: return null
        val sessionId = prefs.getString(KEY_SESSION_ID, null) ?: return null

        return AuthSession(
            userId = userId,
            username = username,
            email = email,
            role = role,
            token = token,
            expiresAt = expiresAt,
            refreshToken = refreshToken,
            refreshExpiresAt = refreshExpiresAt,
            sessionId = sessionId,
        )
    }

    fun save(session: AuthSession) {
        prefs.edit()
            .putString(KEY_USER_ID, session.userId)
            .putString(KEY_USERNAME, session.username)
            .putString(KEY_EMAIL, session.email)
            .putString(KEY_ROLE, session.role)
            .putString(KEY_TOKEN, session.token)
            .putString(KEY_EXPIRES_AT, session.expiresAt)
            .putString(KEY_REFRESH_TOKEN, session.refreshToken)
            .putString(KEY_REFRESH_EXPIRES_AT, session.refreshExpiresAt)
            .putString(KEY_SESSION_ID, session.sessionId)
            .apply()
    }

    fun clear() {
        prefs.edit().clear().apply()
    }

    fun clientSessionId(): String {
        val existing = prefs.getString(KEY_CLIENT_SESSION_ID, null)
        if (!existing.isNullOrBlank()) {
            return existing
        }

        val generated = UUID.randomUUID().toString()
        prefs.edit().putString(KEY_CLIENT_SESSION_ID, generated).apply()
        return generated
    }

    private fun createPrefs(context: Context): SharedPreferences {
        return try {
            val key = MasterKey.Builder(context)
                .setKeyScheme(MasterKey.KeyScheme.AES256_GCM)
                .build()

            EncryptedSharedPreferences.create(
                context,
                PREFS_NAME,
                key,
                EncryptedSharedPreferences.PrefKeyEncryptionScheme.AES256_SIV,
                EncryptedSharedPreferences.PrefValueEncryptionScheme.AES256_GCM,
            )
        } catch (_: Exception) {
            context.getSharedPreferences(PREFS_NAME_FALLBACK, Context.MODE_PRIVATE)
        }
    }

    private companion object {
        const val PREFS_NAME = "recipes_secure_session"
        const val PREFS_NAME_FALLBACK = "recipes_secure_session_fallback"
        const val KEY_USER_ID = "user_id"
        const val KEY_USERNAME = "username"
        const val KEY_EMAIL = "email"
        const val KEY_ROLE = "role"
        const val KEY_TOKEN = "token"
        const val KEY_EXPIRES_AT = "expires_at"
        const val KEY_REFRESH_TOKEN = "refresh_token"
        const val KEY_REFRESH_EXPIRES_AT = "refresh_expires_at"
        const val KEY_SESSION_ID = "session_id"
        const val KEY_CLIENT_SESSION_ID = "client_session_id"
    }
}
