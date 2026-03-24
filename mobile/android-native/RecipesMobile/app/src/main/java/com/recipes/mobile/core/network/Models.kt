package com.recipes.mobile.core.network

data class SearchRecipeItem(
    val id: String,
    val name: String,
    val source: String,
    val matchPercent: Double,
)

data class SearchPayload(
    val mode: String,
    val total: Int,
    val results: List<SearchRecipeItem>,
)

data class FavoriteItem(
    val id: String,
    val userId: String,
    val recipeId: String,
    val recipeName: String?,
)

data class FavoriteQueueAction(
    val op: String,
    val recipeId: String,
    val queuedAtEpochMs: Long,
)

data class LoginResponse(
    val id: String,
    val username: String,
    val email: String,
    val role: String,
    val token: String,
    val expiresAt: String,
    val refreshToken: String,
    val refreshExpiresAt: String,
    val sessionId: String,
)

data class RefreshResponse(
    val token: String,
    val expiresAt: String,
    val refreshToken: String,
    val refreshExpiresAt: String,
    val sessionId: String,
)

data class SessionItem(
    val sessionId: String,
    val createdAt: String,
    val lastUsedAt: String?,
    val userAgent: String?,
    val ipAddress: String?,
)

data class AuthSession(
    val userId: String,
    val username: String,
    val email: String,
    val role: String,
    val token: String,
    val expiresAt: String,
    val refreshToken: String,
    val refreshExpiresAt: String,
    val sessionId: String,
)
