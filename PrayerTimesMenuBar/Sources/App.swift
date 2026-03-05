import SwiftUI

@main
struct PrayerTimesMenuBarApp: App {
    @NSApplicationDelegateAdaptor(AppDelegate.self) private var appDelegate

    var body: some Scene {
        MenuBarExtra {
            MenuBarView(viewModel: appDelegate.viewModel)
        } label: {
            HStack(spacing: 4) {
                Image(systemName: "moon.stars")
                Text(appDelegate.viewModel.menuBarTitle)
            }
        }
        .menuBarExtraStyle(.window)
    }
}

/// App delegate that owns the view model and starts the refresh loop
/// immediately at launch, before the user ever opens the menu.
final class AppDelegate: NSObject, NSApplicationDelegate {
    let viewModel = PrayerTimerViewModel()

    func applicationDidFinishLaunching(_ notification: Notification) {
        viewModel.start()
    }
}
