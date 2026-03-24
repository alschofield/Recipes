import SwiftUI
import UIKit

struct FavoritesView: View {
    @Binding var session: AuthSession?

    @State private var userID: String = ""
    @State private var token: String = ""
    @State private var recipeID: String = ""
    @State private var loading: Bool = false
    @State private var error: String?
    @State private var notice: String?
    @State private var favorites: [FavoriteItem] = []
    @State private var queue: [FavoriteQueueAction] = FavoriteQueueStore.compact(FavoriteQueueStore.load())

    private let apiClient = RecipesAPIClient()

    var body: some View {
        NavigationStack {
            List {
                Section("User") {
                    if let session {
                        Text("Signed in as \(session.username)")
                            .font(.footnote)
                        Button("Use saved session") {
                            userID = session.userId
                            token = session.token
                            notice = "Using latest saved session token"
                        }
                        .accessibilityLabel("Use saved session")
                        .accessibilityHint("Fills user ID and token from secure session")
                    } else {
                        Text("Sign in from Profile to use secure session")
                            .font(.footnote)
                    }

                    TextField("User ID", text: $userID)
                        .textInputAutocapitalization(.never)
                        .autocorrectionDisabled(true)
                        .accessibilityLabel("User ID")
                    SecureField("Access token", text: $token)
                        .textInputAutocapitalization(.never)
                        .autocorrectionDisabled(true)
                        .accessibilityLabel("Access token")
                }

                Section("Actions") {
                    Button(loading ? "Loading..." : "Load Favorites") {
                        Task { await loadFavorites() }
                    }
                    .disabled(loading)
                    .accessibilityLabel("Load favorites")

                    Button(loading ? "Syncing..." : "Sync Queue (\(queue.count))") {
                        Task { await syncQueue() }
                    }
                    .disabled(loading)
                    .accessibilityLabel("Sync favorites queue")
                    .accessibilityHint("Replays offline actions and refreshes from server")
                }

                Section("Mutate") {
                    TextField("Recipe ID", text: $recipeID)
                        .textInputAutocapitalization(.never)
                        .autocorrectionDisabled(true)
                        .accessibilityLabel("Recipe ID")

                    HStack {
                        Button("Add") {
                            Task { await addFavorite() }
                        }
                        .disabled(loading)
                        .accessibilityLabel("Add favorite")

                        Button("Remove") {
                            Task { await removeFavorite() }
                        }
                        .disabled(loading)
                        .accessibilityLabel("Remove favorite")
                    }
                }

                if let error {
                    Section("Error") {
                        Text(error)
                            .foregroundStyle(.red)
                    }
                }

                if let notice {
                    Section("Status") {
                        Text(notice)
                    }
                }

                Section("Favorites") {
                    ForEach(favorites) { item in
                        VStack(alignment: .leading, spacing: 4) {
                            Text(item.recipeName ?? item.recipeId)
                                .font(.headline)
                            Text("recipeId: \(item.recipeId)")
                                .font(.footnote)
                                .foregroundStyle(.secondary)
                        }
                        .padding(.vertical, 4)
                    }
                }
            }
            .listStyle(.insetGrouped)
            .scrollContentBackground(.hidden)
            .navigationTitle("Saved Recipes")
            .onAppear {
                if let session {
                    userID = session.userId
                    token = session.token
                }
            }
            .onChange(of: session?.sessionId) { _, _ in
                if let session {
                    userID = session.userId
                    token = session.token
                }
            }
            .onChange(of: notice) { _, next in
                guard let next, !next.isEmpty else { return }
                UIAccessibility.post(notification: .announcement, argument: next)
            }
            .onChange(of: error) { _, next in
                guard let next, !next.isEmpty else { return }
                UIAccessibility.post(notification: .announcement, argument: "Favorites error. \(next)")
            }
        }
    }

    private func loadFavorites() async {
        guard !userID.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty,
              !token.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty else {
            return
        }

        loading = true
        defer { loading = false }
        error = nil

        do {
            favorites = try await apiClient.listFavorites(
                userID: userID.trimmingCharacters(in: .whitespacesAndNewlines),
                token: token.trimmingCharacters(in: .whitespacesAndNewlines)
            )
        } catch {
            self.error = error.localizedDescription
        }
    }

    private func addFavorite() async {
        let trimmedID = recipeID.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmedID.isEmpty else { return }

        loading = true
        defer { loading = false }
        error = nil

        do {
            try await apiClient.addFavorite(
                userID: userID.trimmingCharacters(in: .whitespacesAndNewlines),
                recipeID: trimmedID,
                token: token.trimmingCharacters(in: .whitespacesAndNewlines)
            )
            await loadFavorites()
        } catch {
            queue = FavoriteQueueStore.compact(queue + [FavoriteQueueAction(op: "add", recipeId: trimmedID)])
            FavoriteQueueStore.save(queue)
            notice = "Saved action offline. Sync queue to reconcile."
        }
    }

    private func removeFavorite() async {
        let trimmedID = recipeID.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmedID.isEmpty else { return }

        loading = true
        defer { loading = false }
        error = nil

        do {
            try await apiClient.removeFavorite(
                userID: userID.trimmingCharacters(in: .whitespacesAndNewlines),
                recipeID: trimmedID,
                token: token.trimmingCharacters(in: .whitespacesAndNewlines)
            )
            await loadFavorites()
        } catch {
            queue = FavoriteQueueStore.compact(queue + [FavoriteQueueAction(op: "remove", recipeId: trimmedID)])
            FavoriteQueueStore.save(queue)
            notice = "Saved action offline. Sync queue to reconcile."
        }
    }

    private func syncQueue() async {
        guard !queue.isEmpty else { return }
        loading = true
        defer { loading = false }
        error = nil
        notice = nil

        var remaining: [FavoriteQueueAction] = []
        var replayed = 0
        for action in queue {
            do {
                if action.op == "add" {
                    try await apiClient.addFavorite(
                        userID: userID.trimmingCharacters(in: .whitespacesAndNewlines),
                        recipeID: action.recipeId,
                        token: token.trimmingCharacters(in: .whitespacesAndNewlines)
                    )
                } else {
                    try await apiClient.removeFavorite(
                        userID: userID.trimmingCharacters(in: .whitespacesAndNewlines),
                        recipeID: action.recipeId,
                        token: token.trimmingCharacters(in: .whitespacesAndNewlines)
                    )
                }
                replayed += 1
            } catch {
                remaining.append(action)
            }
        }

        let compacted = FavoriteQueueStore.compact(remaining)
        queue = compacted
        FavoriteQueueStore.save(compacted)
        notice = "Replayed \(replayed) action(s). Pending: \(compacted.count)."
        await loadFavorites()
    }
}

enum FavoriteQueueStore {
    private static let key = "recipes.favorite.queue.v1"

    static func load() -> [FavoriteQueueAction] {
        guard let data = UserDefaults.standard.data(forKey: key) else {
            return []
        }
        return (try? JSONDecoder().decode([FavoriteQueueAction].self, from: data)) ?? []
    }

    static func save(_ queue: [FavoriteQueueAction]) {
        let data = try? JSONEncoder().encode(queue)
        UserDefaults.standard.set(data, forKey: key)
    }

    static func compact(_ queue: [FavoriteQueueAction]) -> [FavoriteQueueAction] {
        var out: [FavoriteQueueAction] = []

        for action in queue.sorted(by: { $0.queuedAt < $1.queuedAt }) {
            guard let idx = out.lastIndex(where: { $0.recipeId == action.recipeId }) else {
                out.append(action)
                continue
            }

            let previous = out[idx]
            if previous.op == "add" && action.op == "remove" {
                out.remove(at: idx)
                continue
            }

            if previous.op == "remove" && action.op == "add" {
                out[idx] = action
                continue
            }

            if previous.op == action.op {
                out[idx] = action
                continue
            }

            out.append(action)
        }

        return out
    }
}
