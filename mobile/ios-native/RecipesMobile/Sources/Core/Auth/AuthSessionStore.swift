import Foundation
import Security

enum AuthSessionStore {
    private static let service = "com.recipes.mobile.auth"
    private static let account = "session.v1"
    private static let clientSessionKey = "recipes.mobile.clientSessionID"

    static func load() -> AuthSession? {
        guard let data = readKeychain() else {
            return nil
        }
        return try? JSONDecoder().decode(AuthSession.self, from: data)
    }

    static func save(_ session: AuthSession) {
        guard let data = try? JSONEncoder().encode(session) else {
            return
        }
        upsertKeychain(data: data)
    }

    static func clear() {
        deleteKeychain()
    }

    static func clientSessionID() -> String {
        if let existing = UserDefaults.standard.string(forKey: clientSessionKey), !existing.isEmpty {
            return existing
        }

        let generated = UUID().uuidString
        UserDefaults.standard.set(generated, forKey: clientSessionKey)
        return generated
    }

    private static func query() -> [String: Any] {
        [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: account,
        ]
    }

    private static func readKeychain() -> Data? {
        var request = query()
        request[kSecReturnData as String] = true
        request[kSecMatchLimit as String] = kSecMatchLimitOne

        var result: CFTypeRef?
        let status = SecItemCopyMatching(request as CFDictionary, &result)
        guard status == errSecSuccess else {
            return nil
        }
        return result as? Data
    }

    private static func upsertKeychain(data: Data) {
        let status = SecItemAdd(
            query().merging([kSecValueData as String: data]) { _, new in new } as CFDictionary,
            nil
        )

        if status == errSecDuplicateItem {
            let update: [String: Any] = [kSecValueData as String: data]
            SecItemUpdate(query() as CFDictionary, update as CFDictionary)
        }
    }

    private static func deleteKeychain() {
        SecItemDelete(query() as CFDictionary)
    }
}
