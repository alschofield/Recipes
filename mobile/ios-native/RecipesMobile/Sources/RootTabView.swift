import SwiftUI

struct RootTabView: View {
    @State private var session: AuthSession? = AuthSessionStore.load()

    var body: some View {
        ZStack {
            BrandTheme.appBackground
                .ignoresSafeArea()

            TabView {
                SearchView()
                    .tabItem {
                        Label("Discover", systemImage: "magnifyingglass")
                    }

                FavoritesView(session: $session)
                    .tabItem {
                        Label("Saved", systemImage: "heart")
                    }

                AccountView(session: $session)
                    .tabItem {
                        Label("Profile", systemImage: "person")
                    }
            }
            .tint(BrandTheme.accent)
        }
    }
}
