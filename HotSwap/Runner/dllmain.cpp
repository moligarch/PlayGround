// Runner.cpp
#include "pch.h"

#pragma comment(lib, "Version.lib")
// Global state
std::atomic<bool> g_Running = false;
std::thread g_WorkerThread;

// Helper to get version string from this DLL's resources
std::wstring GetMyVersion() {
    WCHAR fileName[MAX_PATH];
    HMODULE hMod = GetModuleHandleW(L"runner.dll");
    if (!hMod) return L"Unknown";

    GetModuleFileNameW(hMod, fileName, MAX_PATH);

    DWORD handle = 0;
    DWORD size = GetFileVersionInfoSizeW(fileName, &handle);
    if (size == 0) return L"NoVersion";

    std::vector<BYTE> versionData(size);
    if (!GetFileVersionInfoW(fileName, handle, size, versionData.data())) return L"InfoFailed";

    VS_FIXEDFILEINFO* fileInfo = nullptr;
    UINT len = 0;
    if (VerQueryValueW(versionData.data(), L"\\", (void**)&fileInfo, &len)) {
        return std::to_wstring(HIWORD(fileInfo->dwFileVersionMS)) + L"." +
            std::to_wstring(LOWORD(fileInfo->dwFileVersionMS)) + L"." +
            std::to_wstring(HIWORD(fileInfo->dwFileVersionLS)) + L"." +
            std::to_wstring(LOWORD(fileInfo->dwFileVersionLS));
    }
    return L"QueryFailed";
}

void WorkerTask() {
    std::wstring ver = GetMyVersion();
    while (g_Running) {
        std::wcout << L"[Runner.dll] Version: " << ver << L" | Working..." << std::endl;
        std::this_thread::sleep_for(std::chrono::milliseconds(1000));
    }
}

extern "C" __declspec(dllexport) void __stdcall Start() {
    if (g_Running) return;
    g_Running = true;
    g_WorkerThread = std::thread(WorkerTask);
    std::cout << "[Runner.dll] Started." << std::endl;
}

extern "C" __declspec(dllexport) void __stdcall Stop() {
    if (!g_Running) return;
    g_Running = false;
    if (g_WorkerThread.joinable()) {
        g_WorkerThread.join();
    }
    std::cout << "[Runner.dll] Stopped." << std::endl;
}