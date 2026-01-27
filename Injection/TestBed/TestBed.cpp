// TestBed.cpp : This file contains the 'main' function. Program execution begins and ends there.
//

//#include <iostream>
//#include <thread>
//#include <chrono>
//
//int main()
//{
//    std::cout << "Waiting For something odd ....\n";
//    for (int i{}; i < 100; i++) {
//        std::cout << ".\n";
//        std::this_thread::sleep_for(std::chrono::seconds(1));
//    }
//    std::cout << "GoodBye\n";
//    return 0;
// }

#include <windows.h>
#include <iostream>

int main() {
    HKEY hKey;
    const wchar_t* subKey = L"Software\\CyberCore";
    const wchar_t* valueName = L"user_id";

    // Open or create the registry key
    if (RegCreateKeyExW(
        HKEY_CURRENT_USER,
        subKey,
        0,
        nullptr,
        0,
        KEY_SET_VALUE,
        nullptr,
        &hKey,
        nullptr
    ) != ERROR_SUCCESS) {
        std::cerr << "Failed to open registry key\n";
        return 1;
    }

    // Empty wide string -> just L'\0'
    const wchar_t emptyString[] = L"";

    // Write REG_SZ (must include the null terminator)
    LONG result = RegSetValueExW(
        hKey,
        valueName,
        0,
        REG_SZ,
        reinterpret_cast<const BYTE*>(emptyString),
        sizeof(emptyString) // 2 bytes: UTF-16 null terminator
    );

    if (result != ERROR_SUCCESS) {
        std::cerr << "Failed to set registry value\n";
    }

    RegCloseKey(hKey);
    return 0;
}

