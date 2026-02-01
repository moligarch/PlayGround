// Copyright 2026 HookMe Authors.
#include "ProHook/core/hook_manager.h"

#include <memory>
#include <string>
#include <vector>

#include <Windows.h>

#include "ProHook/hooks/i_hook.h"
#include "ProHook/utils/resolver.h"

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
        if (!provider_) return false;

        bool all_successful = true;

        for (auto& hook : pending_hooks_) {
            // Convert string to wstring for the new Resolver interface
            std::wstring w_mod_name(hook.module_name.begin(), hook.module_name.end());

            // 1. Get module handle safely
            void* h_module = utils::Resolver::GetSafeModuleHandle(nullptr, w_mod_name);

            if (!h_module) {
                h_module = utils::Resolver::LocalLoadLibrary(hook.module_name);
            }

            if (!h_module) {
                all_successful = false;
                continue;
            }

            // 2. Get proc address safely (Local)
            void* target_addr = utils::Resolver::GetSafeProcAddress(nullptr, h_module, hook.function_name);

            if (!target_addr) {
                all_successful = false;
                continue;
            }

            // 3. Apply the hook
            if (!provider_->Install(target_addr, hook.proxy_func, hook.original_func)) {
                all_successful = false;
            }
        }

        pending_hooks_.clear();
        return all_successful;
    }

    void HookManager::Teardown() {
        if (provider_) {
            provider_->Uninstall();
        }
    }

}  // namespace prohook