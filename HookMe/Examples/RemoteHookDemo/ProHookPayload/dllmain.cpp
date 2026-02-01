// Copyright 2026 HookMe Authors.
#include <windows.h>
#include <winternl.h>

#include "Prohook/core/hook_manager.h"
#include "Prohook/hooks/detours_hook.h"

// Define the NtCreateFile signature manually as it's not in standard headers
using pNtCreateFile = NTSTATUS(NTAPI*)(
    PHANDLE FileHandle,
    ACCESS_MASK DesiredAccess,
    POBJECT_ATTRIBUTES ObjectAttributes,
    PIO_STATUS_BLOCK IoStatusBlock,
    PLARGE_INTEGER AllocationSize,
    ULONG FileAttributes,
    ULONG ShareAccess,
    ULONG CreateDisposition,
    ULONG CreateOptions,
    PVOID EaBuffer,
    ULONG EaLength);

pNtCreateFile g_original_nt_create_file = nullptr;
// Our Proxy: Intercepts every file open/create request
NTSTATUS NTAPI MyProxyNtCreateFile(
    PHANDLE FileHandle,
    ACCESS_MASK DesiredAccess,
    POBJECT_ATTRIBUTES ObjectAttributes,
    PIO_STATUS_BLOCK IoStatusBlock,
    PLARGE_INTEGER AllocationSize,
    ULONG FileAttributes,
    ULONG ShareAccess,
    ULONG CreateDisposition,
    ULONG CreateOptions,
    PVOID EaBuffer,
    ULONG EaLength) {

        if (ObjectAttributes && ObjectAttributes->ObjectName) {
            // ObjectName is a UNICODE_STRING
            std::wstring path(ObjectAttributes->ObjectName->Buffer,
                ObjectAttributes->ObjectName->Length / sizeof(wchar_t));

            std::wstring log_msg = L"[ProHook] NtCreateFile: " + path;
            OutputDebugStringW(log_msg.c_str());
        }

    // Call the original function
    return g_original_nt_create_file(FileHandle, DesiredAccess, ObjectAttributes,
        IoStatusBlock, AllocationSize, FileAttributes,
        ShareAccess, CreateDisposition, CreateOptions,
        EaBuffer, EaLength);
}

void InitializeHooks() {
    auto& engine = prohook::core::HookManager::Instance();
    engine.SetHookProvider(std::make_unique<prohook::hooks::DetoursHook>());

    // Hooking ntdll directly to catch all file activity
    engine.AddHook("ntdll.dll!NtCreateFile",
        reinterpret_cast<void*>(MyProxyNtCreateFile),
        reinterpret_cast<void**>(&g_original_nt_create_file));

    if (engine.DeployAll()) {
        OutputDebugStringW(L"[ProHook] File Monitoring deployed.");
    }
}

BOOL APIENTRY DllMain(HMODULE hModule, DWORD reason, LPVOID reserved) {
    if (reason == DLL_PROCESS_ATTACH) {
        DisableThreadLibraryCalls(hModule);
        InitializeHooks();
    }
    else if (reason == DLL_PROCESS_DETACH) {
        prohook::core::HookManager::Instance().Teardown();
    }
    return TRUE;
}