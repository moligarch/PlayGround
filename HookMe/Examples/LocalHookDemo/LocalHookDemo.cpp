// Copyright 2026 HookMe Authors.
#include <windows.h>
#include <iostream>
#include <memory>

#include "ProHook/core/hook_manager.h"
#include "ProHook/hooks/detours_hook.h" // We include the concrete provider here

// 1. Define the function signature and the 'Original' pointer.
using pMessageBoxW = int (WINAPI*)(HWND, LPCWSTR, LPCWSTR, UINT);
pMessageBoxW g_original_message_box = nullptr;

// 2. Define our Proxy function.
int WINAPI MyProxyMessageBoxW(HWND hWnd, LPCWSTR lpText, LPCWSTR lpCaption, UINT uType) {
    std::wcout << L"[ProHook] Intercepted MessageBoxW!" << std::endl;
    std::wcout << L"  Original Text: " << lpText << std::endl;

    // We can modify the arguments before calling the original.
    std::wstring new_text = L"Hooked: " + std::wstring(lpText);

    // Call the original function via the trampoline.
    return g_original_message_box(hWnd, new_text.c_str(), L"HookMe Security", uType);
}

int main() {
    auto& engine = prohook::core::HookManager::Instance();

    // 3. Setup the Engine with the Detours provider.
    engine.SetHookProvider(std::make_unique<prohook::hooks::DetoursHook>());

    // 4. Register the hook using our "Friendly" string format.
    // Note: We cast g_original_message_box to void** so the engine can write to it.
    engine.AddHook("user32.dll!MessageBoxW",
        reinterpret_cast<void*>(MyProxyMessageBoxW),
        reinterpret_cast<void**>(&g_original_message_box));

    std::cout << "Deploying hooks..." << std::endl;
    if (engine.DeployAll()) {
        std::cout << "Hooks deployed successfully!\n" << std::endl;
    }
    else {
        std::cerr << "Failed to deploy hooks." << std::endl;
        return 1;
    }

    // 5. Trigger the hook.
    MessageBoxW(NULL, L"This is a test message.", L"Original Title", MB_OK);

    // 6. Cleanup.
    std::cout << "Cleaning up..." << std::endl;
    engine.Teardown();

    // This call should be back to normal now.
    MessageBoxW(NULL, L"The hook should be gone now.", L"Post-Teardown", MB_OK);

    return 0;
}