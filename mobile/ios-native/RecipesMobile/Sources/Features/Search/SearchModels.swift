import Foundation

struct SearchRecipeItem: Decodable, Identifiable {
    let id: String
    let name: String
    let source: String
    let matchPercent: Double?
}

struct SearchPagination: Decodable {
    let total: Int
}

struct SearchResponse: Decodable {
    let mode: String
    let pagination: SearchPagination
    let results: [SearchRecipeItem]
}

struct SearchRequestPayload: Encodable {
    let ingredients: [String]
    let mode: String
    let complex: Bool
    let dbOnly: Bool
}

struct FavoriteItem: Decodable, Identifiable {
    let id: String
    let userId: String
    let recipeId: String
    let recipeName: String?
}

struct FavoriteQueueAction: Codable, Identifiable {
    let id: UUID
    let op: String
    let recipeId: String
    let queuedAt: Date

    private enum CodingKeys: String, CodingKey {
        case id
        case op
        case recipeId
        case queuedAt
    }

    init(op: String, recipeId: String) {
        self.id = UUID()
        self.op = op
        self.recipeId = recipeId
        self.queuedAt = Date()
    }

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        id = try container.decodeIfPresent(UUID.self, forKey: .id) ?? UUID()
        op = try container.decode(String.self, forKey: .op)
        recipeId = try container.decode(String.self, forKey: .recipeId)
        queuedAt = try container.decodeIfPresent(Date.self, forKey: .queuedAt) ?? Date()
    }
}

struct LoginResponse: Decodable {
    let id: String
    let username: String
    let email: String
    let role: String
    let token: String
    let expiresAt: String
    let refreshToken: String
    let refreshExpiresAt: String
    let sessionId: String
}

struct RefreshResponse: Decodable {
    let token: String
    let expiresAt: String
    let refreshToken: String
    let refreshExpiresAt: String
    let sessionId: String
}

struct SessionItem: Decodable, Identifiable {
    let sessionId: String
    let createdAt: String
    let lastUsedAt: String?
    let userAgent: String?
    let ipAddress: String?

    var id: String { sessionId }
}

struct AuthSession: Codable {
    let userId: String
    let username: String
    let email: String
    let role: String
    let token: String
    let expiresAt: String
    let refreshToken: String
    let refreshExpiresAt: String
    let sessionId: String
}
