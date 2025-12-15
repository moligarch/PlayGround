#include <iostream>
#include <string>
#include <vector>

#include <windows.h>

#pragma comment(lib, "version.lib")
#pragma comment(lib, "user32.lib")

// Function Pointers for the DLL
typedef void(__stdcall* FnStart)();
typedef void(__stdcall* FnStop)();

struct LanguageAndCodePage {
    WORD wLanguage;
    WORD wCodePage;
};

// ---------------------------------------------------------
// ROBUST VERSION UPDATER (WinAPI Structure Compliant)
// ---------------------------------------------------------
bool UpdateVersionResource(const std::wstring& filePath, const std::wstring& newVersion) {
    DWORD handle = 0;
    DWORD size = GetFileVersionInfoSizeW(filePath.c_str(), &handle);
    if (size == 0) {
        std::cerr << "[Patcher] Error: No version info found or file inaccessible." << std::endl;
        return false;
    }

    // 1. Load the Version Resource into a writable buffer
    std::vector<BYTE> versionData(size);
    if (!GetFileVersionInfoW(filePath.c_str(), handle, size, versionData.data())) {
        std::cerr << "[Patcher] Error: Failed to load version info." << std::endl;
        return false;
    }

    // 2. Update the Fixed Binary Version (VS_FIXEDFILEINFO)
    // This controls what GetFileVersionInfo returns programmatically.
    VS_FIXEDFILEINFO* pFileInfo = nullptr;
    UINT len = 0;
    if (VerQueryValueW(versionData.data(), L"\\", (void**)&pFileInfo, &len)) {
        // Parse "2.0.0.0" -> 2,0,0,0
        int major = 0, minor = 0, build = 0, revision = 0;
        swscanf_s(newVersion.c_str(), L"%d.%d.%d.%d", &major, &minor, &build, &revision);

        pFileInfo->dwFileVersionMS = MAKELONG(minor, major);
        pFileInfo->dwFileVersionLS = MAKELONG(revision, build);
        pFileInfo->dwProductVersionMS = MAKELONG(minor, major);
        pFileInfo->dwProductVersionLS = MAKELONG(revision, build);

        std::cout << "[Patcher] Updated FixedFileInfo binary version." << std::endl;
    }

    // 3. Update the String Table (What CFF Explorer/Properties shows)
    // We must find the correct Language/CodePage to build the query path.
    struct LanguageAndCodePage* pTransTable = nullptr;
    if (VerQueryValueW(versionData.data(), L"\\VarFileInfo\\Translation", (void**)&pTransTable, &len)) {
        // Calculate how many languages are available
        int count = len / sizeof(LanguageAndCodePage);

        for (int i = 0; i < count; i++) {
            WCHAR subBlock[50];
            // Format: \StringFileInfo\040904b0\FileVersion
            // 0409 = US English, 04b0 = Unicode
            swprintf_s(subBlock, L"\\StringFileInfo\\%04x%04x\\",
                pTransTable[i].wLanguage, pTransTable[i].wCodePage);

            std::wstring baseBlock = subBlock;
            std::wstring keys[] = { L"FileVersion", L"ProductVersion" };

            for (const auto& key : keys) {
                std::wstring query = baseBlock + key;
                LPWSTR pValue = nullptr;
                UINT valLen = 0;

                // VerQueryValue returns a pointer DIRECTLY into 'versionData' buffer
                if (VerQueryValueW(versionData.data(), query.c_str(), (void**)&pValue, &valLen)) {
                    // SAFETY CHECK: Ensure new string fits in the old space.
                    // versionData is a raw binary block. Changing lengths requires complex shifting 
                    // of the entire tree which VerQueryValue does not support.
                    // Since 1.0.0.0 and 2.0.0.0 are same length, this is safe.
                    if (newVersion.length() <= wcslen(pValue)) {
                        wcscpy_s(pValue, valLen, newVersion.c_str());
                        std::wcout << L"[Patcher] Updated " << key << L" for LangID "
                            << std::hex << pTransTable[i].wLanguage << std::endl;
                    }
                    else {
                        std::cerr << "[Patcher] Warning: New version string is longer than original. Skipping string update to prevent corruption." << std::endl;
                    }
                }
            }
        }
    }

    // 4. Commit changes back to file
    HANDLE hUpdate = BeginUpdateResourceW(filePath.c_str(), FALSE);
    if (!hUpdate) return false;

    // Note: We use MAKELANGID(LANG_NEUTRAL, SUBLANG_NEUTRAL) to try and replace the default.
    // Ideally, we should use pTransTable[0].wLanguage, but UpdateResource requires the Resource Language ID,
    // which might differ from the internal text language. 1033 (US) is standard for VS.
    // If this fails to replace, try 1033 explicitly.
    if (!UpdateResourceW(hUpdate, RT_VERSION, MAKEINTRESOURCE(1), 1033, versionData.data(), size)) {
        std::cerr << "[Patcher] Error: UpdateResource failed. " << GetLastError() << std::endl;
        return false;
    }

    if (!EndUpdateResourceW(hUpdate, FALSE)) {
        std::cerr << "[Patcher] Error: EndUpdateResource failed." << std::endl;
        return false;
    }

    return true;
}

// ---------------------------------------------------------
// UTILS
// ---------------------------------------------------------
bool SafeMoveFile(const std::wstring& src, const std::wstring& dst) {
    int attempts = 20;
    while (attempts--) {
        if (MoveFileExW(src.c_str(), dst.c_str(), MOVEFILE_REPLACE_EXISTING | MOVEFILE_COPY_ALLOWED)) {
            return true;
        }
        Sleep(100);
    }
    return false;
}

void RunDllCycle(const std::wstring& dllName) {
    std::wcout << L"[Host] Loading " << dllName << L"..." << std::endl;

    // Use LoadLibraryEx to ensure we don't lock the file if we only wanted data (not needed here, but good practice)
    HMODULE hDll = LoadLibraryW(dllName.c_str());
    if (!hDll) {
        std::cerr << "[Host] Failed to load DLL." << std::endl;
        return;
    }

    FnStart startFunc = (FnStart)GetProcAddress(hDll, "Start");
    FnStop stopFunc = (FnStop)GetProcAddress(hDll, "Stop");

    if (startFunc && stopFunc) {
        startFunc();
        Sleep(2000);
        stopFunc();
    }

    FreeLibrary(hDll);
}

int main() {
    std::wstring originalDll = L"Runner.dll";
    std::wstring backuplDll = L"Runner.dll.bak";
    std::wstring copyDll = L"Runner.dll.new";

    std::cout << "--- PHASE 1: Initial Run (v1) ---" << std::endl;
    RunDllCycle(originalDll);

    std::cout << "\n--- PHASE 2: Updating ---" << std::endl;

    // 1. Copy runner to modify it later and also back it up
    if (!CopyFileW(originalDll.c_str(), copyDll.c_str(), FALSE)) {
        std::cerr << "Failed to copy file." << std::endl;
        return 1;
    }

    if (!CopyFileW(originalDll.c_str(), backuplDll.c_str(), FALSE)) {
        std::cerr << "Failed to backup file." << std::endl;
        return 1;
    }

    // 2. WinAPI Update Version
    if (!UpdateVersionResource(copyDll, L"2.0.0.0")) {
        return 1;
    }

    // 3. Swap: Move copy over original
    if (SafeMoveFile(copyDll, originalDll)) {
        std::cout << "[Host] File swapped successfully." << std::endl;
    }
    else {
        std::cerr << "[Host] Failed to swap file. Error: " << GetLastError() << std::endl;
        return 1;
    }

    std::cout << "\n--- PHASE 3: Reload Run (v2) ---" << std::endl;
    RunDllCycle(originalDll);

    std::cout << "\nDone." << std::endl;
    getchar();
    return 0;
}