import UIKit
import SwiftUI
import Shared

/// Bridges Kotlin's Compose Multiplatform UI into a SwiftUI hierarchy. The Shared
/// framework's `MainViewController(apiBaseUrl:)` does double duty: it starts Koin
/// the first time it's called and returns the Compose UI controller.
///
/// `apiBaseUrl` is hard-coded for now. For real builds, read from `Info.plist`
/// or a Build Configuration setting and switch on debug/release.
struct ComposeView: UIViewControllerRepresentable {
    func makeUIViewController(context: Self.Context) -> UIViewController {
        MainViewControllerKt.MainViewController(apiBaseUrl: "http://localhost:8080")
    }

    func updateUIViewController(_ uiViewController: UIViewController, context: Self.Context) {}
}

struct ContentView: View {
    var body: some View {
        ComposeView()
            .ignoresSafeArea()
    }
}
