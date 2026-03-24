import SwiftUI
import UIKit

struct AccountView: View {
    @Binding var session: AuthSession?

    @State private var username: String = ""
    @State private var password: String = ""
    @State private var loading: Bool = false
    @State private var error: String?
    @State private var notice: String?
    @State private var sessions: [SessionItem] = []

    private let apiClient = RecipesAPIClient()

    var body: some View {
        NavigationStack {
            List {
                Section {
                    Text("Manage login, secure sessions, and device access.")
                        .font(.footnote)
                        .foregroundStyle(.secondary)
                }

                Section("Login") {
                    TextField("Username or email", text: $username)
                        .textInputAutocapitalization(.never)
                        .autocorrectionDisabled(true)
                        .accessibilityLabel("Username or email")
                    SecureField("Password", text: $password)
                        .textInputAutocapitalization(.never)
                        .autocorrectionDisabled(true)
                        .accessibilityLabel("Password")

                    Button(loading ? "Working..." : "Login") {
                        Task { await login() }
                    }
                    .disabled(loading)
                    .accessibilityLabel("Login")
                }

                if let session {
                    Section("Current Session") {
                        Text("User: \(session.username)")
                        Text("User ID: \(session.userId)")
                        Text("Session ID: \(session.sessionId)")
                        Text("Access expires: \(session.expiresAt)")
                    }

                    Section("Session Controls") {
                        Button("Refresh token") {
                            Task { await refresh() }
                        }
                        .disabled(loading)
                        .accessibilityLabel("Refresh token")

                        Button("List active sessions") {
                            Task { await listSessions() }
                        }
                        .disabled(loading)
                        .accessibilityLabel("List active sessions")

                        Button("Logout this session") {
                            Task { await logoutSession() }
                        }
                        .disabled(loading)
                        .accessibilityLabel("Logout this session")

                        Button("Logout all sessions") {
                            Task { await logoutAll() }
                        }
                        .disabled(loading)
                        .accessibilityLabel("Logout all sessions")
                    }
                }

                if !sessions.isEmpty {
                    Section("Active Sessions") {
                        ForEach(sessions) { item in
                            VStack(alignment: .leading, spacing: 4) {
                                Text(item.sessionId)
                                    .font(.headline)
                                Text("Created: \(item.createdAt)")
                                    .font(.footnote)
                                if let lastUsedAt = item.lastUsedAt {
                                    Text("Last used: \(lastUsedAt)")
                                        .font(.footnote)
                                }
                                if let ipAddress = item.ipAddress {
                                    Text("IP: \(ipAddress)")
                                        .font(.footnote)
                                }
                            }
                            .padding(.vertical, 2)
                        }
                    }
                }

                if let notice {
                    Section("Status") {
                        Text(notice)
                    }
                }

                if let error {
                    Section("Error") {
                        Text(error)
                            .foregroundStyle(.red)
                    }
                }
            }
            .listStyle(.insetGrouped)
            .scrollContentBackground(.hidden)
            .navigationTitle("Profile")
            .onChange(of: notice) { _, next in
                guard let next, !next.isEmpty else { return }
                UIAccessibility.post(notification: .announcement, argument: next)
            }
            .onChange(of: error) { _, next in
                guard let next, !next.isEmpty else { return }
                UIAccessibility.post(notification: .announcement, argument: "Account error. \(next)")
            }
        }
    }

    private func login() async {
        let user = username.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !user.isEmpty, !password.isEmpty else {
            error = "Username/email and password are required"
            return
        }

        loading = true
        defer { loading = false }
        error = nil
        notice = nil

        do {
            let response = try await apiClient.login(
                username: user,
                password: password,
                clientSessionID: AuthSessionStore.clientSessionID()
            )

            let next = AuthSession(
                userId: response.id,
                username: response.username,
                email: response.email,
                role: response.role,
                token: response.token,
                expiresAt: response.expiresAt,
                refreshToken: response.refreshToken,
                refreshExpiresAt: response.refreshExpiresAt,
                sessionId: response.sessionId
            )

            AuthSessionStore.save(next)
            session = next
            password = ""
            notice = "Logged in and session saved in keychain"
        } catch {
            self.error = error.localizedDescription
        }
    }

    private func refresh() async {
        guard let session else { return }

        loading = true
        defer { loading = false }
        error = nil
        notice = nil

        do {
            let refreshed = try await apiClient.refresh(refreshToken: session.refreshToken)
            let next = AuthSession(
                userId: session.userId,
                username: session.username,
                email: session.email,
                role: session.role,
                token: refreshed.token,
                expiresAt: refreshed.expiresAt,
                refreshToken: refreshed.refreshToken,
                refreshExpiresAt: refreshed.refreshExpiresAt,
                sessionId: refreshed.sessionId
            )

            AuthSessionStore.save(next)
            self.session = next
            notice = "Session refreshed"
        } catch {
            self.error = error.localizedDescription
        }
    }

    private func listSessions() async {
        guard let session else { return }

        loading = true
        defer { loading = false }
        error = nil
        notice = nil

        do {
            sessions = try await apiClient.listSessions(userID: session.userId, token: session.token)
            notice = "Loaded active sessions"
        } catch {
            self.error = error.localizedDescription
        }
    }

    private func logoutSession() async {
        guard let session else { return }

        loading = true
        defer { loading = false }
        error = nil
        notice = nil

        do {
            try await apiClient.logoutSession(refreshToken: session.refreshToken)
            AuthSessionStore.clear()
            self.session = nil
            sessions = []
            notice = "Current session revoked"
        } catch {
            self.error = error.localizedDescription
        }
    }

    private func logoutAll() async {
        guard let session else { return }

        loading = true
        defer { loading = false }
        error = nil
        notice = nil

        do {
            try await apiClient.logout(refreshToken: session.refreshToken)
            AuthSessionStore.clear()
            self.session = nil
            sessions = []
            notice = "Logged out all sessions"
        } catch {
            self.error = error.localizedDescription
        }
    }
}
