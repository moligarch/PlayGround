// Copyright 2026 HookMe Authors.
#ifndef PROHOOK_INCLUDE_HOOKS_DETOURS_HOOK_H_
#define PROHOOK_INCLUDE_HOOKS_DETOURS_HOOK_H_
#include "ProHook/hooks/i_hook.h"

#include <vector>

namespace prohook::hooks {

    class DetoursHook : public IHook {
    public:
        DetoursHook() = default;
        ~DetoursHook() override { Uninstall(); }

        // Disallow copy/assign
        DetoursHook(const DetoursHook&) = delete;
        DetoursHook& operator=(const DetoursHook&) = delete;

        bool Install(void* target, void* detour, void** original) override;
        bool Uninstall() override;
        bool IsInstalled() const override { return is_installed_; }

    private:
        bool is_installed_ = false;

        // We need to store original/target pairs to perform a bulk Uninstall.
        struct HookEntry {
            void* target;
            void** original;
            void* proxy;
        };
        std::vector<HookEntry> active_hooks_;
    };

}  // namespace prohook::hooks

#endif  // PROHOOK_INCLUDE_HOOKS_DETOURS_HOOK_H_