import Foundation
import SwiftUI

/// View model that manages prayer time data and periodic refresh.
@Observable
final class PrayerTimerViewModel {
    // MARK: - Published state

    /// The formatted string to show in the menubar.
    var menuBarTitle: String = "..."

    /// Today's full prayer schedule.
    var todayResponse: TodayResponse?

    /// Whether data is currently being fetched.
    var isLoading: Bool = false

    /// The last error message, if any.
    var errorMessage: String?

    // MARK: - Settings

    /// The selected display format. Persisted separately via AppStorage in the view layer.
    var displayFormat: DisplayFormat = .nameAndTime {
        didSet {
            if oldValue != displayFormat {
                Task { await refreshMenuBarTitle() }
            }
        }
    }

    // MARK: - Private

    private let cli = CLIService()
    private var refreshTimer: Timer?
    private var hasStarted = false

    // MARK: - Lifecycle

    /// Starts the periodic refresh loop. Safe to call multiple times — only the first call takes effect.
    func start() {
        guard !hasStarted else { return }
        hasStarted = true
        Task { await refreshAll() }
        startTimer()
    }

    /// Stops the refresh timer.
    func stop() {
        refreshTimer?.invalidate()
        refreshTimer = nil
    }

    // MARK: - Refresh

    /// Refreshes all data: menubar title + today's schedule.
    func refreshAll() async {
        isLoading = todayResponse == nil // only show loading on first fetch
        errorMessage = nil

        async let nextTask: Void = refreshMenuBarTitle()
        async let todayTask: Void = refreshToday()

        _ = await (nextTask, todayTask)

        isLoading = false
    }

    /// Refreshes only the menubar title string using the CLI `next --format` command.
    func refreshMenuBarTitle() async {
        do {
            let formatted = try await cli.fetchNextFormatted(format: displayFormat.rawValue)
            menuBarTitle = formatted.isEmpty ? "..." : formatted
            errorMessage = nil
        } catch {
            // Keep the old title if we had one; only show error if we have nothing
            if menuBarTitle == "..." {
                menuBarTitle = "Prayer Times"
            }
            errorMessage = error.localizedDescription
        }
    }

    /// Refreshes today's full prayer schedule.
    func refreshToday() async {
        do {
            todayResponse = try await cli.fetchToday()
            errorMessage = nil
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    // MARK: - Timer

    private func startTimer() {
        refreshTimer?.invalidate()
        refreshTimer = Timer.scheduledTimer(withTimeInterval: 60, repeats: true) { [weak self] _ in
            guard let self else { return }
            Task { await self.refreshAll() }
        }
    }
}
