// Copyright 2026 HookMe Authors.
#include "ProHook/core/hook_manager.h"

#include <memory>
#include <string>
#include <vector>

#include <Windows.h>

#include "ProHook/hooks/i_hook.h"
#include "core/resolver.h"

namespace prohook::core {

    HookManager& HookManager::Instance() {
        static HookManager instance;
        return instance;
    }

    void HookManager::SetHookProvider(std::unique_ptr<hooks::IHook> provider) {
        provider_ = std::move(provider);
    }

    void HookManager::AddHook(const std::string& target_spec, void* proxy,
        void** original) {
        size_t delimiter = target_spec.find('!');
        if (delimiter == std::string::npos) return;

        HookContext ctx;
        ctx.module_name = target_spec.substr(0, delimiter);
        ctx.function_name = target_spec.substr(delimiter + 1);
        ctx.proxy_func = proxy;
        ctx.original_func = original;

        pending_hooks_.emplace_back(std::move(ctx));
    }

    bool HookManager::DeployAll() {
        if (!provider_) {
            // In a real EDR, you'd log an error here: "No hook provider assigned."
            return false;
        }

        bool all_successful = true;

        for (auto& hook : pending_hooks_) {
            // 1. Use Safe wrapper to get or load the module
            HMODULE h_module = GetModuleHandleA(hook.module_name.c_str());
            if (!h_module) {
                h_module = core::Resolver::SafeLoadLibrary(hook.module_name);
            }

            if (!h_module) {
                all_successful = false;
                continue;
            }

            // 2. Use Safe wrapper to find the export (handles Forwarders & Manual PE parsing)
            void* target_addr = core::Resolver::SafeGetProcAddress(h_module,
                hook.function_name);

            if (!target_addr) {
                all_successful = false;
                continue;
            }

            // 3. Apply the hook
            if (!provider_->Install(target_addr, hook.proxy_func, hook.original_func)) {
                all_successful = false;
            }
        }

        // Clear pending hooks after deployment to prevent double-hooking if 
        // DeployAll is called again.
        pending_hooks_.clear();

        return all_successful;
    }

    void HookManager::Teardown() {
        if (provider_) {
            provider_->Uninstall();
        }
    }

}  // namespace prohook