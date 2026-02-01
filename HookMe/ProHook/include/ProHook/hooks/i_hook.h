// Copyright 2026 HookMe Authors.
#ifndef PROHOOK_INCLUDE_HOOKS_I_HOOK_H_
#define PROHOOK_INCLUDE_HOOKS_I_HOOK_H_

namespace prohook::hooks {

	// Interface for hooking techniques.
	class IHook {
	public:
		virtual ~IHook() = default;

		// Applies the redirection.
		// target: The address to be hooked.
		// detour: The address of the proxy function.
		// original: [Out] Pointer to the trampoline for calling the original code.
		virtual bool Install(void* target, void* detour, void** original) = 0;

		// Removes the hook and restores original bytes/pointers.
		virtual bool Uninstall() = 0;

		virtual bool IsInstalled() const = 0;
	};

}  // namespace prohook::hooks

#endif  // PROHOOK_INCLUDE_HOOKS_I_HOOK_H_