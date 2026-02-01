// Copyright 2026 HookMe Authors.
#include "ProHook/hooks/detours_hook.h"

#include <windows.h>

#include <detours.h>

namespace prohook::hooks {

    bool DetoursHook::Install(void* target, void* detour, void** original) {
        if (!target || !detour || !original) return false;

        // Detours requires that *original initially points to the target.
        *original = target;

        DetourTransactionBegin();
        DetourUpdateThread(GetCurrentThread());

        LONG error = DetourAttach(original, detour);
        if (error != NO_ERROR) {
            DetourTransactionAbort();
            return false;
        }

        if (DetourTransactionCommit() != NO_ERROR) {
            DetourTransactionAbort();
            return false;
        }

        active_hooks_.emplace_back(HookEntry{ target, original, detour });
        is_installed_ = true;
        return true;
    }

    bool DetoursHook::Uninstall() {
        if (!is_installed_ || active_hooks_.empty()) return true;

        LONG error = DetourTransactionBegin();
        if (error != NO_ERROR) return false;

        error = DetourUpdateThread(GetCurrentThread());
        if (error != NO_ERROR) {
            DetourTransactionAbort();
            return false;
        }

        for (auto& hook : active_hooks_) {
            // Pass the proxy_func as the second argument!
            error = DetourDetach(hook.original, hook.proxy);
            if (error != NO_ERROR) {
                DetourTransactionAbort();
                return false;
            }
        }

        error = DetourTransactionCommit();
        if (error != NO_ERROR) {
            DetourTransactionAbort();
            return false;
        }

        active_hooks_.clear();
        is_installed_ = false;
        return true;
    }

}  // namespace prohook::hooks