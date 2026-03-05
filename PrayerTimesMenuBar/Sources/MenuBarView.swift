import SwiftUI
import ServiceManagement

struct MenuBarView: View {
    @Bindable var viewModel: PrayerTimerViewModel
    @AppStorage("displayFormat") private var storedFormat: String = DisplayFormat.nameAndTime.rawValue

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            if viewModel.isLoading {
                loadingView
            } else if let error = viewModel.errorMessage, viewModel.todayResponse == nil {
                errorView(message: error)
            } else if let today = viewModel.todayResponse {
                headerSection(today: today)
                Divider().padding(.vertical, 6)
                prayerListSection(today: today)
                Divider().padding(.vertical, 6)
                formatPickerSection
            }

            Divider().padding(.vertical, 6)
            footerSection
        }
        .padding(.horizontal, 12)
        .padding(.vertical, 10)
        .frame(width: 260)
        .onAppear {
            // Sync persisted format into view model
            if let format = DisplayFormat(rawValue: storedFormat) {
                viewModel.displayFormat = format
            }
        }
        .onChange(of: storedFormat) { _, newValue in
            if let format = DisplayFormat(rawValue: newValue) {
                viewModel.displayFormat = format
            }
        }
    }

    // MARK: - Header

    private func headerSection(today: TodayResponse) -> some View {
        VStack(alignment: .leading, spacing: 2) {
            Text("Prayer Times")
                .font(.headline)

            if let city = today.location.city, !city.isEmpty,
               let country = today.location.country, !country.isEmpty {
                Text("\(city), \(country)")
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }

            Text(today.date.gregorian)
                .font(.caption)
                .foregroundStyle(.secondary)

            if !today.date.hijri.isEmpty {
                Text(today.date.hijri)
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }
        }
    }

    // MARK: - Prayer List

    private func prayerListSection(today: TodayResponse) -> some View {
        VStack(alignment: .leading, spacing: 4) {
            ForEach(today.sortedPrayers) { prayer in
                prayerRow(prayer: prayer, today: today)
            }
        }
    }

    private func prayerRow(prayer: PrayerEntry, today: TodayResponse) -> some View {
        let isCurrent = today.current?.lowercased() == prayer.id
        let isNext = today.next?.prayer.lowercased() == prayer.id

        return HStack {
            HStack(spacing: 4) {
                if isNext {
                    Image(systemName: "arrow.right")
                        .font(.caption2)
                        .foregroundStyle(Color.accentColor)
                }
                Text(prayer.name)
                    .fontWeight(isNext ? .semibold : .regular)
            }

            Spacer()

            HStack(spacing: 6) {
                Text(prayer.time)
                    .monospacedDigit()

                if isNext, let remaining = today.next?.remaining {
                    Text("(\(remaining))")
                        .font(.caption)
                        .foregroundStyle(Color.accentColor)
                }
            }
        }
        .foregroundStyle(isCurrent ? .secondary : .primary)
        .padding(.vertical, 2)
        .padding(.horizontal, 4)
        .background {
            if isNext {
                RoundedRectangle(cornerRadius: 4)
                    .fill(Color.accentColor.opacity(0.1))
            }
        }
    }

    // MARK: - Format Picker

    private var formatPickerSection: some View {
        VStack(alignment: .leading, spacing: 4) {
            Text("Display Format")
                .font(.caption)
                .foregroundStyle(.secondary)

            Picker("Format", selection: $storedFormat) {
                ForEach(DisplayFormat.allCases) { format in
                    Text(format.displayName)
                        .tag(format.rawValue)
                }
            }
            .labelsHidden()
            .pickerStyle(.menu)
        }
    }

    // MARK: - Footer

    private var footerSection: some View {
        VStack(spacing: 6) {
            HStack {
                Toggle("Open at Login", isOn: Binding(
                    get: { SMAppService.mainApp.status == .enabled },
                    set: { newValue in
                        do {
                            if newValue {
                                try SMAppService.mainApp.register()
                            } else {
                                try SMAppService.mainApp.unregister()
                            }
                        } catch {
                            // Silently fail — user may not have granted permission
                        }
                    }
                ))
                .toggleStyle(.switch)
                .controlSize(.small)
            }

            HStack {
                Button("Refresh") {
                    Task { await viewModel.refreshAll() }
                }

                Spacer()

                Button("Quit") {
                    NSApplication.shared.terminate(nil)
                }
                .keyboardShortcut("q")
            }
        }
    }

    // MARK: - States

    private var loadingView: some View {
        VStack(spacing: 8) {
            ProgressView()
                .controlSize(.small)
            Text("Loading prayer times...")
                .font(.caption)
                .foregroundStyle(.secondary)
        }
        .frame(maxWidth: .infinity)
        .padding(.vertical, 20)
    }

    private func errorView(message: String) -> some View {
        VStack(spacing: 8) {
            Image(systemName: "exclamationmark.triangle")
                .font(.title2)
                .foregroundStyle(.yellow)

            Text("Unable to load prayer times")
                .font(.caption)
                .fontWeight(.medium)

            Text(message)
                .font(.caption2)
                .foregroundStyle(.secondary)
                .multilineTextAlignment(.center)

            Button("Retry") {
                Task { await viewModel.refreshAll() }
            }
            .controlSize(.small)
        }
        .frame(maxWidth: .infinity)
        .padding(.vertical, 12)
    }
}
